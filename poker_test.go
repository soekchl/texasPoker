package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	//	testSortPoker(t)
	//	testCheckStraight(t)
	//	testGetPokerType(t)
	testComparePoker(t)
}

func testComparePoker(t *testing.T) {
	check := func() {
		poker1 := []int32{}
		poker2 := []int32{}
		for i := 0; i < 2; i++ {
			poker1 = append(poker1, rand.Int31n(52)+1)
			poker2 = append(poker2, rand.Int31n(52)+1)
		}

		for i := 0; i < 5; i++ {
			r := rand.Int31n(52) + 1
			poker1 = append(poker1, r)
			poker2 = append(poker2, r)
		}
		t.Log("-----------------------")
		rr, _ := ComparePoker(poker1, poker2)
		switch rr {
		case 1:
			t.Log("1赢")
		case -1:
			t.Log("1输")
		default:
			t.Log("平")
			t.Log(poker1)
			t.Log(poker2)
		}

		ty, np := getPokerType(poker1)
		t.Log(typeToName[ty])
		showPoker(t, np)

		ty, np = getPokerType(poker2)
		t.Log(typeToName[ty])
		showPoker(t, np)
	}

	for i := 0; i < 10; i++ {
		check()
	}
}

func testSortPoker(t *testing.T) {
	for i := 0; i < 3; i++ {
		var poker []int32
		for i := 0; i < 10; i++ {
			poker = append(poker, rand.Int31n(52)+1)
		}

		t.Log(poker)
		sortPoker(poker)
		t.Log("\t", poker)
		showPoker(t, poker)
	}
}

func testCheckStraight(t *testing.T) {
	poker := []int32{1, 2, 4, 5}
	t.Log(poker, checkStraight(poker))

	poker = []int32{1, 2, 3, 4, 5}
	t.Log(poker, checkStraight(poker))

	poker = []int32{6, 2, 3, 4, 5}
	t.Log(poker, checkStraight(poker))

	poker = []int32{1, 11, 12, 13, 10}
	t.Log(poker, checkStraight(poker))

	check := func() {
		poker := []int32{}
		for i := 0; i < 5; i++ {
			poker = append(poker, rand.Int31n(52)+1)
		}

		result := checkStraight(poker)
		if result {
			t.Log(poker, result)
			showPoker(t, poker)
		}
	}

	for i := 0; i < 1000; i++ {
		check()
	}
}

func testGetPokerType(t *testing.T) {
	check := func() {
		poker := []int32{}
		for i := 0; i < 7; i++ {
			poker = append(poker, rand.Int31n(52)+1)
		}
		t.Log("-----------------------")
		t.Log(poker)
		showPoker(t, poker)

		ty, np := getPokerType(poker)
		t.Log(typeToName[ty])
		showPoker(t, np)
	}

	for i := 0; i < 10; i++ {
		check()
	}
}

func showPoker(t *testing.T, poker []int32) {
	str := ""
	for i := 0; i < len(poker); i++ {
		str = fmt.Sprintf("%s%-4v", str, (poker[i]-1)%PokerLimit+1)
	}
	t.Log(str)
}
