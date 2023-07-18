package session

import (
	"net"
	"relay/internal/msg"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Session struct {
	Room        uuid.UUID
	FirstAddr   *net.UDPAddr //向relay服务器申请room的地址
	SecondAddr  *net.UDPAddr //向relay服务器加入room的地址
	sendMessage SendFunc
}

func (s *Session) handlePacket(addr *net.UDPAddr, data []byte) {
	msgType := msg.MessageType(data)
	switch msgType {
	case msg.TypeCreateRoomRequest:
		s.handleCreateRoomRequest(addr, data)
	case msg.TypeJoinRoomRequest:
		s.handleJoinRoomRequest(addr, data)
	default:
		s.handleUnknownPacket(addr, data)
	}
}

func (s *Session) handleCreateRoomRequest(addr *net.UDPAddr, data []byte) {
	//Session已经创建，依然收到CreateRoomRequest，说明前面的Response可能已经丢失，需要重发一遍
}

func (s *Session) handleJoinRoomRequest(addr *net.UDPAddr, data []byte) {
	//Session已经创建，并且与SecondAddr的映射已经做完，依然收到JoinRoomRequest，说明前面的Response可能已经丢失，需要重发一遍
}

func (s *Session) handleUnknownPacket(addr *net.UDPAddr, data []byte) {
	// TODO: 限速
	if addr.String() == s.FirstAddr.String() {
		logrus.Debug("Relay message to SecondAddr")
		s.sendMessage(s.SecondAddr, data)
	} else if addr.String() == s.SecondAddr.String() {
		logrus.Debug("Relay message to FirstAddr")
		s.sendMessage(s.FirstAddr, data)
	} else {
		logrus.Debugf("Room(%s) recieved relay message from unknown address(%s)", s.Room, addr.String())
	}
}
