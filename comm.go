package main

import (
	"time"

	"github.com/gorilla/websocket"
)

//签到
const (
	SIGN_IN      = 1000
	SIGN_IN_RESP = 2000

	//机器人签到
	ROBOT_SIGN_IN      = 1066
	ROBOT_SIGN_IN_RESP = 2066

	//出牌
	PLAY_CARD      = 1001
	PLAY_CARD_RESP = 2001

	//请求出牌
	REQ_PLAY_CARD      = 1055
	REQ_PLAY_CARD_RESP = 2055

	//查询结果
	QUERY_RESULT      = 1012
	QUERY_RESULT_RESP = 2012

	//发送消息
	SEND_MSG      = 1003
	SEND_MSG_RESP = 2003

	//发送语音
	SEND_VOICE      = 1034
	SEND_VOICE_RESP = 2034

	//查看在线用户
	GET_USERS      = 1004
	GET_USERS_RESP = 2004

	//查看在线机器人
	GET_ROBOTS      = 1080
	GET_ROBOTS_RESP = 2080

	//请求玩家
	REQ_PLAY      = 1005
	REQ_PLAY_RESP = 2005

	//玩家同意
	REQ_PLAY_YES      = 1006
	REQ_PLAY_YES_RESP = 2006

	//玩家拒绝
	REQ_PLAY_NO      = 1010
	REQ_PLAY_NO_RESP = 2010

	//主动认输
	REQ_GIVEUP      = 1007
	REQ_GIVEUP_RESP = 2007

	//初始数据
	REQ_INIT_DATA      = 1030
	REQ_INIT_DATA_RESP = 2030

	//开始游戏
	START_GAME      = 1035
	START_GAME_RESP = 2035

	//改变用户
	CHANGE_USER      = 1040
	CHANGE_USER_RESP = 2040

	//下线通知
	OFFLINE_MSG      = 1050
	OFFLINE_MSG_RESP = 2050
)

//签到类型
const (
	ROBOT_TYPE = 1
	HUMAN_TYPE = 2

	STATUS_ONLIN_IDLE = 1
	STATUS_ONLIE_DONG = 2
	STATUS_OFFLINE    = 3
)

const SYSTEM_NAME = "MyBoss"

/*
	发送消息命令
*/
type CardInfo struct {
	Count int    `json:"count"`
	SCore int    `json:"score"`
	Name  string `json:"name"`
}

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

/*
UserId     string    `json:"userid"`
	NickName   string    `json:"nickname"`
	Status     int       `json:"status"`
	PlayerType int       `json:"playertype"`
	Avatar     string    `json:"avatar"`
	Memo       string    `json:"memo"`
	LoginTime  time.Time `json:"logintime"`
	Decoration int       `json:"decoration"`
	Candy      int       `json:"candy"`
	Icecream   int       `json:"icecrea"`
*/
/*

 */
type Player struct {
	CurrConn   *websocket.Conn
	SignInTime time.Time `json:"logintime"`
	NickName   string    `json:"nickname"`
	CurrSCore  int       `json:"currscore"`
	PlayerType int       `json:"playertype"`
	CurrCard   string    `json:"currcard"`
	ToNickName string    `json:"tonickname"`
	Status     int       `json:"status"`
	Avatar     string    `json:"avatar"`
	Role       string    `json:"role"`
	Decoration int       `json:"decoration"`
}

/* 用户注册*/
type SignUp struct {
	UserName  string `json:"username"`
	Pwd       string `json:"pwd"`
	Problem   string `json:"problem"`
	Answer    string `json:"answer"`
	HeadImage string `json:"headimage"`
}

type SignUpResp struct {
	ResultCode string `json:"retcode"`
	ResultMsg  string `json:"retmsg"`
}

/*找回密码*/
type ResetPwd struct {
	UserName string `json:"username"`
	Problem  string `json:"problem"`
	Answer   string `json:"answer"`
	NewPwd   string `json:"newpwd"`
}

type ResetPwdResp struct {
	ResultCode string `json:"retcode"`
	ResultMsg  string `json:"retmsg"`
}

/*得到用户信息*/
type GetUser struct {
	UserName string `json:"username"`
}

type GetUserResp struct {
	ResultCode string `json:"retcode"`
	ResultMsg  string `json:"retmsg"`
	UserName   string `json:"username"`
	Problem    string `json:"problem"`
	Answer     string `json:"answer"`
	HeadImage  string `json:"headimage"`
}
