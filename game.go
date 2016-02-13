package main

import ()

type Game struct {
	roomId string
	host   string
	round  int
	stage  int
	score  map[string]int
}
