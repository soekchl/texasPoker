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

type Room struct {
	Id           int64
	Chip         int64                     // 筹码
	Play         bool                      // 游戏中
	BankerIndex  int                       // 庄家下标
	PlayUserList [MaxPlayCount]*OnlineUser // 游戏中玩家数据
	CommonPoker  []int32                   // 公共牌
	Status       int                       // 1-开始 2-发手牌 3-Bet 4-发底牌(3) 5-Bet 6-发底牌(1) 7-Bet 8-发底牌(1) 9-Bet 10-Over
	MinBet       int64                     // 当前轮最小下注额度
	AllBetMoney  int64                     // 总下注金额
	Poker        [53]int32                 // 1~52 扑克牌
	LeaveTimes   time.Duration             // 剩余时间
}

const (
	ShowRoomId   = 10000000
	TimeOutLimit = 6000
)

var (
	chanPlayGame = make(chan *OnlineUser, 32) // 参与游戏数据
	RoomMap      = make(map[int64]*Room)      // 房间列表
	RoomMapMutex sync.RWMutex                 // 房间锁
	statusToName = make(map[int]string, 12)   // 房间状态
)

func init() {
	go CreateNewRoom()
	go RoomLoop()
	rand.Seed(time.Now().UnixNano())

	statusToName[1] = "开始游戏"
	statusToName[2] = "发手牌"
	statusToName[3] = "下注"
	statusToName[4] = "发3张底牌"
	statusToName[5] = "下注"
	statusToName[6] = "发一张底牌"
	statusToName[7] = "下注"
	statusToName[8] = "发一张底牌"
	statusToName[9] = "下注"
	statusToName[10] = "游戏结算"
}

// 房间开始游戏
func (room *Room) GameLoop() {
	Warn("房间监控 id = ", room.Id%ShowRoomId, " 开启")
	n := 0
	for {
		count := 0
		for _, v := range room.PlayUserList {
			if v != nil {
				count++
			}
		}
		if count < 1 {
			n++
			if n >= TimeOutLimit { // 超过一定时间没有用户进入房间 删除房间
				go DelRoom(room.Id)
				return
			}
			if n%100 == 0 {
				Info("房间 id = ", room.Id%ShowRoomId, " 人数=", count)
			}
		} else { // 用户进入
			Debug("房间 id = ", room.Id%ShowRoomId, " 人数=", count)
			n = 0
			canPlayCount := 0
			// 检查也参与游戏用户状态 --- 金额是否充足
			for _, v := range room.PlayUserList {
				if v == nil {
					continue
				}

				if v.Money > 0 {
					v.Played = true
					canPlayCount++
				} else {
					v.Played = false
				}
			}

			if canPlayCount > 1 {
				for room.Status = 1; room.Status <= 10; room.Status++ {
					Notice("房间 id = ", room.Id%ShowRoomId, " 状态 = ", room.Status, " ", statusToName[room.Status])
					switch room.Status {
					case 3: // bet
						fallthrough
					case 5: // bet
						fallthrough
					case 7: // bet
						fallthrough
					case 9:
						room.Bet()
						room.MinBet = 0 // 下注好后清空
					case 1: // 1-开始
						room.SetBlind()
						room.SendDataToClient(0, false)
					case 2: // 2-发手牌
						room.SendUserPoker()
						room.SendDataToClient(0, false)
					case 4: // 发3张底牌
						room.SendCommonPoker(3)
					case 6: // 发1张底牌
						fallthrough
					case 8: // 发1张底牌
						room.SendCommonPoker(1)
					case 10: // TODO 游戏结束 比拼胜负 奖池划分   公布结果 等待1秒
					}

					Debug(fmt.Sprintf("%#v", room))
					for _, v := range room.PlayUserList {
						if v != nil {
							Debug(fmt.Sprintf("%#v", v))
						}
					}
				}
			}
		}
		time.Sleep(time.Second / 10)
	}
}

// 轮流下注  轮流限时监听各个用户
func (room *Room) Bet() {
	index := room.BankerIndex
	nowBet := int64(0)
	for _, v := range room.PlayUserList {
		if v != nil && v.BetNowMoney > nowBet {
			nowBet = v.BetNowMoney
		}
	}
	for {
		Debug("房间 id = ", room.Id%ShowRoomId, " 用户座位=", index+1, " 下注")
		u := room.PlayUserList[index]

		if u != nil && u.Played {
			room.LeaveTimes = time.Second * 15 // 剩余等待时间
			// TODO 发送 当前所有用户下注 信息  nowBet当前下注
			room.SendDataToClient(index+1, false)
			st := time.Now()
			recving := true
			for recving {
				select {
				case buff, ok := <-u.session.ByteRecvChan:
					if ok {
						data := u.session.Format(buff)
						// TODO 处理从用户这边接收的信息
						Debug(data)
					}
				case <-time.After(room.LeaveTimes):
				}
				if recving {
					room.LeaveTimes -= time.Since(st) // 扣除已消耗时间
				}
			}
			room.LeaveTimes = 0
		}

		// 查找下一个下注人 并且判断是否下注完成
		over := false
		for i := 0; i < MaxPlayCount; i++ {
			index = (index + 1) % MaxPlayCount
			if room.PlayUserList[index] != nil && room.PlayUserList[index].Played && !room.PlayUserList[index].BetOk {
				over = true
				break
			}
		}
		if over { // 本轮下注完成
			break
		}
	}

	// 每一轮下注结束 重置下注相关信息
	for _, v := range room.PlayUserList {
		if v != nil {
			v.BetNowMoney = 0
			v.BetOk = false
		}
	}
}

