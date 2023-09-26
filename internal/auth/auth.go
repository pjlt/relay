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
	"relay/internal/db"
	"relay/internal/msg"

	"github.com/sirupsen/logrus"
)

type Authenticator struct {
	//db *gorm.DB
}

func NewAuthenticator(path string) *Authenticator {
	return &Authenticator{
		//db: db.DB,
	}
}

func (a *Authenticator) Auth(addr *net.UDPAddr, request *msg.CreateRoomRequest, data []byte) int32 {
	// 如果不校验IP:Port，其他人捕获到合法的CreateRoomRequest包，发出一模一样的内容，也能使用relay服务器的资源
	if request.IP != binary.LittleEndian.Uint32(addr.IP) || request.Port != uint32(addr.Port) {
		logrus.Warnf("Packet(user:%s) address invalid", request.Username)
		return msg.Err_AddressInvalid
	}
	user, err := db.QueryByUserName(request.Username)
	if err != nil {
		return msg.Err_AuthFailed
	}
	h := hmac.New(sha1.New, []byte(user.Key))
	h.Write(data)
	sum := string(h.Sum(nil))
	logrus.Debug("Integrity:", request.Integrity, ", Sum:", sum)
	if request.Integrity == sum {
		return msg.Err_OK
	} else {
		return msg.Err_AuthFailed
	}
}
