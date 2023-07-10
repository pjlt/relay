package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"time"

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

func (a *Authenticator) Auth(username string, integrity string, timestamp time.Time, data []byte) bool {
	user := User{Username: username}
	result := a.db.First(&user)
	if result.Error != nil {
		logrus.Errorf("Select table 'users' with {username:'%s'} failed with: %v", username, result.Error)
		return false
	}
	gap := time.Now().Unix() - timestamp.Unix()
	if gap < -3*60 || gap > 3*60 {
		logrus.Warnf("Packet(user:%s) timestamp not in valid range: %d", username, gap)
		return false
	}
	h := hmac.New(sha1.New, []byte(user.Password))
	h.Write(data)
	sum := string(h.Sum(nil))
	return integrity == sum
}
