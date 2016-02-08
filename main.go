package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	// "time"
)

var CurrentUsers *UserList //在线用户列表

type MessageReply struct {
	Type    string `json:"type"`
	Uname   string `json:"uname"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}

type UidCookieReply struct {
	Type string `json:"type"`
	Uid  string `json:"uid"`
}

type UserCountChangeReply struct {
	Type      string `json:"type"`
	UserCount int    `json:"user_count"`
}

type ServiceError struct {
	Msg string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s",e.Msg)
}

func Error(msg string) error {
	return &ServiceError{msg}
}

func main() {

	http.Handle("/", websocket.Handler(WsServer))

	fmt.Println("listen on port 8003")

	if err := http.ListenAndServe(":8003", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
