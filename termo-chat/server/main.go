package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/austinov/go-recipes/termo-chat/common/proto"

	"github.com/ugorji/go/codec"
)

const (
	version    = 1
	numSenders = 2
)

type (
	peer struct {
		name string
		conn net.Conn
	}

	room struct {
		peers map[string]peer // key is peer id
	}

	tube struct {
		c net.Conn
		p proto.Packet
	}
)

var (
	netw        = "tcp"
	laddr       = ""
	closed      = false
	readTimeout = 20 * time.Second
	tch         = make(chan tube) // channel to send messages
	listener    net.Listener

	mur   sync.Mutex
	rooms = make(map[string]room) // key is room id

	muc   sync.Mutex
	conns = make(map[net.Conn]string) // value is room id
)

func main() {
	flag.StringVar(&laddr, "addr", ":8822",
		"The syntax of addr is \"host:port\", like \"127.0.0.1:8822\". "+
			"If host is omitted, as in \":8822\", Listen listens on all available interfaces.")
	flag.Parse()

	var err error
	if listener, err = net.Listen(netw, laddr); err != nil {
		log.Fatal(err)
	}
	defer closeListener()
	defer close(tch)

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

	// start message senders
	for i := 0; i < numSenders; i++ {
		go send(done)
	}

	for {
		select {
		case <-done:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Println("accept error:", err)
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
					} else if err == io.EOF {
						return
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
					log.Println("read error:", rerr)
				}
				<-time.After(time.Second)
			} else {
				if err := func() error {
					if p, err := proto.Decode(b); err != nil {
						return err
					} else if err := checkPacket(p); err != nil {
						return err
					} else {
						handleAction(conn, p)
					}
					return nil
				}(); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func handleAction(conn net.Conn, p proto.Packet) {
	switch p.Action {
	case proto.BookRoom:
		bookRoom(conn, p)
	case proto.JoinRoom:
		joinRoom(conn, p)
	case proto.SendMsg:
		sendMsg(conn, p)
	case proto.LeaveRoom:
		leaveRoom(conn, p)
	default:
		sendError(conn, p.Version, p.Action, errors.New("protocol action unexpected"))
	}
}

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
		return errors.New("protocol version unsupported")
	}
	if p.Data.PeerId == "" {
		return errors.New("peer id not assigned")
	}
	return nil
}

func checkPacketData(p proto.Packet) error {
	if p.Data.RoomId == "" {
		return errors.New("room number not assigned")
	}
	if p.Data.PeerId == "" {
		return errors.New("peer id not assigned")
	}
	if _, ok := getRoom(p.Data.RoomId); !ok {
		return errors.New("room " + p.Data.RoomId + " not found")
	}
	return nil
}

func bookRoom(conn net.Conn, p proto.Packet) {
	var roomId string
	rebooked := false
	if p.Data.RoomId != "" {
		p.Action = proto.UpdateRoom
		// re-book room, check existing room
		if _, ok := getRoom(p.Data.RoomId); !ok {
			// book room with came id
			rebooked = true
			roomId = p.Data.RoomId
		} else {
			// room exists join the peer
			joinRoom(conn, p)
			return
		}
	} else {
		roomId = generateRoomId()
	}
	peers := make(map[string]peer)
	peers[p.Data.PeerId] = peer{
		name: p.Data.PeerName,
		conn: conn,
	}
	room := room{
		peers: peers,
	}
	setRoom(roomId, room)

	// join connection with room
	setConnRoom(conn, roomId)

	roomPeers := peersByRoomArray(room)

	packet := proto.NewPacketData(
		version,
		p.Action,
		proto.DataPacket{
			RoomId: roomId,
			Peers:  roomPeers,
		})
	tch <- tube{conn, packet}
	if rebooked {
		log.Printf("peer [%s] re-booked the room %s - %v\n", p.Data.PeerId, roomId, roomPeers)
	} else {
		log.Printf("peer [%s] booked the room %s - %v\n", p.Data.PeerId, roomId, roomPeers)
	}
}

func joinRoom(conn net.Conn, p proto.Packet) {
	if err := checkPacketData(p); err != nil {
		sendError(conn, p.Version, p.Action, err)
		return
	}
	// find room by id
	room, ok := getRoom(p.Data.RoomId)
	if !ok {
		sendError(conn, p.Version, p.Action, errors.New("room "+p.Data.RoomId+" not found"))
		return
	}
	// find peer in the room by MAC
	if pr, ok := room.peers[p.Data.PeerId]; !ok {
		// if peer not found then add
		room.peers[p.Data.PeerId] = peer{
			name: p.Data.PeerName,
			conn: conn,
		}
	} else if pr.name != p.Data.PeerName {
		pr.name = p.Data.PeerName
	}
	// join connection with room
	setConnRoom(conn, p.Data.RoomId)

	peers := peersByRoomArray(room)

	// send response only to initiator
	packet := proto.NewPacketData(
		version,
		p.Action,
		proto.DataPacket{
			RoomId: p.Data.RoomId,
			Peers:  peers,
		})
	tch <- tube{conn, packet}

	// send update info to other peers
	pd := proto.NewPacketData(
		version,
		proto.UpdateRoom,
		proto.DataPacket{
			RoomId: p.Data.RoomId,
			Peers:  peers,
		})
	sendOthers(pd, room, p.Data.PeerId)
	log.Printf("peer [%s] joined the room %s - %v\n", p.Data.PeerId, p.Data.RoomId, peers)
}

