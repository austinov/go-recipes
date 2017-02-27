package net

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/austinov/go-recipes/backoff"
	"github.com/austinov/go-recipes/termo-chat/client/view"
	"github.com/austinov/go-recipes/termo-chat/common/proto"

	"github.com/ugorji/go/codec"
)

const (
	version           byte = 1
	attemptsToConnect byte = 5
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
	h.view = v
	vch := h.view.Show()

	if err := h.connect(); err != nil {
		h.view.ViewMessage(view.InfoMessage, "", "Unable to connect to the server. Please, try later...")
		<-time.After(3 * time.Second)
		h.view.Quit()
		return err
	}
	defer h.disconnect()

	dch := make(chan []byte) // data channel
	ech := make(chan error)  // error channel

	go func() {
		h.handleConnection(dch, ech)
	}()
	go func() {
		h.processData(dch, ech)
	}()

	if room == "" {
		if err := h.bookRoom(); err != nil {
			return errors.New(fmt.Sprintf("Error booking room: %s", err))
		}
	} else {
		if err := h.joinRoom(room); err != nil {
			return errors.New(fmt.Sprintf("Error joining room: %s", err))
		}
	}

	<-vch

	return nil
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

func (h *NetHandler) bookRoom() error {
	p := proto.NewPacketData(
		version,
		proto.BookRoom,
		proto.DataPacket{
			RoomId:   h.roomId,
			PeerName: h.peerName,
			PeerId:   h.peerId,
		})
	return h.send(p)
}

func (h *NetHandler) joinRoom(roomId string) error {
	h.roomId = roomId
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

var (
	tryConnectMsg = ""
)

func (h *NetHandler) connect() error {
	// first disconnect
	h.disconnect()
	// show message
	if tryConnectMsg == "" {
		tryConnectMsg = "Try to connect to the server..."
		h.view.ViewMessage(view.InfoMessage, "", tryConnectMsg)
	} else {
		h.view.ViewMessage(view.TailMessage, "", "...")
	}
	// try connect
	var err error
	eb := backoff.NewExpBackoff()
	for {
		if h.conn, err = net.Dial(h.network, h.address); err != nil {
			if eb.Attempts() == uint64(attemptsToConnect) {
				break
			}
			<-eb.Delay()
		} else {
			tryConnectMsg = ""
			h.view.ViewMessage(view.InfoMessage, "", "You are connected to the server!")
			return nil
		}
	}
	return err
}

func (h *NetHandler) disconnect() {
	if h.conn == nil {
		return
	}
	h.conn.Close()
	h.conn = nil
}

func (h *NetHandler) checkConnection(enforce bool, ech chan<- error) bool {
	if h.conn == nil || enforce {
		if err := h.connect(); err != nil {
			return false
		}
		// re-book room with existence id
		if err := h.bookRoom(); err != nil {
			ech <- errors.New(fmt.Sprintf("Error re-booking room: %s", err))
		}
	}
	return true
}

// handleConnection reads data from connection and then push the data to dch channel
func (h *NetHandler) handleConnection(dch chan<- []byte, ech chan<- error) {
	for {
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
				} else if err == io.EOF {
					h.checkConnection(true, ech)
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

// processData processes data and errors
func (h *NetHandler) processData(dch <-chan []byte, ech <-chan error) {
	for {
		select {
		case b := <-dch:
			msg, err := proto.Decode(b)
			if err != nil {
				h.view.ViewMessage(view.ErrorMessage, "", err.Error())
				continue
			}
			if version != msg.Version {
				h.view.ViewMessage(view.ErrorMessage, "", fmt.Sprintf("Unsupported protocol version: %d", msg.Version))
				continue
			}
			if msg.Err != "" {
				h.view.ViewMessage(view.ErrorMessage, "", msg.Err)
				continue
			}
			switch msg.Action {
			case proto.BookRoom:
				if msg.Data.RoomId == "" {
					h.view.ViewMessage(view.ErrorMessage, "", "Incorrect data format: room not found")
					continue
				}
				h.roomId = msg.Data.RoomId
				h.view.ViewMessage(view.InfoMessage, "",
					fmt.Sprintf("You are booked the room number %s. Please send this number to your peers to join the room.", h.roomId))
				h.view.ViewRoom(h.roomId)
				h.view.UpdatePeers(msg.Data.Peers)
			case proto.JoinRoom:
				h.view.ViewMessage(view.InfoMessage, "",
					fmt.Sprintf("You are joined the room number %s.", h.roomId))
				h.view.ViewRoom(h.roomId)
				h.view.UpdatePeers(msg.Data.Peers)
			case proto.SendMsg:
				h.view.ViewMessage(view.ChatMessage, msg.Data.Sender, msg.Data.Message)
			case proto.UpdateRoom:
				h.view.UpdatePeers(msg.Data.Peers)
			default:
				h.view.ViewMessage(view.ErrorMessage, "", fmt.Sprintf("Unexpected protocol action: %d", msg.Action))
			}
		case err := <-ech:
			h.view.ViewMessage(view.ErrorMessage, "", err.Error())
		}
	}
}

func (h *NetHandler) send(p proto.Packet) error {
	if h.conn == nil {
		return errors.New("Missing connection with the server.\n" + tryConnectMsg)
	}
	var (
		b  []byte
		mh codec.MsgpackHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	if err := enc.Encode(p); err != nil {
		return err
	}
	_, err := h.conn.Write(b)
	if err != nil {
		return errors.New("Unable to send message to server: " + err.Error())
	}
	return nil
}

func getPeerId() string {
	// generate 32 bits timestamp
	unix32bits := uint32(time.Now().UTC().Unix())
	buff := make([]byte, 12)
	numRead, err := rand.Read(buff)
	if numRead != len(buff) || err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x-%x",
		unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:12])
}
