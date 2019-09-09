package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const SIGN_IN = 1
const PLAY_CARD = 2
const PLAY_CARD_RESULT = 9
const SEND_MSG = 3
const GET_USERS = 4
const REQ_PLAY = 5
const REQ_PALYOK = 6
const REQ_GIVEUP = 7
const REQ_GIVEUPOK = 8

const STATUS_ONLIE_DONG = 1
const STATUS_ONLIN_IDLE = 2
const STATUS_OFFLINE = 3

/*
	发送消息命令
*/
type CommandMsg struct {
	Type     int    `json:"type"`
	FromId   string `json:"fromid"`
	ToId     string `json:"toid"`
	NickName string `json:"nickname"`
	Message  string `json:"message"`
	SCore    int    `json:"score"`
}

type CommandMsgResp struct {
	Type    int    `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Status 1:在线空闲 2：在线游戏中 3:离线
type User struct {
	UserId     string    `json:"userid"`
	NickName   string    `json:"nickname"`
	Status     int       `json:"status"`
	Avatar     string    `json:"avatar"`
	Memo       string    `json:"memo"`
	LoginTime  time.Time `json:"logintime"`
	Decoration int       `json:"decoration"`
	Candy      int       `json:"candy"`
	Icecream   int       `json:"icecream"`
}

/*

 */
type Player struct {
	CurrConn   *websocket.Conn
	SignInTime time.Time
	NickName   string
	CurrCard   int
	Status     int
	LoginTime  time.Time
	Avatar     string
	Memo       string
	Decoration int
	Candy      int
	Icecream   int
}
