package main

import (
	"math/rand"
	"time"
)

const MAX_CARD_NUM = 450

const WIN_SCORE = 30

//判定分数 补牌 开始下一轮
func gameResult(room *Room) (map[string]int, bool) {
	score := make(map[string]int)
	revertGuess := make(map[int]string)
	var wrongCardUid string
	var allGuessRight = true
	var gameover = false

	for uid, guess := range room.guess {
		revertGuess[guess] = uid
	}
	for uid, guess := range room.guess {
		//不能选自己的牌暂时在客户端判断
		if guess == room.realcard {
			score[uid]++
			score[room.host]++
		} else {
			wrongCardUid = revertGuess[guess]
			score[wrongCardUid]++
			allGuessRight = false
		}
	}
	if allGuessRight == true {
		score[room.host] = 0
	}
	totalScore := room.score
	for _, u := range room.userlist {
		totalScore[u.uid] = totalScore[u.uid] + score[u.uid]
		if totalScore[u.uid] > WIN_SCORE || totalScore[u.uid] == WIN_SCORE {
			gameover = true
		}
	}
	return totalScore, gameover
}

//TODO 考虑牌抓完的情况
//TODO 考虑重连
//抓牌
func drawCards(num int, stack []int) ([]int, []int) {
	rand.Seed(time.Now().Unix())
	getCards := []int{}
	for i := 0; i < num; i++ {
		index := rand.Intn(len(stack))
		getCards = append(getCards, stack[index])
		stack = append(stack[:index], stack[index+1:]...)
	}
	return getCards, stack
}

func nextHost(curHost string, ulist UserList) string {
	var index, next int
	var flag = false
	for i, u := range ulist {
		if curHost == u.uid {
			index = i
			flag = true
			break
		}
	}
	//玩家还存在
	if flag == true {
		if index == len(ulist)-1 {
			next = 0
		} else {
			next = index + 1
		}
	} else {
		next = 0
	}
	return ulist[next].uid
}
