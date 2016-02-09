package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
  "github.com/larspensjo/config"
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
  serverConfig,err := getConfig("server");
  if(err != nil){
    log.Fatal("server config error:", err)
  }
	fmt.Println("listen on port "+serverConfig["port"])

	if err := http.ListenAndServe(":"+serverConfig["port"], nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func getConfig(sec string) (map[string]string,error){
  targetConfig := make(map[string]string)
  cfg, err := config.ReadDefault("config.ini")
	if err != nil {
		return targetConfig,Error("unable to open config file or wrong fomart")
	}
	sections := cfg.Sections()
	if len(sections) == 0 {
		return targetConfig,Error("no "+ sec +" config")
	}
  for _, section := range sections {
    if section != sec{
      continue
    }
    sectionData, _ := cfg.SectionOptions(section)
    for _, key := range sectionData {
      value, err := cfg.String(section, key)
      if err == nil {
        targetConfig[key] = value
      }
    }
    break
  }
  return targetConfig,nil
}
