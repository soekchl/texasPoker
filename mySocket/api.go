package mySocket

import (
	"fmt"
	"net"
	"time"
)

type Handler interface {
	HandleSession(*Session)
}

func Dial(network, address string, sendChanSize int) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("[Dial] Error: %v", err)
	}

	return newSession(&conn, sendChanSize), nil
}

func DialTimeout(network, address string, timeout time.Duration, sendChanSize int) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, fmt.Errorf("[DialTimeout] Error: %v", err)
	}

	return newSession(&conn, sendChanSize), nil
}

func Accept(listener net.Listener) (net.Conn, error) {
	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			if err != nil {
				err = fmt.Errorf("[Accept] Error: %v", err)
			}
			return nil, err
		}
		return conn, nil
	}
}
