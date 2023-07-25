package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"net"
	"relay/internal/msg"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Authenticator struct {
	db *gorm.DB
}

// 结构体'User'默认对应数据库表'users'
type User struct {
	ID       string
	Username string
	Password string
}

func NewAuthenticator(path string) *Authenticator {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		logrus.Errorf("Failed to open sqlite database(%s): %v", path, err)
		return nil
	}
	return &Authenticator{
		db: db,
	}
}

func (a *Authenticator) Auth(addr *net.UDPAddr, request *msg.CreateRoomRequest, data []byte) int32 {
	// 如果不校验IP:Port，其他人捕获到合法的CreateRoomRequest包，发出一模一样的内容，也能使用relay服务器的资源
	if request.IP != binary.LittleEndian.Uint32(addr.IP) || request.Port != uint32(addr.Port) {
		logrus.Warnf("Packet(user:%s) address invalid", request.Username)
		return msg.Err_AddressInvalid
	}
	// 校验时间没有意义
	// gap := time.Now().Unix() - request.Time.Unix()
	// if gap < -3*60 || gap > 3*60 {
	// 	logrus.Warnf("Packet(user:%s) timestamp not in valid range: %d", request.Username, gap)
	// 	return msg.Err_TimeInvalid
	// }
	user := User{Username: request.Username}
	result := a.db.First(&user)
	if result.Error != nil {
		logrus.Errorf("Select table 'users' with {username:'%s'} failed with: %v", request.Username, result.Error)
		return msg.Err_AuthFailed
	}
	h := hmac.New(sha1.New, []byte(user.Password))
	h.Write(data)
	sum := string(h.Sum(nil))
	logrus.Debug("Integrity:", request.Integrity, ", Sum:", sum)
	if request.Integrity == sum {
		return msg.Err_OK
	} else {
		return msg.Err_AuthFailed
	}
}
