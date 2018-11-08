package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"texasPoker/mySocket"
	"time"

	. "github.com/soekchl/myUtils"
)

/*
	1、管理所有机器人开启和关闭
	2、让机器人 正常玩游戏
	3、当机器人 钱输光了 退出，在 1~2 补进去 并且 房间至少有2个位置

	4、先 完成所有情况全部跟住 概率弃牌
	4、按照底牌和自己的牌，概率 	跟住、加注、过、弃牌、allIn
*/

var (
	robotKey      = 1
	robotMap      = make(map[int]*mySocket.Session)
	robotMapMutex sync.RWMutex
)

func init() {
	go robotManage()
}

// 机器人管理
func robotManage() {
	time.Sleep(time.Second)
	// 初始化开启
	for i := 0; i < rand.Intn(4)+2; i++ {
		go robot(robotKey)
		robotKey++
	}
}

func robot(key int) {
	client, err := mySocket.Dial("tcp", serverPort, 5)
	if err != nil {
		Error(err)
		return
	}
	Notice("机器人 ", key, " 连接成功!")
	robotLoop(client, key)
}

func robotLoop(session *mySocket.Session, key int) {
	addRobot(key, session)
	defer func() {
		delRobot(key)
		session.Close()
	}()

	var mySeat int32 = 0 // 自己的位置
	var myInfo *UserInfo

	// 接收服务器信息 处理信息
	for {
		data, err := session.Receive()
		if err != nil {
			Error(err)
			return
		}
		ri := &RoomInfo{}
		err = json.Unmarshal(data.Body, ri)
		if err != nil {
			Error(err)
			return
		}
		mySeat = data.Seq
		Debug("Client Recv: ", ri.RoomStatus, " ", fmt.Sprintf("%#v", ri))
		for k, v := range ri.PlayUserList {
			if v.SeatNumber == mySeat {
				myInfo = &ri.PlayUserList[k]
			}
		}
		if ri.BetSeat == int(mySeat) && myInfo != nil && myInfo.Played {
			switch ri.RoomStatus {
			case 3: // bet
				fallthrough
			case 5: // bet
				fallthrough
			case 7: // bet
				fallthrough
			case 9:
				Debug(fmt.Sprintf("本轮我下注 座位号=%v 已下注=%v 总金额=%v 我的牌=%v 公共牌=%v",
					mySeat,
					myInfo.BetNowMoney,
					myInfo.BetAllMoney,
					showPokerNumber(myInfo.Poker),
					ri.CommonPoker))
			}
		}
	}
}

// 增加robot
func addRobot(key int, session *mySocket.Session) {
	robotMapMutex.Lock()
	defer robotMapMutex.Unlock()
	robotMap[key] = session
}

// 删除robot
func delRobot(key int) {
	robotMapMutex.Lock()
	defer robotMapMutex.Unlock()
	delete(robotMap, key)
}
