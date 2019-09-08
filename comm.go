package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const SIGN_IN = 1
const PLAY_CARD = 2
const SEND_MSG = 3
const GET_USERS = 4
const REQ_PLAY = 5
const REQ_PALYOK = 6
const REQ_GIVEUP = 7
const REQ_GIVEUPOK = 8

const PUSH_PLAY_RET = 12

/*
	发送消息命令
*/
type CommandMsg struct {
	Type     int    `json:"type"`
	UserId   string `json:"userid"`
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

// Status 1:在线空闲  2：在线游戏中  3:离线
type User struct {
	Id        string    `json:"id"`
	NickName  string    `json:"nickname"`
	Status    int       `json:"status"`
	LoginTime time.Time `json:"login_time"`
}

/*

 */
type Player struct {
	CurrConn   *websocket.Conn
	SignInTime time.Time
	NickName   string
	LastCard   int
	CurrCard   int
}
