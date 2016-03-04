package main

import (
	"math/rand"
	"time"
	"github.com/ManiacMike/gwork"
	"sort"
)

const MAX_CARD_NUM = 450

const WIN_SCORE = 30

type DixitRoom struct {
		RoomId	  string
		host      string
		round     int
		score     map[string]int
		cardStack []int
		realcard  int
		falsecard map[string]int
		guess     map[string]int
		keyword   string
		gworkRoom *gwork.Room
}

var DixitRoomList map[string]DixitRoom

func (room *DixitRoom) StartGame(creator string) {
	gwork.MysqlQuery("UPDATE `game` SET status = 1 WHERE `id`= " + room.RoomId)
	room.host = creator
	room.score = make(map[string]int)
	room.round = 1
	room.realcard = 0
	room.falsecard = make(map[string]int)
	room.guess = make(map[string]int)
	room.keyword = ""
	var stack []int
	for index := 1; index < MAX_CARD_NUM+1; index++ {
		stack = append(stack, index)
	}
	room.cardStack = stack
	//抽牌
	for _, user := range room.gworkRoom.Userlist {
		cards, cardStack := drawCards(6, room.cardStack)
		room.cardStack = cardStack
		//user.cards = cards
		// room.Userlist[index] = user
		replyBody := make(map[string]interface{})
		replyBody["type"] = "start"
		replyBody["cards"] = cards
		replyBody["host"] = room.host
		room.gworkRoom.Push(user, gwork.JsonEncode(replyBody))
	}
	DixitRoomList[room.RoomId] = *room
}

func (room *DixitRoom) HostPick(keyword string, card int) {
	room.keyword = keyword
	room.realcard = card
	reply := make(map[string]string)
	reply["type"] = "hostpick"
	reply["keyword"] = keyword
	replyBody := gwork.JsonEncode(reply)
	room.gworkRoom.Broadcast(replyBody)
	DixitRoomList[room.RoomId] = *room
}

func (room *DixitRoom) GuestPick(uid string, card int) {
	room.falsecard[uid] = card
	flag := true
	for _, u := range room.gworkRoom.Userlist {
		if u.Uid != room.host {
			_, exist := room.falsecard[u.Uid]
			if exist == false {
				flag = false
			}
		}
	}
	if flag == true {
		cards := []int{}
		for _, c := range room.falsecard {
			cards = append(cards, c)
		}
		cards = append(cards, room.realcard)
		sort.Ints(cards)
		replyBody := make(map[string]interface{})
		replyBody["type"] = "showcards"
		replyBody["cards"] = cards
		room.gworkRoom.Broadcast(gwork.JsonEncode(replyBody))
	}
	DixitRoomList[room.RoomId] = *room
}

func (room *DixitRoom) Guess(uid string, card int) {
	room.guess[uid] = card
	flag := true
	for _, u := range room.gworkRoom.Userlist {
		if u.Uid != room.host {
			_, exist := room.guess[u.Uid]
			if exist == false {
				flag = false
			}
		}
	}
	if flag == true {
		score, gameover := gameResult(room)
		room.score = score
		cards := room.falsecard
		cards[room.host] = room.realcard
		replyBody := make(map[string]interface{})
		replyBody["type"] = "result"
		replyBody["gameover"] = gameover
		replyBody["cards"] = cards
		replyBody["score"] = score
		if gameover == false {
			fillCards := room.roundInit()
			replyBody["round"] = room.round
			replyBody["host"] = room.host
			for _, u := range room.gworkRoom.Userlist {
				replyBody["fillcard"] = fillCards[u.Uid]
				room.gworkRoom.Push(u, gwork.JsonEncode(replyBody))
			}
		} else {
			room.gworkRoom.Broadcast(gwork.JsonEncode(replyBody))
		}
	}
	DixitRoomList[room.RoomId] = *room
}

func (room *DixitRoom) roundInit() map[string]int {
	fillCards := make(map[string]int)
	for _, u := range room.gworkRoom.Userlist {
		cards, cardStack := drawCards(1, room.cardStack)
		room.cardStack = cardStack
		fillCards[u.Uid] = cards[0]
	}
	room.realcard = 0
	room.falsecard = make(map[string]int)
	room.guess = make(map[string]int)
	room.keyword = ""
	room.round++
	room.host = nextHost(room.host, room.gworkRoom.Userlist)
	DixitRoomList[room.RoomId] = *room
	return fillCards
}

//判定分数 补牌 开始下一轮
func gameResult(room *DixitRoom) (map[string]int, bool) {
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
	for _, u := range room.gworkRoom.Userlist {
		totalScore[u.Uid] = totalScore[u.Uid] + score[u.Uid]
		if totalScore[u.Uid] > WIN_SCORE || totalScore[u.Uid] == WIN_SCORE {
			gameover = true
		}
	}
	return totalScore, gameover
}

//TODO 考虑重连
//抓牌
func drawCards(num int, stack []int) ([]int, []int) {
	rand.Seed(time.Now().Unix())
	getCards := []int{}
	for i := 0; i < num; i++ {
		if len(stack) < 10 {
			for index := 1; index < MAX_CARD_NUM+1; index++ {
				stack = append(stack, index)
			}
		}
		index := rand.Intn(len(stack))
		getCards = append(getCards, stack[index])
		stack = append(stack[:index], stack[index+1:]...)
	}
	return getCards, stack
}

func nextHost(curHost string, ulist gwork.UserList) string {
	var index, next int
	var flag = false
	for i, u := range ulist {
		if curHost == u.Uid {
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
	return ulist[next].Uid
}
