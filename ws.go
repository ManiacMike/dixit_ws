package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"time"
)

func WsServer(ws *websocket.Conn){
	var err error
	if nil == CurrentUsers {
		CurrentUsers = new(UserList)
	}
  uid := ws.Request().PostFormValue("uid")
	if uid == "" {
    // return Error("uid missing")
	}
  roomId := ws.Request().PostFormValue("room_id")
  if roomId == "" {
    // return Error("roomId missing")
	}
	userExist, index := CurrentUsers.Exist(uid)
	if userExist == true {
		CurrentUsers.ChangeConn(index, ws)
	} else {
		fmt.Println("create new user")
		uid = CurrentUsers.New(ws, uid, roomId)
	}
	go PushUserCount()

	for {
		var receiveMsg string

		if err = websocket.Message.Receive(ws, &receiveMsg); err != nil {
			fmt.Println("Can't receive,user ", uid, " lost connection")
			CurrentUsers.Remove(uid)
			break
		}

		receiveNodes := JsonStrToStruct(receiveMsg)
		fmt.Println("Received back from client: ", receiveNodes)
		reply := MessageReply{Type: "message", Uname: receiveNodes["uname"].(string), Content: receiveNodes["content"].(string), Time: time.Now().Unix()}
		replyBody, err := json.Marshal(reply)
		if err != nil {
			panic(err.Error())
		}
		replyBodyStr := string(replyBody)
		go Broadcast(replyBodyStr)
	}
  // return nil
}

func PushUserCount() {
	userCount := UserCountChangeReply{"user_count", len(*CurrentUsers)}
	replyBody, err := json.Marshal(userCount)
	if err != nil {
		panic(err.Error())
	}
	replyBodyStr := string(replyBody)
	Broadcast(replyBodyStr)
}

func Broadcast(replyBodyStr string) error {
	fmt.Println("current user", len(*CurrentUsers))
	for _, user := range *CurrentUsers {
		if err := websocket.Message.Send(user.con, replyBodyStr); err != nil {
			fmt.Println("Can't send user ", user.uid, " lost connection")
			CurrentUsers.Remove(user.uid)
			break
		}
	}
	return nil
}
