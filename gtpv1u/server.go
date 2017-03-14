package gtpv1u

import (
	"net"
	"time"
	"sync"
	"fmt"
)

type Server struct {
	addr 			string
	service 		Service
	ch          chan struct{}
	waitGroup * sync.WaitGroup
}

type Service interface {
	GTPPacketHandle(req * Packet, conn * net.UDPConn, addr net.Addr) error
}

func NewServer(addr string, service Service) * Server {
	s := &Server{
		addr: 		addr,
		service: 	service,
		ch:			make(chan struct{}),
		waitGroup:	&sync.WaitGroup{},
	}
	return s
}

func (s *Server) ListenAndServe(errch chan error) error {
	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	defer conn.Close()

	for {
		select {
		case <- s.ch:
			return nil
		default:
		}
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		b := make([]byte, 4096)
		n, addr, err := conn.ReadFrom(b)
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Timeout() {
				continue
			}
			return err
		}

		s.waitGroup.Add(1)
		go func(p []byte, addr net.Addr) {
			defer s.waitGroup.Done()

			pac,err := DecodePacket(b)
			if err != nil {
				errch <- fmt.Errorf("[GTP Packet Decode] %s", err.Error())
				return
			}

			s.service.GTPPacketHandle(pac, conn, addr)
		}(b[:n], addr)
	}
	return nil
}
