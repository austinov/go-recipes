package net

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/austinov/go-recipes/backoff"
	"github.com/austinov/go-recipes/termo-chat/client/view"
	"github.com/austinov/go-recipes/termo-chat/common/proto"

	"github.com/ugorji/go/codec"
)

const (
	version           byte = 1
	attemptsToConnect byte = 2
)

type NetHandler struct {
	network  string
	address  string
	conn     net.Conn
	roomId   string
	peerId   string
	peerName string
	view     view.View
}

func New(network, address, peerName string) *NetHandler {
	return &NetHandler{
		network:  network,
		address:  address,
		peerId:   getPeerId(),
		peerName: peerName,
	}
}

func (h *NetHandler) Init(v view.View, room string) error {
	if err := h.connect(); err != nil {
		return err
	}
	defer h.disconnect()

	h.view = v
	vch := h.view.Show()

	dch := make(chan []byte)    // data channel
	ech := make(chan error)     // error channel
	done := make(chan struct{}) // done channel

	go func() {
		h.handleConnection(dch, ech, done)
	}()
	go func() {
		h.processData(dch, ech, done)
	}()

	if room == "" {
		if err := h.bookRoom(h.peerName); err != nil {
			return errors.New(fmt.Sprintf("Error booking room: %s", err))
		}
	} else {
		if err := h.joinRoom(room, h.peerName); err != nil {
			return errors.New(fmt.Sprintf("Error joining room: %s", err))
		}
	}

	<-vch

	return nil
}

func (h *NetHandler) bookRoom(peerName string) error {
	h.peerName = peerName
	p := proto.NewPacketData(
		version,
		proto.BookRoom,
		proto.DataPacket{
			PeerName: h.peerName,
			PeerId:   h.peerId,
		})
	return h.send(p)
}

func (h *NetHandler) joinRoom(roomId, peerName string) error {
	h.roomId = roomId
	h.peerName = peerName
	p := proto.NewPacketData(
		version,
		proto.JoinRoom,
		proto.DataPacket{
			RoomId:   h.roomId,
			PeerId:   h.peerId,
			PeerName: h.peerName,
		})
	return h.send(p)
}

func (h *NetHandler) SendMessage(message string) error {
	p := proto.NewPacketData(
		version,
		proto.SendMsg,
		proto.DataPacket{
			RoomId:  h.roomId,
			PeerId:  h.peerId,
			Message: message,
		})
	return h.send(p)
}

func (h *NetHandler) connect() error {
	h.closeConn()
	var err error
	eb := backoff.NewExpBackoff()
	for {
		//log.Println("Try connect to server...")
		if h.conn, err = net.Dial(h.network, h.address); err != nil {
			if eb.Attempts() == uint64(attemptsToConnect) {
				break
			}
			<-eb.Delay()
		} else {
			return nil
		}
	}
	return err
}

func (h *NetHandler) disconnect() {
	h.closeConn()
}

func (h *NetHandler) closeConn() {
	if h.conn == nil {
		return
	}
	defer h.conn.Close()
	if h.roomId != "" {
		p := proto.NewPacketData(
			version,
			proto.LeaveRoom,
			proto.DataPacket{
				RoomId: h.roomId,
				PeerId: h.peerId,
			})
		if err := h.send(p); err != nil {
			log.Printf("Error leaving room: %v", err)
		}
	}
}

func (h *NetHandler) checkConnection(force bool, ech chan<- error) bool {
	if h.conn == nil || force {
		if err := h.connect(); err != nil {
			ech <- err
			return false
		}
	}
	return true
}

// handleConnection reads data from connection and then push the data to dch channel
func (h *NetHandler) handleConnection(dch chan<- []byte, ech chan<- error, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			data := make([]byte, 0)
			for {
				if ok := h.checkConnection(false, ech); !ok {
					break
				}
				if err := h.conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
					ech <- err
					if ok := h.checkConnection(true, ech); !ok {
						break
					}
				}
				buf := make([]byte, 255)
				n, err := h.conn.Read(buf)
				if err != nil {
					if neterr, ok := err.(net.Error); err != io.EOF && ok && !neterr.Timeout() {
						ech <- err
					}
					break
				} else {
					if n > 0 {
						data = append(data, buf[:n]...)
					}
					if n < len(buf) {
						break
					}
				}
			}
			if len(data) > 0 {
				dch <- data
			} else {
				// Take a small pause in the case of empty data or some error
				<-time.After(time.Second)
			}
		}
	}
}

// processData processes data and errors
func (h *NetHandler) processData(dch <-chan []byte, ech <-chan error, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case b := <-dch:
			msg, err := decode(b)
			if err != nil {
				log.Println(err)
			}
			if version != msg.Version {
				log.Printf("Unsupported protocol version: %d", msg.Version)
			}
			if msg.Err != "" {
				log.Println(msg.Err)
			}
			switch msg.Action {
			case proto.BookRoom:
				if msg.Data.RoomId == "" {
					log.Println("Incorrect data format: room not found")
					continue
				}
				h.roomId = msg.Data.RoomId
				h.view.UpdatePeers(msg.Data.Peers)
			case proto.JoinRoom:
				h.view.UpdatePeers(msg.Data.Peers)
			case proto.SendMsg:
				h.view.ReceiveMessage(view.ChatMessage, msg.Data.Sender, msg.Data.Message)
			case proto.UpdateRoom:
				h.view.UpdatePeers(msg.Data.Peers)
			default:
				log.Printf("Unexpected protocol action: %d", msg.Action)
			}
		case err := <-ech:
			h.view.ReceiveMessage(view.ErrorMessage, "", err.Error())
		}
	}
}

func (h *NetHandler) send(p proto.Packet) error {
	var (
		b  []byte
		mh codec.MsgpackHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	if err := enc.Encode(p); err != nil {
		return err
	}
	_, err := h.conn.Write(b)
	return err
}

func decode(b []byte) (proto.Packet, error) {
	var (
		p  proto.Packet
		mh codec.MsgpackHandle
	)
	dec := codec.NewDecoderBytes(b, &mh)
	err := dec.Decode(&p)
	return p, err
}

//TODO extract
func getPeerId() string {
	// generate 32 bits timestamp
	unix32bits := uint32(time.Now().UTC().Unix())
	buff := make([]byte, 12)
	numRead, err := rand.Read(buff)
	if numRead != len(buff) || err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x-%x\n", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:12])
}
