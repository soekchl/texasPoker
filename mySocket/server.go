package mySocket

import (
	"fmt"
	"net"
)

type Server struct {
	listener     net.Listener
	handler      Handler
	sendChanSize int
}

type HandlerFunc func(*Session)

type General interface {
	ExecFunc(fd *FormatData, session *Session)
}

func (f HandlerFunc) HandleSession(session *Session) {
	f(session)
}

func Listen(network, address string, sendChanSize int, handler Handler) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, fmt.Errorf("[Listen] Error: %v", err)
	}
	return NewServer(listener, sendChanSize, handler), nil
}

func NewServer(listener net.Listener, sendChanSize int, handler Handler) *Server {
	return &Server{
		listener:     listener,
		handler:      handler,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Serve() error {
	for {
		conn, err := Accept(server.listener)
		if err != nil {
			return fmt.Errorf("[Server.Serve] Error: %v", err)
		}

		go func() {
			session := newSession(&conn, server.sendChanSize)
			server.handler.HandleSession(session) // 处理函数
		}()
	}
}

func (server *Server) Stop() {
	server.listener.Close()
}
