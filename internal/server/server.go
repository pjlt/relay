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
	svr := &Server{
		socket:     socket,
		stopChan:   make(chan struct{}),
		stopedChan: make(chan struct{}, 2),
		sessionMgr: session.NewManager(),
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
