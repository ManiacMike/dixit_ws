package main

import (
	"database/sql"
	"fmt"
	"github.com/ManiacMike/gwork"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

var dixitMysqlDsn string

func main() {
	dbConfig, err := gwork.GetConfig("config.ini", "mysql")
	if err != nil {
		log.Fatal("mysql config error:", err)
		os.Exit(-1)
	} else {
		dixitMysqlDsn = dbConfig["dixit"] + "?charset=utf8"
	}

	gwork.SetGetConnCallback(func(uid string, room *gwork.Room) {
		ulist := []string{}
		for _, u := range room.Userlist {
			ulist = append(ulist, u.Uid)
		}
		reply := map[string]interface{}{
			"type":       "user_connect",
			"uid":        uid,
			"user_count": len(room.Userlist),
			"user_list":  strings.Join(ulist, ","),
		}
		room.Broadcast(reply)
		// updateMysqlUserlist(ulist,room.RoomId)
	})

	gwork.SetLoseConnCallback(func(uid string, room *gwork.Room) {
		ulist := []string{}
		for _, u := range room.Userlist {
			ulist = append(ulist, u.Uid)
		}
		reply := map[string]interface{}{
			"type":       "user_disconnect",
			"uid":        uid,
			"user_count": len(room.Userlist),
			"user_list":  strings.Join(ulist, ","),
		}
		room.Broadcast(reply)
		// updateMysqlUserlist(ulist,room.RoomId)
	})

	gwork.SetRequestHandler(func(receiveNodes map[string]interface{}, uid string, room *gwork.Room) {
		receiveType := receiveNodes["type"]
		rid := room.RoomId
		DixitRoomList = make(map[string]*DixitRoom)
		gameRoom, ok := DixitRoomList[rid]
		if ok == false {
			gameRoom = &DixitRoom{RoomId: rid, gworkRoom: room}
		}
		if receiveType == "start" {
			gameRoom.StartGame(uid)
		} else if receiveType == "hostpick" {
			gameRoom.HostPick(receiveNodes["keyword"].(string), receiveNodes["card"].(int))
		} else if receiveType == "guestpick" {
			gameRoom.GuestPick(uid, receiveNodes["card"].(int))
		} else if receiveType == "guess" {
			gameRoom.Guess(uid, receiveNodes["card"].(int))
		}
	})

	gwork.Start()
}

func updateMysqlUserlist(ulist []string, rid string) bool {
	db, err := sql.Open("mysql", dixitMysqlDsn)
	if err != nil {
		panic(err)
		return false
	}
	defer db.Close()
	ustr := strings.Join(ulist, ",")
	fmt.Println("UPDATE `game` SET user_list= ? WHERE `id`= ?", ustr, rid)
	db.Query("UPDATE `game` SET user_list= ? WHERE `id`= ?", ustr, rid)
	return true
}
