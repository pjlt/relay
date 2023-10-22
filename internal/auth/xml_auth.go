package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"net"
	"relay/internal/common"
	"relay/internal/conf"
	"relay/internal/msg"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type XmlAuthenticator struct {
	lastToken     string
	currToken     string
	validDuration time.Duration
	stopChan      chan struct{}
	mutex         sync.Mutex
	users         map[string]string
}

func NewXmlAuthenticator() Authenticator {
	token := common.RandStr(common.Fixed16)
	a := &XmlAuthenticator{
		lastToken:     token,
		currToken:     token,
		validDuration: time.Second * 5,
		users:         make(map[string]string),
	}
	if !a.init() {
		return nil
	}
	go a.changeToken()
	return a
}

func (a *XmlAuthenticator) changeToken() {
	ticker := time.NewTicker(a.validDuration)
	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.mutex.Lock()
			a.lastToken = a.currToken
			a.currToken = common.RandStr(common.Fixed16)
			a.mutex.Unlock()
		}
	}
}

func (a *XmlAuthenticator) init() bool {
	if len(conf.Xml.Auth.Users) == 0 {
		return false
	}
	for i := 0; i < len(conf.Xml.Auth.Users); i++ {
		length := len(conf.Xml.Auth.Users[i].Username)
		if length > common.Fixed16 {
			logrus.Errorf("Username '%s' too long, must be <= 16 bytes", conf.Xml.Auth.Users[i].Username)
			return false
		}
		length = len(conf.Xml.Auth.Users[i].Password)
		if length > common.Fixed16 {
			logrus.Errorf("Password for user '%s' too long, must be <= 16 bytes", conf.Xml.Auth.Users[i].Username)
			return false
		}
		_, exists := a.users[conf.Xml.Auth.Users[i].Password]
		if exists {
			logrus.Errorf("Username '%s' duplicated", conf.Xml.Auth.Users[i].Username)
			return false
		}
		a.users[conf.Xml.Auth.Users[i].Username] = conf.Xml.Auth.Users[i].Password
	}
	return true
}

func (a *XmlAuthenticator) Stop() {
	a.stopChan <- struct{}{}
}

func (a *XmlAuthenticator) Token() string {
	a.mutex.Lock()
	token := a.currToken
	a.mutex.Unlock()
	return token
}

func (a *XmlAuthenticator) Auth(addr *net.UDPAddr, request *msg.CreateRoomRequest, data []byte) int32 {
	a.mutex.Lock()
	lastToken := a.lastToken
	currToken := a.currToken
	a.mutex.Unlock()
	// 校验Token
	if lastToken != request.Token && currToken != request.Token {
		logrus.Warnf("Packet(user:%s) token invalid", request.Username)
		return msg.Err_AuthFailed
	}
	// 如果不校验IP:Port，其他人捕获到合法的CreateRoomRequest包，发出一模一样的内容，也能使用relay服务器的资源
	if request.IP != binary.LittleEndian.Uint32(addr.IP) || request.Port != uint32(addr.Port) {
		logrus.Warnf("Packet(user:%s) address invalid", request.Username)
		return msg.Err_AddressInvalid
	}
	// 校验hmac
	passwd, exists := a.users[request.Username]
	if !exists {
		return msg.Err_AuthFailed
	}
	h := hmac.New(sha1.New, []byte(passwd))
	h.Write(data)
	sum := string(h.Sum(nil))
	logrus.Debug("Integrity:", request.Integrity, ", Sum:", sum)
	if request.Integrity == sum {
		return msg.Err_OK
	} else {
		return msg.Err_AuthFailed
	}
}