// 发送信息给客户端 index-当前下注下标
func (room *Room) SendDataToClient(betSeat int, over bool) {
	// 发送 当前所有用户下注 信息  nowBet当前下注
	ri := &RoomInfo{
		BetSeat:     betSeat,
		LeaveTime:   int(room.LeaveTimes / time.Second),
		RoomStatus:  room.Status,
		CommonPoker: room.CommonPoker,
		MinBet:      room.MinBet,
	}

	for k, v := range room.PlayUserList {
		if v == nil {
			continue
		}
		ri.PlayUserList = []UserInfo{}
		for _, vv := range room.PlayUserList {
			if vv == nil {
				continue
			}
			ui := UserInfo{
				Money:       vv.Money,
				SeatNumber:  vv.SeatNumber,
				Played:      vv.Played,
				BetAllMoney: vv.BetAllMoney,
				BetNowMoney: vv.BetNowMoney,
				BetOk:       vv.BetOk,
			}
			if vv.id == v.id { // 当自己的时候发送
				ui.Poker = []int32{vv.Poker[0], vv.Poker[1]}
			}
			ri.PlayUserList = append(ri.PlayUserList, ui)
		}

		buff, err := json.Marshal(ri)
		if err != nil {
			Error(err)
			return
		}

		err = v.session.Send(&mySocket.FormatData{Id: 2001, Body: buff, Seq: int32(k + 1)})
		if err != nil {
			Error(err)
		}
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

// 参与用户数据初始化
func (room *Room) userDataInit() {
	for _, v := range room.PlayUserList {
		if v != nil {
			v.BetAllMoney = 0
			v.BetNowMoney = 0
			v.Poker = [2]int32{0, 0}
			v.Played = true
		}
	}
}

// 设置庄家，下 大盲注，小盲注
func (room *Room) SetBlind() {

	// 参与用户数据初始化
	room.userDataInit()
	room.MinBet = room.Chip

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

	Debug("房间 id = ", room.Id%ShowRoomId, " 大盲注 = ", bigBlindIndex, " 小盲注 = ", smallBlindIndex)

	if room.PlayUserList[bigBlindIndex].Money > room.Chip {
		room.PlayUserList[bigBlindIndex].Money -= room.Chip
		room.PlayUserList[bigBlindIndex].BetAllMoney = room.Chip
		room.PlayUserList[bigBlindIndex].BetNowMoney = room.Chip
	} else {
		room.PlayUserList[bigBlindIndex].BetAllMoney = room.PlayUserList[bigBlindIndex].Money
		room.PlayUserList[bigBlindIndex].BetNowMoney = room.PlayUserList[bigBlindIndex].Money
		room.PlayUserList[bigBlindIndex].Money = 0
	}

	if room.PlayUserList[smallBlindIndex].Money > room.Chip/2 {
		room.PlayUserList[smallBlindIndex].Money -= room.Chip / 2
		room.PlayUserList[smallBlindIndex].BetAllMoney = room.Chip / 2
		room.PlayUserList[smallBlindIndex].BetNowMoney = room.Chip / 2
	} else {
		room.PlayUserList[smallBlindIndex].BetAllMoney = room.PlayUserList[smallBlindIndex].Money
		room.PlayUserList[smallBlindIndex].BetNowMoney = room.PlayUserList[smallBlindIndex].Money
		room.PlayUserList[smallBlindIndex].Money = 0
	}

	room.Play = true
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
	Debug("JoinGame ", user.id%ShowRoomId)
	RoomMapMutex.RLock()
	defer RoomMapMutex.RUnlock()
	for k, v := range RoomMap {
		for kk, vv := range v.PlayUserList {
			if vv == nil {
				user.RoomId = k
				user.SeatNumber = int32(kk + 1)
				v.PlayUserList[kk] = user
				// TODO 用户进入房间信息 反馈给用户 和 其他所有用户
				return
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
	RoomMapMutex.Lock()
	defer RoomMapMutex.Unlock()
	for _, v := range RoomMap {
		for _, vv := range v.PlayUserList {
			if vv == nil {
				return 0 // 有空位置就退出
			}
		}
	}

	// 所有房间满 创建新房间

	id := time.Now().UnixNano()
	room := &Room{
		Id:   id,
		Chip: 10,
	}
	room.RoomInit()

	RoomMap[id] = room
	Notice("新房间 id = ", id, " 创建成功！")
	go room.GameLoop()
	return id
}

// 整个房间队列
func RoomLoop() {
	for {
		select {
		case newPlay := <-chanPlayGame:
			//			Debug(newPlay.id)
			go JoinGame(newPlay)
		}
	}
}

func DelRoom(Id int64) {
	RoomMapMutex.Lock()
	defer RoomMapMutex.Unlock()
	delete(RoomMap, Id)
}
