package main

import ()

type Game struct {
	roomId string
	host   string
	round  int
	stage  int
	score  map[string]int
}

const MAX_CARD_NUM = 450

//第一次发牌
func (game *Game) releaseCards(){

}

//收到host选择的牌和关键字
func (game *Game) pickHostCard(){

}

//其他玩家选择混淆牌
func (game *Game) pickGuestCard(){

}

//选择卡片
func (game *Game) guessCard(){

}

//补牌
func (game *Game) _drawCards(){

}
