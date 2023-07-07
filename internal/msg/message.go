package msg

import (
	"bytes"
	"encoding/binary"
)

const (
	TypeUnknown            uint32 = 0x123000
	TypeCreateRoomRequest  uint32 = 0x123001
	TypeCreateRoomResponse uint32 = 0x123002
	TypeJoinRoomRequest    uint32 = 0x123003
	TypeJoinRoomResponse   uint32 = 0x123004
)

const (
	MsgMagic uint32 = 0x847292
)

type baseMessage struct {
	Magic     uint32
	Type      uint32
	UTC       uint64
	ID        [16]byte
	Username  [16]byte
	Integrity [16]byte
}

type CreateRoomRequest struct {
	ID        string
	Username  string
	Integrity string
}

type CreateRoomResponse struct{}

type JoinRoomRequest struct {
	Room string
}

type JoinRoomResponse struct{}

func MessageType(data []byte) uint32 {
	if data == nil || len(data) < 64 {
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
	if len(data) < 64 {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	request := CreateRoomRequest{
		ID:        string(msg.ID[:]),
		Username:  string(msg.Username[:]),
		Integrity: string(msg.Integrity[:]),
	}
	return &request
}

func ParseCreateRoomResponse(data []byte) *CreateRoomResponse {
	if len(data) < 64 {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	// TODO: 填充response
	response := CreateRoomResponse{}
	return &response
}

func ParseJoinRoomRequest(data []byte) *JoinRoomRequest {
	if len(data) < 64 {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	// TODO: 填充request
	request := JoinRoomRequest{}
	return &request
}

func ParseJoinRoomResponse(data []byte) *JoinRoomResponse {
	if len(data) < 64 {
		return nil
	}
	reader := bytes.NewReader(data)
	var msg baseMessage
	binary.Read(reader, binary.LittleEndian, &msg)
	// TODO: 填充response
	response := JoinRoomResponse{}
	return &response
}

func NewCreateRoomResponse(room string) *CreateRoomResponse {
	return &CreateRoomResponse{
		// TODO:
	}
}

func NewJoinRoomResponse() *JoinRoomResponse {
	return &JoinRoomResponse{
		//TODO:
	}
}

func (m *CreateRoomRequest) ToBytes() []byte {
	return nil
}

func (m *CreateRoomResponse) ToBytes() []byte {
	return nil
}

func (m *JoinRoomRequest) ToBytes() []byte {
	return nil
}

func (m *JoinRoomResponse) ToBytes() []byte {
	return nil
}
