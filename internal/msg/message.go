package msg

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	TypeUnknown            uint32 = 0x123000
	TypeCreateRoomRequest  uint32 = 0x123001
	TypeCreateRoomResponse uint32 = 0x123002
	TypeJoinRoomRequest    uint32 = 0x123003
	TypeJoinRoomResponse   uint32 = 0x123004
	TypeReflexRequest      uint32 = 0x124001
	TypeReflexResponse     uint32 = 0x124002
)

const (
	Err_OK             int32 = 0
	Err_AuthFailed     int32 = 1
	Err_AddressInvalid int32 = 2
	Err_TimeInvalid    int32 = 3
)

const (
	MsgMagic uint32 = 0x847292df
)

// 暂时简化，实际布局共用一个结构体，有些消息没用到的字段就忽略
type baseMessage struct {
	Magic     uint32
	Type      uint32
	Errcode   int32
	Time      int64
	IP        uint32
	Port      uint32
	ID        [16]byte
	Username  [16]byte
	Room      [16]byte
	Integrity [20]byte
}

const BaseMessageSize = 4*5 + 8 /*Time*/ + 16*3 + 20 /*Integrity*/
const IntegritySize = 20

type CreateRoomRequest struct {
	ID        string
	Username  string
	Time      time.Time
	IP        uint32
	Port      uint32
	Integrity string
}

type CreateRoomResponse struct {
	ID      string
	ErrCode int32
	Room    uuid.UUID
}

type JoinRoomRequest struct {
	ID        string
	Username  string
	Time      time.Time
	Room      uuid.UUID
	Integrity string
}

type JoinRoomResponse struct {
	ID      string
	ErrCode int32
	Room    uuid.UUID
}

type ReflexResponse struct {
	Addr *net.UDPAddr
}

func clen(data []byte) int {
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return i
		}
	}
	return len(data)
}

func MessageType(data []byte) uint32 {
	if data == nil {
		logrus.Debug("Received data == nil")
		return TypeUnknown
	}
	if len(data) != BaseMessageSize {
		logrus.Debugf("len(data) == %d != BaseMessageSize == %d", len(data), BaseMessageSize)
		return TypeUnknown
	}
	reader := bytes.NewReader(data)
	var msgType uint32
	var magic uint32
	binary.Read(reader, binary.LittleEndian, &magic)
	if magic != MsgMagic {
		logrus.Debugf("magic != MsgMagic")
		return TypeUnknown
	}
	binary.Read(reader, binary.LittleEndian, &msgType)
	if msgType == TypeCreateRoomRequest ||
		msgType == TypeCreateRoomResponse ||
		msgType == TypeJoinRoomRequest ||
		msgType == TypeJoinRoomResponse ||
		msgType == TypeReflexRequest ||
		msgType == TypeReflexResponse {
		return msgType
	} else {
		logrus.Debugf("msgType == 0x%x", msgType)
		return TypeUnknown
	}
}

func IsCreateRoomRequest(data []byte) bool {
	return MessageType(data) == TypeCreateRoomRequest
}

func IsCreateRoomResponse(data []byte) bool {
	return MessageType(data) == TypeCreateRoomResponse
}

func IsJoinRoomRequest(data []byte) bool {
	return MessageType(data) == TypeJoinRoomRequest
}

func IsJoinRoomResponse(data []byte) bool {
	return MessageType(data) == TypeJoinRoomResponse
}

func IsReflexRequest(data []byte) bool {
	return MessageType(data) == TypeReflexRequest
}

func IsReflexResponse(data []byte) bool {
	return MessageType(data) == TypeReflexResponse
}

func ParseCreateRoomRequest(data []byte) *CreateRoomRequest {
	if len(data) != BaseMessageSize {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	request := CreateRoomRequest{
		ID:        string(msg.ID[:]),
		Username:  string(msg.Username[:clen(msg.Username[:])]),
		Time:      time.Unix(msg.Time, 0),
		IP:        msg.IP,
		Port:      msg.Port,
		Integrity: string(msg.Integrity[:]),
	}
	return &request
}

func ParseCreateRoomResponse(data []byte) *CreateRoomResponse {
	if len(data) != BaseMessageSize {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	room, err := uuid.FromBytes(msg.Room[:])
	if err != nil {
		return nil
	}
	response := CreateRoomResponse{
		ID:      string(msg.ID[:]),
		ErrCode: msg.Errcode,
		Room:    room,
	}
	return &response
}

func ParseJoinRoomRequest(data []byte) *JoinRoomRequest {
	if len(data) != BaseMessageSize {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	room, err := uuid.FromBytes(msg.Room[:])
	if err != nil {
		return nil
	}
	request := JoinRoomRequest{
		ID:        string(msg.ID[:]),
		Username:  string(msg.Username[:clen(msg.Username[:])]),
		Time:      time.Unix(msg.Time, 0),
		Room:      room,
		Integrity: string(msg.Integrity[:]),
	}
	return &request
}

func ParseJoinRoomResponse(data []byte) *JoinRoomResponse {
	if len(data) != BaseMessageSize {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	room, err := uuid.FromBytes(msg.Room[:])
	if err != nil {
		return nil
	}
	response := JoinRoomResponse{
		ID:      string(msg.ID[:]),
		ErrCode: msg.Errcode,
		Room:    room,
	}
	return &response
}

func NewCreateRoomResponse(ID string, errCode int32, room uuid.UUID) *CreateRoomResponse {
	// 是否加入校验？
	return &CreateRoomResponse{
		ID:      ID,
		ErrCode: errCode,
		Room:    room,
	}
}

func NewJoinRoomResponse(ID string, errCode int32, room uuid.UUID) *JoinRoomResponse {
	// 是否加入校验？
	return &JoinRoomResponse{
		ID:      ID,
		ErrCode: errCode,
		Room:    room,
	}
}

func NewReflexResponse(addr *net.UDPAddr) *ReflexResponse {
	return &ReflexResponse{
		Addr: addr,
	}
}

func (m *CreateRoomResponse) ToBytes() []byte {
	if len(m.ID) != 16 || len(m.Room) != 16 {
		return nil
	}
	msg := baseMessage{
		Magic:   MsgMagic,
		Type:    TypeCreateRoomResponse,
		Errcode: m.ErrCode,
		Time:    time.Now().Unix(),
	}
	copy(msg.ID[:], []byte(m.ID))
	copy(msg.Room[:], m.Room[:])
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, msg)
	// integrity?
	return buffer.Bytes()
}

func (m *JoinRoomResponse) ToBytes() []byte {
	if len(m.ID) != 16 || len(m.Room) != 16 {
		return nil
	}
	msg := baseMessage{
		Magic:   MsgMagic,
		Type:    TypeCreateRoomResponse,
		Errcode: m.ErrCode,
		Time:    time.Now().Unix(),
		Room:    m.Room,
	}
	copy(msg.ID[:], []byte(m.ID))
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, msg)
	// integrity?
	return buffer.Bytes()
}

func (m *ReflexResponse) ToBytes() []byte {
	msg := baseMessage{
		Magic: MsgMagic,
		Type:  TypeReflexResponse,
		IP:    binary.LittleEndian.Uint32(m.Addr.IP),
		Port:  uint32(m.Addr.Port),
	}
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, msg)
	return buffer.Bytes()
}
