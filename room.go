package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"strings"
)

type User struct {
	uid string
	con *websocket.Conn
}

type UserList []User

type Room struct {
	roomId   string
	userlist []User
	game Game
}

func (room *Room) New(ws *websocket.Conn, uid string) string {
	room.userlist = append(room.userlist, User{uid, ws})
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
	for _,user := range room.userlist{
		userlist = append(userlist,user.uid)
	}
	userCount := UserCountChangeReply{event, uid, len(room.userlist), strings.Join(userlist, ",")}
	replyBody, err := json.Marshal(userCount)
	if err != nil {
		panic(err.Error())
	}
	replyBodyStr := string(replyBody)
	room.Broadcast(replyBodyStr)
}

func (room *Room) Broadcast(replyBodyStr string) error {
	fmt.Println("current ",room.roomId," room user", len(room.userlist))
	for _, user := range room.userlist {
		if err := websocket.Message.Send(user.con, replyBodyStr); err != nil {
			fmt.Println("Can't send user ", user.uid, " lost connection")
			room.Remove(user.uid)
			break
		}
	}
	return nil
}
