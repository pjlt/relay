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

package server

import (
	"net"
	"os"
	"relay/internal/session"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	socket     *net.UDPConn
	stopChan   chan struct{}
	stopedChan chan struct{}
	sessionMgr *session.SessionManager
}

func New(ip string, port uint16) *Server {
	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		logrus.Errorf("Parse ip %s failed", ip)
		return nil
	}
	socket, err := net.ListenUDP("udp", &net.UDPAddr{IP: ipaddr, Port: int(port)})
	if err != nil {
		logrus.Errorf("ListenUDP on %s:%d failed: %v", ip, port, err)
		return nil
	}
	sessionMgr := session.NewManager()
	if sessionMgr == nil {
		return nil
	}
	svr := &Server{
		socket:     socket,
		stopChan:   make(chan struct{}),
		stopedChan: make(chan struct{}, 2),
		sessionMgr: sessionMgr,
	}
	svr.sessionMgr.SetSendFunc(svr.sendMessage)
	return svr
}

func (svr *Server) Start() {
	go svr.start()
}

func (svr *Server) Stop() {
	svr.stopChan <- struct{}{}
}

func (svr *Server) ReadTimeout() time.Duration {
	return 50 * time.Millisecond
}

func (svr *Server) StopedChan() chan struct{} {
	return svr.stopedChan
}

func (svr *Server) PrintStats() {

}

func (svr *Server) start() {
	defer func() {
		svr.stopedChan <- struct{}{}
	}()
	data := make([]byte, 65536)
	for {
		select {
		case <-svr.stopChan:
			return
		default:
		}
		svr.socket.SetDeadline(time.Now().Add(svr.ReadTimeout()))
		nread, remoteAddr, err := svr.socket.ReadFromUDP(data)
		if err != nil {
			if os.IsTimeout(err) {
				svr.sessionMgr.HandleIdle()
				continue
			} else {
				logrus.Errorf("ReadFromUDP error: %v", err)
				os.Exit(-1)
			}
		}
		if nread == 0 {
			continue
		}
		svr.sessionMgr.HandlePacket(remoteAddr, data[:nread])
	}
}

func (svr *Server) sendMessage(addr *net.UDPAddr, data []byte) {
	svr.socket.WriteToUDP(data, addr)
}
