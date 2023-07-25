package session

import (
	"net"
	"relay/internal/auth"
	"relay/internal/conf"
	"relay/internal/msg"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SendFunc func(addr *net.UDPAddr, data []byte)

type SessionManager struct {
	addrToSessions map[string]*Session
	roomToSessions map[string]*Session
	sendMessage    SendFunc
	authenticator  *auth.Authenticator
}

func NewManager() *SessionManager {
	authenticator := auth.NewAuthenticator(conf.Xml.DB.Path)
	if authenticator == nil {
		return nil
	}
	return &SessionManager{
		addrToSessions: make(map[string]*Session),
		roomToSessions: make(map[string]*Session),
		authenticator:  authenticator,
	}
}

func (mgr *SessionManager) SetSendFunc(sendFunc SendFunc) {
	mgr.sendMessage = sendFunc
}

func (mgr *SessionManager) HandlePacket(addr *net.UDPAddr, data []byte) {
	addrStr := addr.String()
	s, exists := mgr.addrToSessions[addrStr]
	if !exists {
		if msg.IsReflexRequest(data) {
			mgr.handleReflexRequest(addr, data)
			return
		} else if msg.IsCreateRoomRequest(data) {
			mgr.handleCreateRoomRequest(addr, data)
			return
		} else if msg.IsJoinRoomRequest(data) {
			mgr.handleJoinRoomRequest(addr, data)
			return
		}
		logrus.Debugf("Received packet from %s, but it isn't CreateRoomRequest/JoinRoomRequest", addrStr)
	} else {
		s.handlePacket(addr, data)
	}
}

func (mgr *SessionManager) handleCreateRoomRequest(addr *net.UDPAddr, data []byte) {
	request := msg.ParseCreateRoomRequest(data)
	if request == nil {
		logrus.Debugf("ParseCreateRoomRequest failed")
		return
	}
	errCode := mgr.authenticator.Auth(addr, request, data[:msg.BaseMessageSize-msg.IntegritySize])
	if errCode != msg.Err_OK {
		response := msg.NewCreateRoomResponse(request.ID, errCode, [16]byte{})
		mgr.sendMessage(addr, response.ToBytes())
		return
	}
	var roomUUID uuid.UUID
	var roomStr string
	for {
		roomUUID = uuid.New()
		roomStr = roomUUID.String()
		if _, exists := mgr.roomToSessions[roomStr]; exists {
			continue
		}
		break
	}
	s := &Session{
		Room:        roomUUID,
		FirstAddr:   addr,
		sendMessage: mgr.sendMessage,
	}
	mgr.addrToSessions[addr.String()] = s
	mgr.roomToSessions[roomStr] = s
	response := msg.NewCreateRoomResponse(request.ID, msg.Err_OK, roomUUID)
	logrus.Infof("Send CreateRoomResponse(%s) to %s", roomStr, addr.String())
	mgr.sendMessage(addr, response.ToBytes())
}

func (mgr *SessionManager) handleJoinRoomRequest(addr *net.UDPAddr, data []byte) {
	request := msg.ParseJoinRoomRequest(data)
	if request == nil {
		logrus.Debugf("ParseJoinRoomRequest failed")
		return
	}
	var s *Session
	var exists bool
	if s, exists = mgr.roomToSessions[request.Room.String()]; !exists {
		logrus.Debugf("Received JoinRoomRequest with invalid room id:%s", request.Room)
		return
	}
	s.SecondAddr = addr
	// TODO: 鉴权
	if _, exists = mgr.addrToSessions[addr.String()]; exists {
		logrus.Errorf("Received JoinRoomRequest(room:%s) from addrress(%s), but %s already exist for another session", request.Room, addr.String(), addr.String())
		return
	}
	mgr.addrToSessions[addr.String()] = s
	response := msg.NewJoinRoomResponse(request.ID, msg.Err_OK, request.Room)
	logrus.Infof("Send JoinRoomResponse(%s) to %s", request.Room.String(), addr.String())
	mgr.sendMessage(addr, response.ToBytes())
}

func (mgr *SessionManager) handleReflexRequest(addr *net.UDPAddr, data []byte) {
	response := msg.NewReflexResponse(addr)
	logrus.Debugf("Send ReflexResponse to %s", addr.String())
	mgr.sendMessage(addr, response.ToBytes())
}
