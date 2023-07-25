package session

import (
	"net"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Session struct {
	Room        uuid.UUID
	FirstAddr   *net.UDPAddr //向relay服务器申请room的地址
	SecondAddr  *net.UDPAddr //向relay服务器加入room的地址
	sendMessage SendFunc
}

func (s *Session) RelayPacket(addr *net.UDPAddr, data []byte) {
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
