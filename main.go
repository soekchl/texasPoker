// testNlmt project main.go
package main

import (
	"texasPoker/mySocket"
	"time"

	. "github.com/soekchl/myUtils"
)

func main() {
	go func() {
		time.Sleep(time.Second)
		Client()
	}()

	Server()

}

func Client() {
	client, err := mySocket.Dial("tcp", ":1111", 5)
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
	server, err := mySocket.Listen("tcp", ":1111", 5, mySocket.HandlerFunc(serverLoop))
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
