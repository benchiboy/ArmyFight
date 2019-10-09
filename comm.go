package main

import (
	"time"

	"github.com/gorilla/websocket"
)

//签到
const SIGN_IN = 1000
const SIGN_IN_RESP = 2000

//机器人签到
const ROBOT_SIGN_IN = 1066
const ROBOT_SIGN_IN_RESP = 2066

//出牌
const PLAY_CARD = 1001
const PLAY_CARD_RESP = 2001

//请求出牌
const REQ_PLAY_CARD = 1055
const REQ_PLAY_CARD_RESP = 2055

//查询结果
const QUERY_RESULT = 1012
const QUERY_RESULT_RESP = 2012

//发送消息
const SEND_MSG = 1003
const SEND_MSG_RESP = 2003

//发送语音
const SEND_VOICE = 1034
const SEND_VOICE_RESP = 2034

//查看在线用户
const GET_USERS = 1004
const GET_USERS_RESP = 2004

//请求玩家
const REQ_PLAY = 1005
const REQ_PLAY_RESP = 2005

//玩家同意
const REQ_PLAY_YES = 1006
const REQ_PLAY_YES_RESP = 2006

//玩家拒绝
const REQ_PLAY_NO = 1010
const REQ_PLAY_NO_RESP = 2010

//主动认输
const REQ_GIVEUP = 1007
const REQ_GIVEUP_RESP = 2007

//初始数据
const REQ_INIT_DATA = 1030
const REQ_INIT_DATA_RESP = 2030

//开始游戏
const START_GAME = 1035
const START_GAME_RESP = 2035

//改变用户
const CHANGE_USER = 1040
const CHANGE_USER_RESP = 2040

//下线通知
const OFFLINE_MSG = 1050
const OFFLINE_MSG_RESP = 2050

//签到类型
const ROBOT_TYPE = 1
const HUMAN_TYPE = 2

const STATUS_ONLIE_READY = 1
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
	Role     string `json:"role"`
	SCore    int    `json:"score"`
}

type CommandMsgResp struct {
	Type       int    `json:"type"`
	Success    bool   `json:"success"`
	Role       string `json:"role"`
	FromId     string `json:"fromid"`
	ToId       string `json:"toid"`
	Winner     string `json:"winner"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	AnotherMsg string `json:"anothermsg"`
}

// Status 1:在线空闲 2：在线游戏中 3:离线
type User struct {
	UserId     string    `json:"userid"`
	NickName   string    `json:"nickname"`
	Status     int       `json:"status"`
	PlayerType int       `json:"playertype"`
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
	CurrSCore  int
	PlayerType int
	CurrCard   string
	ToNickName string
	Status     int
	LoginTime  time.Time
	Avatar     string
	Memo       string
	Role       string
	Decoration int
	Candy      int
	Icecream   int
}
