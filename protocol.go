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
	Played       bool         // 游戏中
	LeaveTime    int32        // 剩余等待时间 秒 四舍五入
	PlayUserList []OnlineUser // 游戏中玩家数据 只发送参与中的人的数据
}
