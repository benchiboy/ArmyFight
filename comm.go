package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const PLAY_CARD = 1
const SIGN_IN = 2
const SEND_MSG = 3

/*
	发送消息命令
*/
type CommandMsg struct {
	Type    int    `json:"type"`
	Id      string `json:"id"`
	Message string `json:"message"`
	SCore   int    `json:"score"`
}

type CommandMsgResp struct {
	Type    int    `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

/*

 */
type Player struct {
	CurrConn   *websocket.Conn
	SignInTime time.Time
	LastCard   int
	CurrCard   int
	PlayTime   time.Time
}
