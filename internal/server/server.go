package server

import (
	"net"

	"github.com/sirupsen/logrus"
)

type Server struct {
	listener *net.UDPConn
	stopChan chan struct{}
}

func New(ip string, port uint16) *Server {
	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		logrus.Errorf("Parse ip %s failed", ip)
		return nil
	}
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: ipaddr, Port: int(port)})
	if err != nil {
		logrus.Errorf("ListenUDP on %s:%d failed: %v", ip, port, err)
		return nil
	}
	svr := &Server{
		listener: listener,
		stopChan: make(chan struct{}),
	}
	return svr
}

func (svr *Server) Start() {
	go svr.start()
}

func (svr *Server) Stop() {
	svr.stopChan <- struct{}{}
}

func (svr *Server) PrintStats() {

}

func (svr *Server) start() {
	//
}
