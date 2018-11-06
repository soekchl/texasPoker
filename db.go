package main

import (
	"texasPoker/mySocket"
)

const (
	MaxPlayCount = 10
)

type User struct {
	Id    int64
	Money int64
}

type OnlineUser struct {
	User
	RoomId     int64
	Poker      [2]int32          // 手牌
	SeatNumber int32             // 座位号 1-n
	Played     bool              // true-参与 false-旁观&弃牌
	BetMoney   int64             // 下注金额
	session    *mySocket.Session // 连接socket
}
