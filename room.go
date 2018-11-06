package main

import (
	"sync"
	"time"

	. "github.com/soekchl/myUtils"
)

type Room struct {
	Id           int64
	Chip         int32                     // 筹码
	Play         bool                      // 游戏中
	PlayUserList [MaxPlayCount]*OnlineUser // 游戏中玩家数据
	CommonPoker  [5]int32                  // 公共牌
	Status       int32                     // 1-开始 2-发手牌 3-Bet 4-发底牌(3) 5-Bet 6-发底牌(1) 7-Bet 8-发底牌(1) 9-Bet 10-Over
	AllBetMoney  int64                     // 总下注金额
	Poker        [53]int32                 // 1~52 扑克牌
}

const (
	ShowRoomId   = 10000
	TimeOutLimit = 6000
)

var (
	chanPlayGame = make(chan *OnlineUser, 32) // 参与游戏数据
	RoomMap      = make(map[int64]*Room)      // 房间列表
	RoomMapMutex sync.RWMutex                 // 房间锁
)

func init() {
	go CreateNewRoom()
}

// 房间开始游戏
func (room *Room) PlayGame() {
	Warn("房间监控 id = ", room.Id%ShowRoomId, " 开启")
	n := 0
	for {
		if len(room.PlayUserList) < 1 {
			n++
			if n >= TimeOutLimit { // 超过一定时间没有用户进入房间 删除房间
				go DelRoom(room.Id)
				return
			}
		} else { // 用户进入
			n = 0
			// 1-开始 2-发手牌 3-Bet 4-发底牌(3) 5-Bet 6-发底牌(1) 7-Bet 8-发底牌(1) 9-Bet 10-Over
			if room.Status == 10 && len(room.PlayUserList) > 1 {
				for room.Status = 1; room.Status < 10; {
					// 1-修改用户状态，下注等信息
					// 2-发底牌
				}
			}
		}
		time.Sleep(time.Second / 10)
	}
}

// 整个房间队列
func RoomLoop() {
	for {
		select {
		case newPlay := <-chanPlayGame:
			go JoinGame(newPlay)
		}
	}
}

// 参与游戏
func JoinGame(user *OnlineUser) {
	RoomMapMutex.RLock()
	defer RoomMapMutex.RUnlock()
	for k, v := range RoomMap {
		if len(v.PlayUserList) < MaxPlayCount {
			for kk, vv := range v.PlayUserList {
				if vv == nil {
					user.RoomId = k
					v.PlayUserList[kk] = user
					// TODO 用户进入房间信息 反馈给用户
					return
				}
			}
		}
	}
	// 运行到这里就意味着没有房间
	go CreateNewRoom() // 创建房间
	time.Sleep(time.Second / 10)
	go JoinGame(user) // 从新运行加入
}

// 创建新的房间
func CreateNewRoom() int64 {
	id := time.Now().UnixNano()
	room := &Room{
		Id:   id,
		Chip: 5,
	}
	RoomInit(room)

	RoomMapMutex.Lock()
	defer RoomMapMutex.Unlock()
	RoomMap[id] = room
	Notice("新房间 id = ", id%ShowRoomId, " 创建成功！")
	go room.PlayGame()
	return id
}

// 房间数据初始化
func RoomInit(room *Room) {
	room.AllBetMoney = 0
	room.Play = false
	room.CommonPoker = [5]int32{0, 0, 0, 0, 0}
	room.Status = 10
	for i := 0; i < 53; i++ {
		room.Poker[i] = int32(i)
	}
}

func DelRoom(Id int64) {
	RoomMapMutex.Lock()
	defer RoomMapMutex.Unlock()
	delete(RoomMap, Id)
}
