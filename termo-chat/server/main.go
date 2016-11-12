package main

import (
	"errors"
	"fmt"
	"github.com/austinov/go-recipes/termo-chat/common/proto"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ugorji/go/codec"
)

const (
	version = 1
)

type peer struct {
	name string
	id   string
	conn net.Conn
}

type room struct {
	id    string
	peers map[string]peer // key is peer id
}

var (
	netw        = "tcp"   // TODO from flags
	laddr       = ":8822" // TODO from flags
	closed      = false
	rooms       = make(map[string]room)     // key is room id
	conns       = make(map[net.Conn]string) // value is room id
	readTimeout = 20 * time.Second
	listener    net.Listener
)

func main() {
	var err error
	listener, err = net.Listen(netw, laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closeListener()

	// handle interruption
	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	defer close(interrupt)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-interrupt
		closeListener()
		close(done)
		signal.Stop(interrupt)
	}()

	for {
		select {
		case <-done:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Accept error:", err)
				continue
			}
			go receive(conn, done)
		}
	}
}

func receive(conn net.Conn, done <-chan struct{}) {
	defer closeConn(conn)
	for {
		select {
		case <-done:
			return
		default:
			var rerr error
			b := make([]byte, 0)
			for {
				if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
					return
				}
				temp := make([]byte, 255)
				n, err := conn.Read(temp)
				if err != nil {
					if neterr, ok := err.(net.Error); err != io.EOF && ok && !neterr.Timeout() {
						rerr = err
					}
					break
				}
				if n > 0 {
					b = append(b, temp[:n]...)
				}
				if n < len(temp) {
					break
				}
			}
			if len(b) == 0 || rerr != nil {
				if rerr != nil {
					log.Println("Read error:", rerr)
				}
				<-time.After(time.Second)
			} else {
				if err := func() error {
					p, err := proto.Decode(b)
					if err != nil {
						return err
					}
					if err := checkPacket(p); err != nil {
						return err
					}
					if err := handleAction(conn, p); err != nil {
						return err
					}
					return nil
				}(); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func handleAction(conn net.Conn, p proto.Packet) error {
	switch p.Action {
	case proto.BookRoom:
		return bookRoom(conn, p)
	case proto.JoinRoom:
		return joinRoom(conn, p)
	case proto.SendMsg:
		return sendMsg(conn, p)
	case proto.LeaveRoom:
		return leaveRoom(conn, p)
	default:
		return sendError(conn, p.Version, p.Action, errors.New("Unexpected protocol action"))
	}
}

var roomId uint32 = 0

func peersByRoomArray(r room) []string {
	peers := make([]string, len(r.peers))
	i := 0
	for _, p := range r.peers {
		peers[i] = p.name
		i++
	}
	return peers
}

func checkPacket(p proto.Packet) error {
	if version != p.Version {
		return errors.New("Unsupported protocol version")
	}
	if p.Data.PeerId == "" {
		return errors.New("MAC not assigned")
	}
	return nil
}

func checkPacketData(p proto.Packet) error {
	if p.Data.RoomId == "" {
		return errors.New("Unknown room number")
	}
	if p.Data.PeerId == "" {
		return errors.New("MAC not assigned")
	}
	if _, ok := rooms[p.Data.RoomId]; !ok {
		return errors.New("Room not found")
	}
	return nil
}

func bookRoom(conn net.Conn, p proto.Packet) error {
	// TODO generate room id
	id := fmt.Sprintf("%d", atomic.AddUint32(&roomId, 1))
	peers := make(map[string]peer)
	peers[p.Data.PeerId] = peer{
		name: p.Data.PeerName,
		id:   p.Data.PeerId,
		conn: conn,
	}
	room := room{
		id:    id,
		peers: peers,
	}
	rooms[id] = room

	// join connection with room
	conns[conn] = id

	packet := proto.NewPacketData(
		version,
		proto.BookRoom,
		proto.DataPacket{
			RoomId: id,
			Peers:  peersByRoomArray(room),
		})
	log.Printf("booked room: %#v\n", packet)
	return send(conn, packet)
}

func joinRoom(conn net.Conn, p proto.Packet) error {
	if err := checkPacketData(p); err != nil {
		return sendError(conn, p.Version, p.Action, err)
	}
	// find room by id
	room := rooms[p.Data.RoomId]
	// find peer in the room by MAC
	if pr, ok := room.peers[p.Data.PeerId]; !ok {
		// if peer not found then add
		room.peers[p.Data.PeerId] = peer{
			name: p.Data.PeerName,
			id:   p.Data.PeerId,
			conn: conn,
		}
	} else if pr.name != p.Data.PeerName {
		pr.name = p.Data.PeerName
	}
	peers := peersByRoomArray(room)

	p = proto.NewPacketData(
		version,
		proto.JoinRoom,
		proto.DataPacket{
			RoomId: p.Data.RoomId,
			Peers:  peers,
		})

	// join connection with room
	conns[conn] = room.id

	// send response only to initiator
	if err := send(conn, p); err != nil {
		return err
	}
	// send update info to other peers
	pd := proto.NewPacketData(
		version,
		proto.UpdateRoom,
		proto.DataPacket{
			RoomId: p.Data.RoomId,
			Peers:  peers,
		})
	log.Printf("joined room: %#v\n", pd)
	return sendOthers(pd, room, p.Data.PeerId)
}

func sendMsg(conn net.Conn, p proto.Packet) error {
	if err := checkPacketData(p); err != nil {
		return sendError(conn, p.Version, p.Action, err)
	}
	// find room by id
	room := rooms[p.Data.RoomId]
	pd := proto.NewPacketData(
		version,
		proto.SendMsg,
		proto.DataPacket{
			RoomId:  p.Data.RoomId,
			Peers:   peersByRoomArray(room),
			Message: p.Data.Message,
			Sender:  room.peers[p.Data.PeerId].name,
		})
	log.Printf("Server send other peer=%v, packet=%#v\n", p.Data.PeerId, pd)
	return sendOthers(pd, room, p.Data.PeerId)
}

func leaveRoom(conn net.Conn, p proto.Packet) error {
	if err := checkPacketData(p); err != nil {
		return sendError(conn, p.Version, p.Action, err)
	}
	closeConn(conn)
	return nil
}

func sendOthers(packet proto.Packet, room room, peerId string) error {
	for id, p := range room.peers {
		if id != peerId {
			log.Printf("sendOther peer=%v, to=%#v\n", peerId, packet)
			send(p.conn, packet)
		}
	}
	return nil
}

func sendError(conn net.Conn, version, action byte, err error) error {
	p := proto.NewPacketError(version, action, err)
	return send(conn, p)
}

func send(conn net.Conn, p proto.Packet) error {
	var (
		b  []byte
		mh codec.MsgpackHandle
	)
	enc := codec.NewEncoderBytes(&b, &mh)
	if err := enc.Encode(p); err != nil {
		return err
	}
	_, err := conn.Write(b)
	return err
}

func closeListener() {
	if !closed {
		listener.Close()
		closed = true
	}
}

func closeConn(conn net.Conn) {
	if roomId, ok := conns[conn]; ok {
		room := rooms[roomId]
		for id, peer := range room.peers {
			if peer.conn == conn {
				deletePeer(room, peer)
				delete(room.peers, id)
				break
			}
		}
		delete(conns, conn)
		conn.Close()
	}
}

func deletePeer(r room, p peer) {
	// remove peer from room
	delete(r.peers, p.id)

	// send update all but initiator
	peers := peersByRoomArray(r)

	pd := proto.NewPacketData(
		version,
		proto.UpdateRoom,
		proto.DataPacket{
			RoomId: r.id,
			Peers:  peers,
		})
	sendOthers(pd, r, p.id)
}
