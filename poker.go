package main

import (
	. "github.com/soekchl/myUtils"
)

/*
	1~13	黑桃		Spade
	14~26	红桃		Heart
	27~39	方块		Diamond
	40~52	梅花		Club
*/

const (
	PokerLimit = 13 // 每个花色扑克 13张
)

var typeToName = make(map[int]string)

func init() {
	typeToName[1] = "高牌"
	typeToName[2] = "一对"
	typeToName[3] = "两对"
	typeToName[4] = "三条"
	typeToName[5] = "顺子"
	typeToName[6] = "同花"
	typeToName[7] = "葫芦"
	typeToName[8] = "四条"
	typeToName[9] = "同花顺"
}

/*
	输入
		手牌1 + 公共牌
		手牌2 + 公共牌
	输出
		0->平局 	1->1赢	-1->2赢
		最大的牌
*/
func ComparePoker(poker1, poker2 []int32) (result int, bestPoker []int32) {
	if len(poker1) != 7 || len(poker2) != 7 {
		return
	}

	pt1, result1 := getPokerType(poker1)
	pt2, result2 := getPokerType(poker2)
	if pt1 > pt2 {
		return 1, result1
	}
	if pt2 > pt1 {
		return -1, result2
	}
	for i := 0; i < 5; i++ {
		if result1[i]%PokerLimit > result2[i]%PokerLimit {
			return 1, result1
		}
		if result1[i]%PokerLimit < result2[i]%PokerLimit {
			return -1, result1
		}
	}
	return 0, result1
}

