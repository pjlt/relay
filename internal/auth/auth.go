package auth

import (
	"net"
	"relay/internal/msg"
)

type Authenticator interface {
	Stop()
	Auth(addr *net.UDPAddr, request *msg.CreateRoomRequest, data []byte) int32
	Token() string
}
