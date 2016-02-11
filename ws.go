package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"time"
)

type MessageReply struct {
	Type    string `json:"type"`
	Uid   string `json:"uid"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}

type UserCountChangeReply struct {
	Type      string `json:"type"`
  Uid  string `json:"uid"`
	UserCount int    `json:"user_count"`
}

func WsServer(ws *websocket.Conn){
	var err error
  uid := ws.Request().FormValue("uid")
	if uid == "" {
		fmt.Println("uid missing")
    return
	}
  roomId := ws.Request().FormValue("room_id")
  if roomId == "" {
		fmt.Println("roomId missing")
		return
	}
	room,exist := roomList[roomId]
	if exist == false{
		userlist := []User{}
		room = Room{roomId,userlist}
	}
	userExist, index := room.Exist(uid)
	if userExist == true {
		room.ChangeConn(index, ws)
	} else {
		fmt.Println("create new user")
		uid = room.New(ws, uid)
	}

	for {
		var receiveMsg string

		if err = websocket.Message.Receive(ws, &receiveMsg); err != nil {
			fmt.Println("Can't receive,user ", uid, " lost connection")
			room.Remove(uid)
			break
		}

		receiveNodes := JsonStrToStruct(receiveMsg)
		fmt.Println("Received back from client: ", receiveNodes)
		reply := MessageReply{Type: "message", Uid: receiveNodes["uid"].(string), Content: receiveNodes["content"].(string), Time: time.Now().Unix()}
		replyBody, err := json.Marshal(reply)
		if err != nil {
			panic(err.Error())
		}
		replyBodyStr := string(replyBody)
		go room.Broadcast(replyBodyStr)
	}
  // return nil
}
