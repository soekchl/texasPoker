// testNlmt project main.go
package main

import (
	"math/rand"
	"sync/atomic"
	"texasPoker/mySocket"
	"time"

	. "github.com/soekchl/myUtils"
)

const (
	serverPort = ":1234"
)

var (
	userId int64 = 1001
)

func main() {
	Server()
	rand.Seed(time.Now().UnixNano())
}

func Client() {
	client, err := mySocket.Dial("tcp", serverPort, 5)
	if err != nil {
		Error(err)
		return
	}
	Notice("客户端连接成功!")
	clientLoop(client)
}

func clientLoop(session *mySocket.Session) {
	data := &mySocket.FormatData{
		Id:   1,
		Seq:  2,
		Body: []byte{1, 3, 4, 5, 4},
	}
	err := session.Send(data)
	if err != nil {
		Error(err)
		return
	}
	Notice("Send Ok!")

	data, err = session.Receive()
	if err != nil {
		Error(err)
		return
	}

	Notice("Client Recv: ", data)

}

func Server() {
	server, err := mySocket.Listen("tcp", serverPort, 5, mySocket.HandlerFunc(serverLoop))
	if err != nil {
		Error(err)
		return
	}
	Notice("服务器开启！")
	server.Serve()
}

func serverLoop(session *mySocket.Session) {
	defer session.Close()
	Notice("服务器 接收连接：", session.RemoteAddr())
	chanPlayGame <- &OnlineUser{
		id:      getUserId(),
		Money:   1000,
		session: session,
	}
	for {
		data, err := session.Receive()
		if err != nil {
			Error(err)
			return
		}
		Notice(data)
		session.Send(data)
	}
}

func getUserId() int64 {
	return atomic.AddInt64(&userId, 1)
}
