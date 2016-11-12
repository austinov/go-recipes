package proto

import "github.com/ugorji/go/codec"

// Actions for communicate with server
const (
	// this action used to book room for chat and join user to it
	BookRoom byte = iota
	// this action iused to join user to the room
	JoinRoom
	// this action used to send message to users in the room
	SendMsg
	// this action used to update room info (e.g. list of peers)
	UpdateRoom
	// this action used to disconnect users from chat
	LeaveRoom
)

type Packet struct {
	Version byte
	Action  byte
	Data    DataPacket
	Err     string
}

type DataPacket struct {
	RoomId   string
	Peers    []string
	PeerName string
	PeerId   string
	Message  string
	Sender   string
}

func NewPacket(version, action byte) Packet {
	return Packet{
		Version: version,
		Action:  action,
	}
}

func NewPacketData(version, action byte, data DataPacket) Packet {
	return Packet{
		Version: version,
		Action:  action,
		Data:    data,
	}
}

func NewPacketError(version, action byte, err error) Packet {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	return Packet{
		Version: version,
		Action:  action,
		Err:     errStr,
	}
}

func Decode(b []byte) (Packet, error) {
	var (
		p  Packet
		mh codec.MsgpackHandle
	)
	dec := codec.NewDecoderBytes(b, &mh)
	err := dec.Decode(&p)
	return p, err
}
