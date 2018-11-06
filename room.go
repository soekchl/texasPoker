package main

import (
	"math/rand"
	"sync"
	"time"

	. "github.com/soekchl/myUtils"
)

type Room struct {
	Id           int64
	Chip         int64                     // 筹码
	Play         bool                      // 游戏中
	BankerIndex  int                       // 庄家下标
	PlayUserList [MaxPlayCount]*OnlineUser // 游戏中玩家数据
	CommonPoker  []int32                   // 公共牌
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
	rand.Seed(time.Now().UnixNano())
}

// 房间开始游戏
func (room *Room) GameLoop() {
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
			canPlayCount := 0
			// 检查也参与游戏用户状态 --- 金额是否充足
			for _, v := range room.PlayUserList {
				if v.Money > 0 {
					v.Played = true
					canPlayCount++
				} else {
					v.Played = false
				}
			}

			if room.Status >= 10 && canPlayCount > 1 {
				for room.Status = 1; room.Status <= 10; room.Status++ {
					switch room.Status {
					case 3: // bet
						fallthrough
					case 5: // bet
						fallthrough
					case 7: // bet
						fallthrough
					case 9: // bet
						// TODO 轮流下注  轮流限时监听各个用户
					case 1: // 1-开始
						room.SetBlind()
					case 2: // 2-发手牌
						room.SendUserPoker()
					case 4: // 发3张底牌
						room.SendCommonPoker(3)
					case 6: // 发1张底牌
						fallthrough
					case 8: // 发1张底牌
						room.SendCommonPoker(1)
					case 10: // TODO 游戏结束 比拼胜负 奖池划分   公布结果 等待1秒
					}
				}
			}
		}
		time.Sleep(time.Second / 10)
	}
}

// 每个参与人发2张牌
func (room *Room) SendCommonPoker(count int) {
	for i := 0; i < count; i++ {
		room.CommonPoker = append(room.CommonPoker, room.getOnePoker())
	}
}

// 每个参与人发2张牌
func (room *Room) SendUserPoker() {
	for i := 0; i < MaxPlayCount; i++ {
		u := room.PlayUserList[i]
		if u != nil && u.Played {
			u.Poker[0] = room.getOnePoker()
			u.Poker[1] = room.getOnePoker()
		}
	}
}

// 从房间的扑克中获取一张
func (room *Room) getOnePoker() int32 {
	for {
		r := rand.Intn(52) + 1
		if room.Poker[r] != 0 {
			room.Poker[r] = 0
			return int32(r)
		}
	}
}

// 设置庄家，下 大盲注，小盲注
func (room *Room) SetBlind() {
	for {
		room.BankerIndex = (room.BankerIndex + 1) % MaxPlayCount
		if room.PlayUserList[room.BankerIndex] != nil && room.PlayUserList[room.BankerIndex].Played && room.PlayUserList[room.BankerIndex].Money > 0 {
			break
		}
	}
	bigBlindIndex := (room.BankerIndex - 1 + MaxPlayCount) % MaxPlayCount
	for {
		if room.PlayUserList[bigBlindIndex] != nil && room.PlayUserList[bigBlindIndex].Played && room.PlayUserList[bigBlindIndex].Money > 0 {
			break
		}
		bigBlindIndex = (bigBlindIndex - 1 + MaxPlayCount) % MaxPlayCount
	}
	smallBlindIndex := (bigBlindIndex - 1 + MaxPlayCount) % MaxPlayCount
	for {
		if room.PlayUserList[smallBlindIndex] != nil && room.PlayUserList[smallBlindIndex].Played && room.PlayUserList[smallBlindIndex].Money > 0 {
			break
		}
		smallBlindIndex = (smallBlindIndex - 1 + MaxPlayCount) % MaxPlayCount
	}

	if room.PlayUserList[bigBlindIndex].Money > room.Chip {
		room.PlayUserList[bigBlindIndex].Money -= room.Chip
		room.PlayUserList[bigBlindIndex].BetMoney += room.Chip
	} else {
		room.PlayUserList[bigBlindIndex].BetMoney = room.PlayUserList[bigBlindIndex].Money
		room.PlayUserList[bigBlindIndex].Money = 0
	}

	if room.PlayUserList[bigBlindIndex].Money > room.Chip/2 {
		room.PlayUserList[bigBlindIndex].Money -= room.Chip / 2
		room.PlayUserList[bigBlindIndex].BetMoney += room.Chip / 2
	} else {
		room.PlayUserList[bigBlindIndex].BetMoney = room.PlayUserList[bigBlindIndex].Money
		room.PlayUserList[bigBlindIndex].Money = 0
	}
}

// 房间数据初始化
func (room *Room) RoomInit() {
	room.AllBetMoney = 0
	room.Play = false
	room.CommonPoker = []int32{}
	room.Status = 10
	room.BankerIndex = 0 // 庄家下标 带个标记位往下移
	for i := 0; i < 53; i++ {
		room.Poker[i] = int32(i)
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
	room.RoomInit()

	RoomMapMutex.Lock()
	defer RoomMapMutex.Unlock()
	RoomMap[id] = room
	Notice("新房间 id = ", id%ShowRoomId, " 创建成功！")
	go room.GameLoop()
	return id
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

func DelRoom(Id int64) {
	RoomMapMutex.Lock()
	defer RoomMapMutex.Unlock()
	delete(RoomMap, Id)
}
