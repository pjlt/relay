package msg

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
)

const (
	TypeUnknown            uint32 = 0x123000
	TypeCreateRoomRequest  uint32 = 0x123001
	TypeCreateRoomResponse uint32 = 0x123002
	TypeJoinRoomRequest    uint32 = 0x123003
	TypeJoinRoomResponse   uint32 = 0x123004
)

const (
	Err_OK         int32 = 0
	Err_AuthFailed int32 = 1
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
	ID        [16]byte
	Username  [16]byte
	Room      [16]byte
	Integrity [16]byte
}

const baseMessageSize = 4*3 + 8 + 16*4

type CreateRoomRequest struct {
	ID        string
	Username  string
	Time      time.Time
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

func MessageType(data []byte) uint32 {
	if data == nil || len(data) != baseMessageSize {
		return TypeUnknown
	}
	reader := bytes.NewReader(data)
	var msgType uint32
	var magic uint32
	binary.Read(reader, binary.LittleEndian, &magic)
	if magic != MsgMagic {
		return TypeUnknown
	}
	binary.Read(reader, binary.LittleEndian, &msgType)
	if msgType == TypeCreateRoomRequest || msgType == TypeCreateRoomResponse || msgType == TypeJoinRoomRequest || msgType == TypeJoinRoomResponse {
		return msgType
	} else {
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

func ParseCreateRoomRequest(data []byte) *CreateRoomRequest {
	if len(data) < 84 {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	request := CreateRoomRequest{
		ID:        string(msg.ID[:]),
		Username:  string(msg.Username[:]),
		Time:      time.Unix(msg.Time, 0),
		Integrity: string(msg.Integrity[:]),
	}
	return &request
}

func ParseCreateRoomResponse(data []byte) *CreateRoomResponse {
	if len(data) != baseMessageSize {
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
	if len(data) != baseMessageSize {
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
		Username:  string(msg.Username[:]),
		Time:      time.Unix(msg.Time, 0),
		Room:      room,
		Integrity: string(msg.Integrity[:]),
	}
	return &request
}

func ParseJoinRoomResponse(data []byte) *JoinRoomResponse {
	if len(data) != baseMessageSize {
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

func (m *CreateRoomResponse) ToBytes() []byte {
	if len(m.ID) != 16 || len(m.Room) != 16 {
		return nil
	}
	var msg baseMessage
	msg.Errcode = m.ErrCode
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
	var msg baseMessage
	msg.Errcode = m.ErrCode
	copy(msg.ID[:], []byte(m.ID))
	msg.Room = m.Room
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, msg)
	// integrity?
	return buffer.Bytes()
}
