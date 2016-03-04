package main

import (
	"github.com/ManiacMike/gwork"
)

func main() {
	gwork.Init(func(receiveNodes map[string]interface{}, uid string, room *gwork.Room){
		receiveType := receiveNodes["type"]
		rid := room.RoomId
		DixitRoomList = make(map[string]DixitRoom)
		gameRoom,ok := DixitRoomList[rid]
		if ok == false{
			gameRoom = DixitRoom{RoomId:rid,gworkRoom:room}
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
}
