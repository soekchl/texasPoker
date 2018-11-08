package main

// 主要是通讯协议 使用json
/*
	1		心跳
	2-100 	保留

	client 接入
	server 分配房间
	client 等待开始
	server 发起开始 接收几次信息
	client 下注 等信息
*/

//	client->server (连接进来 自动进入房间)
//		1001	操作-（下注，allIn，check，弃牌）
type ClientCmd struct {
	Bet    int64 // 下注金额
	AllIn  bool  // 全部压入
	Check  bool  // 过
	GiveUp bool  // 弃牌
}

//	server->client
//		2001	房间信息反馈(整体数据，所有人筹码，下注金额，当前下注人，等待时间 等等)
type RoomInfo struct {
	BetSeat      int        // 下注中用户座位号
	LeaveTime    int        // 剩余等待时间 秒
	RoomStatus   int        // 房间状态
	CommonPoker  []int32    // 公共牌
	MinBet       int64      // 本轮最小下注额度
	PlayUserList []UserInfo // 游戏中玩家数据
}

type UserInfo struct {
	Money       int64   // 剩余金额
	Poker       []int32 // 手牌
	SeatNumber  int32   // 座位号 1-n
	Played      bool    // true-参与 false-旁观&弃牌
	BetAllMoney int64   // 下注总金额
	BetNowMoney int64   // 当轮下注金额
	BetOk       bool    // 本轮下注完成
}
