package main

import (
	"fmt"
	"github.com/larspensjo/config"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	// "time"
)

var roomList map[string]Room //在线room列表

type ServiceError struct {
	Msg string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s", e.Msg)
}

func Error(msg string) error {
	return &ServiceError{msg}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	roomList = make(map[string]Room)
	http.Handle("/", websocket.Handler(WsServer))
	serverConfig, err := getConfig("server")
	if err != nil {
		log.Fatal("server config error:", err)
	}
	fmt.Println("listen on port " + serverConfig["port"])

	if err := http.ListenAndServe(":"+serverConfig["port"], nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func getConfig(sec string) (map[string]string, error) {
	targetConfig := make(map[string]string)
	cfg, err := config.ReadDefault("config.ini")
	if err != nil {
		return targetConfig, Error("unable to open config file or wrong fomart")
	}
	sections := cfg.Sections()
	if len(sections) == 0 {
		return targetConfig, Error("no " + sec + " config")
	}
	for _, section := range sections {
		if section != sec {
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
	return targetConfig, nil
}
