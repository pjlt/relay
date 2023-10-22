/*
 * BSD 3-Clause License
 *
 * Copyright (c) 2023 Zhennan Tu <zhennan.tu@gmail.com>
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 *    list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from
 *    this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"net"
	"relay/internal/common"
	"relay/internal/db"
	"relay/internal/msg"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type DBAuthenticator struct {
	lastToken     string
	currToken     string
	validDuration time.Duration
	stopChan      chan struct{}
	mutex         sync.Mutex
}

func NewDBAuthenticator(path string) Authenticator {
	token := common.RandStr(common.Fixed16)
	a := &DBAuthenticator{
		lastToken:     token,
		currToken:     token,
		validDuration: time.Second * 5,
	}
	go a.changeToken()
	return a
}

func (a *DBAuthenticator) changeToken() {
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

func (a *DBAuthenticator) Stop() {
	a.stopChan <- struct{}{}
}

func (a *DBAuthenticator) Token() string {
	a.mutex.Lock()
	token := a.currToken
	a.mutex.Unlock()
	return token
}

func (a *DBAuthenticator) Auth(addr *net.UDPAddr, request *msg.CreateRoomRequest, data []byte) int32 {
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
	user, err := db.QueryByUserName(request.Username)
	if err != nil {
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