/*
	9	同花顺
	8	四条
	7	葫芦
	6	同花
	5	顺子
	4	三条
	3	两对
	2	一对
	1	高牌

	输入 - 公共牌+手牌(7张)
	输出 - 牌类型 + 最大五张牌 + 牌顺序
*/
func getPokerType(poker []int32) (poker_type int, resultPoker []int32) {
	if len(poker) != 7 {
		Error(poker)
		return
	}

	var pokerList [13]int // 各个统计
	var colorList [4]int  // 颜色个数

	sortPoker(poker) // 排序

	for i := 0; i < len(poker); i++ {
		pokerList[(poker[i]-1)%PokerLimit]++
		colorList[(poker[i]-1)/PokerLimit]++
	}

	colorHave := -1
	for i := 0; i < 4; i++ {
		if colorList[i] >= 5 {
			colorHave = i
			break
		}
	}

	// 同花顺
	var tmp []int32
	if colorHave != -1 {
		for i := 0; i < 7; i++ {
			if (poker[i]-1)/PokerLimit == int32(colorHave) {
				tmp = append(tmp, poker[i])
			}
		}
		if checkStraight(tmp[0:5]) {
			poker_type = 9
			resultPoker = tmp[0:5]
			return
		}
		if len(tmp) >= 6 && checkStraight(tmp[1:6]) {
			poker_type = 9
			resultPoker = tmp[1:6]
			return
		}
		if len(tmp) >= 7 && checkStraight(tmp[2:7]) {
			poker_type = 9
			resultPoker = tmp[2:7]
			return
		}
	}

	fourIndex := -1
	threeIndex1 := -1
	threeIndex2 := -1
	twoIndex1 := -1
	twoIndex2 := -1

	// 先找A
	for i := 0; i < 1; i++ {
		if pokerList[i] == 4 {
			fourIndex = i + 1
		}
		if pokerList[i] == 3 {
			if threeIndex1 == -1 {
				threeIndex1 = i + 1
			}
		}
		if pokerList[i] == 2 {
			if twoIndex1 == -1 {
				twoIndex1 = i + 1
			}
		}
	}

	// 2~K
	for i := 12; i >= 1; i-- {
		if pokerList[i] == 4 {
			fourIndex = i + 1
		}
		if pokerList[i] == 3 {
			if threeIndex1 == -1 {
				threeIndex1 = i + 1
			} else {
				threeIndex2 = i + 1
			}
		}
		if pokerList[i] == 2 {
			if twoIndex1 == -1 {
				twoIndex1 = i + 1
			} else if twoIndex2 == -1 {
				twoIndex2 = i + 1
			}
		}
	}

	// 四条
	if fourIndex >= 0 {
		resultPoker = []int32{int32(fourIndex), int32(fourIndex + PokerLimit),
			int32(fourIndex + PokerLimit*2), int32(fourIndex + PokerLimit*3)}
		poker_type = 8
		for i := 0; i < 7; i++ {
			if (poker[i]-1)%PokerLimit+1 != int32(fourIndex) {
				resultPoker = append(resultPoker, poker[i])
				return
			}
		}
	}

	// 葫芦
	if threeIndex1 >= 0 && (twoIndex1 >= 0 || threeIndex2 >= 0) {
		poker_type = 7
		for i := 0; i < 7; i++ { // 三条
			if (poker[i]-1)%PokerLimit+1 == int32(threeIndex1) {
				resultPoker = append(resultPoker, poker[i])
			}
		}
		if twoIndex1 >= 0 {
			for i := 0; i < 7; i++ { // 一对
				if (poker[i]-1)%PokerLimit+1 == int32(twoIndex1) {
					resultPoker = append(resultPoker, poker[i])
				}
			}
		} else {
			for i := 0; i < 7; i++ { // 另一个三条 的 一对
				if (poker[i]-1)%PokerLimit+1 == int32(threeIndex2) {
					resultPoker = append(resultPoker, poker[i])
				}
				if len(resultPoker) == 5 {
					break
				}
			}
		}
		return
	}

	// 同花
	if colorHave != -1 {
		poker_type = 6
		resultPoker = tmp[:5]
		return
	}

	{ // 顺子
		if checkStraight(poker[0:5]) {
			poker_type = 5
			resultPoker = poker[0:5]
			return
		}
		if checkStraight(poker[1:6]) {
			poker_type = 5
			resultPoker = poker[1:6]
			return
		}
		if checkStraight(poker[2:7]) {
			poker_type = 5
			resultPoker = poker[2:7]
			return
		}
	}

	// 三条
	if threeIndex1 >= 0 {
		poker_type = 4
		for i := 0; i < 7; i++ { // 三条
			if (poker[i]-1)%PokerLimit+1 == int32(threeIndex1) {
				resultPoker = append(resultPoker, poker[i])
			}
		}
		for i := 0; i < 7; i++ { // 剩余2张
			if (poker[i]-1)%PokerLimit+1 != int32(threeIndex1) {
				resultPoker = append(resultPoker, poker[i])
			}
			if len(resultPoker) == 5 {
				break
			}
		}
		return
	}

	// 两对
	if twoIndex1 >= 0 && twoIndex2 >= 0 {
		poker_type = 3
		for i := 0; i < 7; i++ { // 一对
			if (poker[i]-1)%PokerLimit+1 == int32(twoIndex1) {
				resultPoker = append(resultPoker, poker[i])
			}
		}
		for i := 0; i < 7; i++ { // 一对
			if (poker[i]-1)%PokerLimit+1 == int32(twoIndex2) {
				resultPoker = append(resultPoker, poker[i])
			}
		}
		for i := 0; i < 7; i++ { // 剩余1张
			if (poker[i]-1)%PokerLimit+1 != int32(twoIndex1) && (poker[i]-1)%PokerLimit+1 != int32(twoIndex2) {
				resultPoker = append(resultPoker, poker[i])
				break
			}
		}
		return
	}

	// 一对
	if twoIndex1 >= 0 {
		poker_type = 2
		for i := 0; i < 7; i++ { // 一对
			if (poker[i]-1)%PokerLimit+1 == int32(twoIndex1) {
				resultPoker = append(resultPoker, poker[i])
			}
		}
		for i := 0; i < 7; i++ { // 剩余3张
			if (poker[i]-1)%PokerLimit+1 != int32(twoIndex1) && (poker[i]-1)%PokerLimit+1 != int32(twoIndex2) {
				resultPoker = append(resultPoker, poker[i])
			}
			if len(resultPoker) == 5 {
				break
			}
		}
		return
	}

	// 高牌
	return 1, poker[:5]
}

// 判断是否为 顺子
func checkStraight(poker []int32) bool {
	if len(poker) != 5 {
		return false
	}

	sortPoker(poker)

	for i := 1; i < 5; i++ {
		if i == 1 && poker[i-1]%PokerLimit == 1 { // 最高的牌 为A的时候 第二张不是K 或 最后一张不是2 就不是顺子
			if poker[1]%PokerLimit != 0 && poker[4]%PokerLimit != 2 {
				return false
			}
		} else {
			if (poker[i-1]-poker[i])%PokerLimit != 1 {
				return false
			}
		}
	}
	return true
}

// 扑克牌排序 降序  A 最大
func sortPoker(poker []int32) {
	n := len(poker)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if poker[j]%PokerLimit == 1 {
				continue
			}
			if poker[j+1]%PokerLimit == 1 || (poker[j]-1)%PokerLimit < (poker[j+1]-1)%PokerLimit {
				poker[j], poker[j+1] = poker[j+1], poker[j]
			}
		}
	}
}

func showPokerNumber(poker []int32) (newPoker []int32) {
	for _, v := range poker {
		newPoker = append(newPoker, (v-1)%PokerLimit+1)
	}
	return
}
