package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"sort"
	"strings"
)

type User struct {
	uid   string
	con   *websocket.Conn
	cards []int
}

type UserList []User

type Room struct {
	roomId    string
	userlist  []User
	host      string
	round     int
	score     map[string]int
	cardStack []int
	realcard  int
	falsecard map[string]int
	guess     map[string]int
	keyword   string
}

func (room *Room) New(ws *websocket.Conn, uid string) string {
	room.userlist = append(room.userlist, User{uid, ws, []int{}})
	fmt.Println("New user connect current user num", len(room.userlist))
	go room.PushUserCount("user_connect", uid)
	roomList[room.roomId] = *room
	return uid
}

func (room *Room) Remove(uid string) {
	flag, find := room.Exist(uid)
	fmt.Println("user disconnect uid: ", uid)
	if flag == true {
		room.userlist = append(room.userlist[:find], room.userlist[find+1:]...)
		go room.PushUserCount("user_disconnect", uid)
		roomList[room.roomId] = *room
	}
}

func (room *Room) ChangeConn(index int, con *websocket.Conn) {
	fmt.Println("visitor exist change connection")
	curUser := (room.userlist)[index]
	curUser.con.Close()
	(room.userlist)[index].con = con
	roomList[room.roomId] = *room
}

func (room *Room) Exist(uid string) (bool, int) {
	var find int
	flag := false
	for i, v := range room.userlist {
		if uid == v.uid {
			find = i
			flag = true
			break
		}
	}
	return flag, find
}

func (room *Room) PushUserCount(event string, uid string) {
	userlist := []string{}
	for _, user := range room.userlist {
		userlist = append(userlist, user.uid)
	}
	userCount := UserCountChangeReply{event, uid, len(room.userlist), strings.Join(userlist, ",")}
	replyBodyStr := JsonEncode(userCount)
	room.Broadcast(replyBodyStr)
}

func (room *Room) Broadcast(replyBodyStr string) error {
	fmt.Println("Broadcast ", room.roomId, " room user", len(room.userlist))
	for _, user := range room.userlist {
		if err := websocket.Message.Send(user.con, replyBodyStr); err != nil {
			fmt.Println("Can't send user ", user.uid, " lost connection")
			room.Remove(user.uid)
			break
		}
	}
	return nil
}

func (room *Room) Push(user User, replyBodyStr string) error {
	fmt.Println("Push ", room.roomId, user.uid)
	if err := websocket.Message.Send(user.con, replyBodyStr); err != nil {
		fmt.Println("Can't send user ", user.uid, " lost connection")
		room.Remove(user.uid)
	}
	return nil
}

func (room *Room) StartGame(creator string) {
	query("UPDATE `game` SET status = 1 WHERE `id`= " + room.roomId)
	room.host = creator
	room.score = make(map[string]int)
	room.round = 1
	var stack []int
	for index := 1; index < MAX_CARD_NUM+1; index++ {
		stack = append(stack, index)
	}
	room.cardStack = stack
	//抽牌
	for index, user := range room.userlist {
		cards, cardStack := drawCards(6, room.cardStack)
		room.cardStack = cardStack
		user.cards = cards
		room.userlist[index] = user
		replyBody := make(map[string]interface{})
		replyBody["type"] = "start"
		replyBody["cards"] = cards
		replyBody["host"] = room.host
		room.Push(user, JsonEncode(replyBody))
	}
}

func (room *Room) HostPick(keyword string, card int) {
	room.keyword = keyword
	room.realcard = card
	reply := make(map[string]string)
	reply["type"] = "hostpick"
	reply["keyword"] = keyword
	replyBody := JsonEncode(reply)
	room.Broadcast(replyBody)
}

func (room *Room) GuestPick(uid string, card int) {
	room.falsecard[uid] = card
	flag := true
	for _, u := range room.userlist {
		if u.uid != room.host {
			_, exist := room.falsecard[u.uid]
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
		room.Broadcast(JsonEncode(replyBody))
	}
}

func (room *Room) Guess(uid string, card int) {
	room.guess[uid] = card
	flag := true
	for _, u := range room.userlist {
		if u.uid != room.host {
			_, exist := room.guess[u.uid]
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
			for _, u := range room.userlist {
				replyBody["fillcard"] = fillCards[u.uid]
				replyBody["host"] = room.host
				replyBody["round"] = room.round
				room.Push(u, JsonEncode(replyBody))
			}
		} else {
			room.Broadcast(JsonEncode(replyBody))
		}
	}
}

func (room *Room) roundInit() map[string]int {
	fillCards := make(map[string]int)
	for _, u := range room.userlist {
		cards, cardStack := drawCards(1, room.cardStack)
		room.cardStack = cardStack
		fillCards[u.uid] = cards[0]
	}
	room.realcard = 0
	room.falsecard = make(map[string]int)
	room.guess = make(map[string]int)
	room.keyword = ""
	room.host = nextHost(room.host, room.userlist)
	return fillCards
}