func sendMsg(conn net.Conn, p proto.Packet) {
	if err := checkPacketData(p); err != nil {
		sendError(conn, p.Version, p.Action, err)
		return
	}
	// find room by id
	room, ok := getRoom(p.Data.RoomId)
	if !ok {
		sendError(conn, p.Version, p.Action, errors.New("room "+p.Data.RoomId+" not found"))
		return
	}
	pd := proto.NewPacketData(
		version,
		proto.SendMsg,
		proto.DataPacket{
			RoomId:  p.Data.RoomId,
			Peers:   peersByRoomArray(room),
			Message: p.Data.Message,
			Sender:  room.peers[p.Data.PeerId].name,
		})
	sendOthers(pd, room, p.Data.PeerId)
}

func leaveRoom(conn net.Conn, p proto.Packet) {
	if err := checkPacketData(p); err != nil {
		sendError(conn, p.Version, p.Action, err)
		return
	}
	closeConn(conn)
}

func sendOthers(packet proto.Packet, room room, peerId string) {
	for id, p := range room.peers {
		if id != peerId {
			tch <- tube{p.conn, packet}
		}
	}
}

func sendError(conn net.Conn, version, action byte, err error) {
	tch <- tube{conn, proto.NewPacketError(version, action, err)}
}

func send(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case t := <-tch:
			var (
				b  []byte
				mh codec.MsgpackHandle
			)
			enc := codec.NewEncoderBytes(&b, &mh)
			if err := enc.Encode(t.p); err != nil {
				log.Println(err)
			} else if _, err := t.c.Write(b); err != nil {
				log.Println(err)
			}
		}
	}
}

func closeListener() {
	if !closed {
		listener.Close()
		closed = true
	}
}

func closeConn(conn net.Conn) {
	if roomId, ok := getConnRoom(conn); ok {
		if room, ok := getRoom(roomId); ok {
			for id, peer := range room.peers {
				if peer.conn == conn {
					deletePeer(roomId, room, id)
					break
				}
			}
		}
		deleteConn(conn)
		conn.Close()
	}
}

func deletePeer(roomId string, r room, peerId string) {
	// remove peer from room
	delete(r.peers, peerId)
	if len(r.peers) == 0 {
		deleteRoom(roomId)
		return
	}

	// send update all but initiator
	pd := proto.NewPacketData(
		version,
		proto.UpdateRoom,
		proto.DataPacket{
			RoomId: roomId,
			Peers:  peersByRoomArray(r),
		})
	sendOthers(pd, r, peerId)
}

func generateRoomId() string {
	attempts := 10
	buff := make([]byte, 5)
	for {
		numRead, err := rand.Read(buff)
		if numRead != len(buff) || err != nil {
			panic(err)
		}
		id := fmt.Sprintf("%x", buff[:])
		// add new room if it doesn't exist
		if ok := setAbsentRoom(id, room{}); ok {
			return id
		}
		if attempts == 0 {
			break
		}
		attempts--
	}
	panic(errors.New("unable to generate unique room id"))
}

func getRoom(id string) (room, bool) {
	mur.Lock()
	r, ok := rooms[id]
	mur.Unlock()
	return r, (ok && r.peers != nil)
}

func setRoom(id string, r room) {
	mur.Lock()
	rooms[id] = r
	mur.Unlock()
}

// setAbsentRoom returns true if room was set
func setAbsentRoom(id string, r room) bool {
	mur.Lock()
	defer mur.Unlock()
	if _, ok := rooms[id]; !ok {
		rooms[id] = r
		return true
	}
	return false
}

func deleteRoom(id string) {
	mur.Lock()
	delete(rooms, id)
	mur.Unlock()
}

func getConnRoom(c net.Conn) (string, bool) {
	muc.Lock()
	id, ok := conns[c]
	muc.Unlock()
	return id, ok
}

func setConnRoom(c net.Conn, id string) {
	muc.Lock()
	conns[c] = id
	muc.Unlock()
}

func deleteConn(c net.Conn) {
	muc.Lock()
	delete(conns, c)
	muc.Unlock()
}
