// ArmFight project ArmFight.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"ArmFight/afuser"
	"ArmFight/dbutil"

	goconf "github.com/pantsing/goconf"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/gorilla/websocket"
)

var (
	GId2ConnMap = &sync.Map{}
	GId2IdMap   = &sync.Map{}
	GConn2IdMap = &sync.Map{}
	dbUrl       string
	ccdbUrl     string
	listenPort  int
	idleConns   int
	openConns   int
)

var addr = flag.String("addr", "172.17.0.3:9080", "http service address")

//var addr = flag.String("addr", "127.0.0.1:9080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

/*
	玩家签到处理
	1:从数据库查询用户的信息 同步到内存中
*/
func signIn(c *websocket.Conn, playerType int, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SignIn============>")
	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Type = SIGN_IN_RESP
	cmdMsgResp.Success = true
	_, ok := GId2ConnMap.Load(cmdMsg.NickName)
	if ok {
		log.Println(cmdMsg.NickName + "用户已经在线")
		cmdMsgResp.Success = false
		cmdMsgResp.Message = "用户已经在线"
		return cmdMsgResp
	}
	if playerType == HUMAN_TYPE {
		user := afuser.New(dbutil.GetDB(), afuser.DEBUG)
		var search afuser.Search
		search.UserName = cmdMsg.NickName
		if e, err := user.Get(search); err != nil {
			cmdMsgResp.Success = false
			cmdMsgResp.Message = "登录账号或密码错误"
			return cmdMsgResp
		} else {
			fmt.Println(e.UserPwd, cmdMsg.Message)
			if e.UserPwd != cmdMsg.Message {
				cmdMsgResp.Success = false
				cmdMsgResp.Message = "登录账号或密码错误"
				return cmdMsgResp
			}
		}
	}
	GId2ConnMap.Store(cmdMsg.NickName, Player{CurrConn: c, SignInTime: time.Now(),
		NickName: cmdMsg.NickName, Status: STATUS_ONLIN_IDLE, PlayerType: playerType})
	GConn2IdMap.Store(c, cmdMsg.NickName)

	cmdMsgResp.FromId = cmdMsg.FromId
	cmdMsgResp.ToId = cmdMsg.FromId
	cmdMsgResp.Message = "登录成功"
	return cmdMsgResp
}

/*
	玩家发牌处理
*/

func playCard(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>PlayCard============>")

	cmdMsg.Role = getRole(cmdMsg.FromId)
	proxyMsg(c, cmdMsg)

	setCard(cmdMsg.FromId, cmdMsg.SCore, cmdMsg.Message)

	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Message = cmdMsg.Message
	cmdMsgResp.Role = getRole(cmdMsg.FromId)
	cmdMsgResp.Type = PLAY_CARD_RESP

	return cmdMsgResp
}

/*
	玩家发牌处理
*/

func reqPlayCard(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayCard============>")

	proxyMsg(c, cmdMsg)

	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Message = cmdMsg.Message
	cmdMsgResp.Type = REQ_PLAY_CARD_RESP

	return cmdMsgResp
}

/*
	玩家发牌对比结果
*/

func queryResult(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>queryResult============>")
	var mScore, sScore int
	var sCard, mCard string
	if getPlayerType(cmdMsg.ToId) == ROBOT_TYPE {
		mScore, mCard = getCard(cmdMsg.FromId)
		sScore, sCard = getCard(cmdMsg.ToId)
	} else {
		sScore, sCard = getCard(cmdMsg.FromId)
		mScore, mCard = getCard(cmdMsg.ToId)
	}
	var cmdMsgResp CommandMsgResp
	if sScore > mScore {
		//出现炸弹，地址 和人员相碰的情况
		if sScore == 101 || sScore == 100 {
			if mScore == 0 {
				cmdMsgResp.Winner = "M"
			} else {
				cmdMsgResp.Winner = "B"
			}
		} else {
			cmdMsgResp.Winner = "S"
			if mScore == 0 {
				cmdMsgResp.Status = "E"
			}
		}
	} else if sScore == mScore {
		cmdMsgResp.Winner = "B"
	} else {
		//出现炸弹，地雷 和人员相碰的情况
		if mScore == 101 || mScore == 100 {
			if sScore == 0 {
				cmdMsgResp.Winner = "S"
			} else {
				cmdMsgResp.Winner = "B"
			}

		} else {
			cmdMsgResp.Winner = "M"
			if sScore == 0 {
				cmdMsgResp.Status = "E"
			}
		}
	}
	//出工兵和地理的情况
	if sScore == 1 && mScore == 100 {
		cmdMsgResp.Winner = "S"
	}
	if mScore == 1 && sScore == 100 {
		cmdMsgResp.Winner = "M"
	}
	//出炸弹和地雷的情况
	if sScore == 100 && mScore == 101 {
		cmdMsgResp.Winner = "B"
	}
	if mScore == 100 && sScore == 101 {
		cmdMsgResp.Winner = "B"
	}

	cmdMsgResp.Type = QUERY_RESULT_RESP
	cmdMsgResp.Message = mCard
	cmdMsgResp.AnotherMsg = sCard
	cmdMsgResp.ToId = cmdMsg.ToId
	cmdMsgResp.FromId = cmdMsg.FromId
	fmt.Println("======>", mCard, sCard)

	if getPlayerType(cmdMsg.ToId) == ROBOT_TYPE {
		cmdMsgResp.Role = "S"
	} else {
		cmdMsgResp.Role = "M"
	}
	proxyMsgResp(cmdMsg.ToId, cmdMsgResp)

	cmdMsgResp.Role = "M"

	return cmdMsgResp
}

/*
	玩家发送消息处理
*/
func sendMsg(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SendMsg============>")

	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = SEND_MSG_RESP
	return cmdMsgResp
}

/*
	玩家发送语音
*/
func sendVoice(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SendVoice============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = SEND_VOICE_RESP
	return cmdMsgResp
}

/*
	功能：得到人类对手类别
*/
func getUsers(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>getUsers============>")
	uList := make([]Player, 0)
	var cmdMsgResp CommandMsgResp
	GId2ConnMap.Range(func(k, v interface{}) bool {

		p, _ := v.(Player)
		fmt.Println(p)
		if cmdMsg.NickName != p.NickName && p.PlayerType == HUMAN_TYPE {
			uList = append(uList, p)
		}
		return true
	})
	userBuf, _ := json.Marshal(uList)
	cmdMsgResp.Message = string(userBuf)
	cmdMsgResp.Success = true
	cmdMsgResp.Type = GET_USERS_RESP
	return cmdMsgResp
}

/*
	功能：得到机器对手列表
*/
func getRobots(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>getRobots============>")
	uList := make([]Player, 0)
	var cmdMsgResp CommandMsgResp
	GId2ConnMap.Range(func(k, v interface{}) bool {
		p, _ := v.(Player)
		if cmdMsg.NickName != p.NickName && p.PlayerType == ROBOT_TYPE {
			log.Println("====>", p)
			uList = append(uList, p)
		}
		return true
	})
	userBuf, _ := json.Marshal(uList)
	cmdMsgResp.Message = string(userBuf)
	cmdMsgResp.Success = true
	cmdMsgResp.Type = GET_ROBOTS_RESP
	return cmdMsgResp
}

/*
	消息转发
*/
func proxyMsg(c *websocket.Conn, cmdMsg CommandMsg) {
	log.Println("=========>proxyMsg============>")
	log.Println(cmdMsg.FromId, cmdMsg.ToId)
	playerObj, ok := GId2ConnMap.Load(cmdMsg.ToId)
	if !ok {
		log.Println(cmdMsg.ToId + "缓存信息没有获取到")
		return
	}
	toPlayer, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return
	}
	if reqBuf, err := json.Marshal(cmdMsg); err != nil {
		log.Println(err)
	} else {
		if err := toPlayer.CurrConn.WriteMessage(websocket.TextMessage, []byte(string(reqBuf))); err != nil {
			log.Println("发送出错")
		}
	}
}

/*
	消息转发
*/
func proxyMsgResp(toId string, cmdMsgResp CommandMsgResp) {
	log.Println("=========>proxyMsgResp============>")
	playerObj, ok := GId2ConnMap.Load(toId)
	if !ok {
		log.Println("缓存信息没有获取到")
	}
	toPlayer, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	if reqBuf, err := json.Marshal(cmdMsgResp); err != nil {
		log.Println(err)
	} else {
		if err := toPlayer.CurrConn.WriteMessage(websocket.TextMessage, []byte(string(reqBuf))); err != nil {
			log.Println("发送出错")
		}
	}
}

/*
	改变用户通知
*/
func changeUser(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>changeUser============>", cmdMsg.FromId, cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp

	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = CHANGE_USER_RESP
	return cmdMsgResp
}

/*
	请求玩家
*/
func reqPlay(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlay============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = REQ_PLAY_RESP
	return cmdMsgResp
}

/*
	开始游戏
*/
func startGame(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>startGame============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)

	cmdMsgResp.Type = START_GAME_RESP
	return cmdMsgResp
}

/*
	另一玩家答应请求
*/
func setRole(playerId string, role string) {
	log.Println("=========>setRole============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.Role = role
	GId2ConnMap.Store(playerId, player)
	//log.Println("=========>setRole============>", playerId)
}

/*
	另一玩家答应请求
*/
func getRole(playerId string) string {
	log.Println("=========>setRole============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	return player.Role
}

/*
	另一玩家答应请求
*/
func getPlayerType(playerId string) int {
	log.Println("=========>setRole============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	return player.PlayerType
}

/*
	存储玩家的出牌
*/
func setCard(playerId string, score int, card string) {
	log.Println("=========>setCard============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.CurrSCore = score
	player.CurrCard = card
	GId2ConnMap.Store(playerId, player)
	log.Println("=========>setCard============>", playerId)
}

/*
	得到
*/
func getCard(playerId string) (int, string) {
	log.Println("=========>getCard============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	return player.CurrSCore, player.CurrCard
}

/*
	另一玩家答应请求
*/
func setStatus(playerId string, status int) {
	//log.Println("=========>setStatus============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.Status = status

	GId2ConnMap.Store(playerId, player)
	//log.Println("=========>setStatus============>", playerId)
}

/*
	存储另一玩家的信息
*/
func setToNickName(playerId string, toNickName string) {
	//log.Println("=========>setStatus============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.ToNickName = toNickName

	GId2ConnMap.Store(playerId, player)
	//log.Println("=========>setStatus============>", playerId)
}

/*
	另一玩家答应请求
*/
func reqPlayYes(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayYes============>")
	var cmdMsgResp CommandMsgResp
	cmdMsg.Message = "对方同意了"

	proxyMsg(c, cmdMsg)

	//初始玩家双方的状态
	setStatus(cmdMsg.FromId, STATUS_ONLIE_READY)
	setStatus(cmdMsg.ToId, STATUS_ONLIE_READY)

	setToNickName(cmdMsg.FromId, cmdMsg.ToId)
	setToNickName(cmdMsg.ToId, cmdMsg.FromId)

	setRole(cmdMsg.FromId, "S")
	setRole(cmdMsg.ToId, "M")

	fmt.Println(cmdMsg.FromId+"角色：", getRole(cmdMsg.FromId))
	fmt.Println(cmdMsg.ToId+"角色：", getRole(cmdMsg.ToId))

	cmdMsgResp.Type = REQ_PLAY_YES_RESP
	return cmdMsgResp
}

/*
	另一玩家拒绝请求
*/
func reqPlayNo(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayNo============>")
	var cmdMsgResp CommandMsgResp
	cmdMsg.Message = "对方拒绝了"
	proxyMsg(c, cmdMsg)

	cmdMsgResp.Type = REQ_PLAY_NO_RESP
	return cmdMsgResp
}

/*
	登录成功后初始化数据
*/
func initData(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>initData============>")
	var cmdMsgResp CommandMsgResp
	cardMap := map[string]CardInfo{
		"gongbing":  CardInfo{Count: 3, SCore: 1, Name: "工兵"},
		"paizhang":  CardInfo{Count: 2, SCore: 2, Name: "排长"},
		"lianzhang": CardInfo{Count: 2, SCore: 3, Name: "连长"},
		"yingzhang": CardInfo{Count: 2, SCore: 4, Name: "营长"},
		"tuanzhang": CardInfo{Count: 2, SCore: 5, Name: "团长"},
		"lvzhang":   CardInfo{Count: 2, SCore: 6, Name: "旅长"},
		"shizhang":  CardInfo{Count: 2, SCore: 7, Name: "师长"},
		"junzhang":  CardInfo{Count: 1, SCore: 8, Name: "军长"},
		"siling":    CardInfo{Count: 1, SCore: 9, Name: "司令"},
		"junqi":     CardInfo{Count: 1, SCore: 0, Name: "军旗"},
		"dilei":     CardInfo{Count: 3, SCore: 100, Name: "地雷"},
		"zhadan":    CardInfo{Count: 2, SCore: 101, Name: "炸弹"}}
	initJson, _ := json.Marshal(cardMap)
	mString := string(initJson)
	cmdMsgResp.Message = mString
	cmdMsgResp.FromId = SYSTEM_NAME
	cmdMsgResp.ToId = cmdMsg.FromId
	cmdMsgResp.Type = REQ_INIT_DATA_RESP
	return cmdMsgResp
}

/*
	发起认输
*/
func giveUp(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>giveUp============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Message = "放弃认输"
	cmdMsgResp.Type = REQ_GIVEUP_RESP
	return cmdMsgResp
}

/*
	另一玩家确认请求
*/
func disconnClear(c *websocket.Conn) {
	log.Println("=========>disconnClear============>")
	playIdObj, ok := GConn2IdMap.Load(c)
	if !ok {
		log.Println("缓存信息没有获取到")
		return
	}
	playId, ret := playIdObj.(string)
	if !ret {
		log.Println("类型断言错误")
		return
	}

	playerObj, ok := GId2ConnMap.Load(playId)
	if !ok {
		log.Println(playId, "缓存信息没有获取到")
		return
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return
	}

	//删除SN对应的缓存
	GId2ConnMap.Delete(playId)
	GConn2IdMap.Delete(c)
	log.Println("处理断开之后的清理")
	//如果此用户有关联用户，需要提醒对方
	if player.ToNickName != "" {
		var cmdMsg CommandMsg
		cmdMsg.Type = OFFLINE_MSG
		cmdMsg.FromId = playId
		cmdMsg.Message = "下线通知"
		cmdMsg.ToId = player.ToNickName
		proxyMsg(nil, cmdMsg)
	}

}

/*
   主流程处理
*/

func gameHandle(w http.ResponseWriter, r *http.Request) {
	log.Println("==================>")
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	var cmdMsg CommandMsg
	var cmdMsgResp CommandMsgResp
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			disconnClear(c)
			break
		}
		if err = json.Unmarshal(message, &cmdMsg); err != nil {
			log.Println("Unmarshal:", err)
		}
		switch cmdMsg.Type {
		case PLAY_CARD:
			cmdMsgResp = playCard(c, cmdMsg)
		case QUERY_RESULT:
			cmdMsgResp = queryResult(c, cmdMsg)
		case SIGN_IN:
			cmdMsgResp = signIn(c, HUMAN_TYPE, cmdMsg)
		case ROBOT_SIGN_IN:
			cmdMsgResp = signIn(c, ROBOT_TYPE, cmdMsg)
		case SEND_MSG:
			cmdMsgResp = sendMsg(c, cmdMsg)
		case SEND_VOICE:
			cmdMsgResp = sendVoice(c, cmdMsg)
		case GET_USERS:
			cmdMsgResp = getUsers(c, cmdMsg)
		case GET_ROBOTS:
			cmdMsgResp = getRobots(c, cmdMsg)
		case REQ_PLAY:
			cmdMsgResp = reqPlay(c, cmdMsg)
		case REQ_PLAY_YES:
			cmdMsgResp = reqPlayYes(c, cmdMsg)
		case REQ_PLAY_NO:
			cmdMsgResp = reqPlayNo(c, cmdMsg)
		case REQ_INIT_DATA:
			cmdMsgResp = initData(c, cmdMsg)
		case REQ_GIVEUP:
			cmdMsgResp = giveUp(c, cmdMsg)
		case START_GAME:
			cmdMsgResp = startGame(c, cmdMsg)
		case CHANGE_USER:
			cmdMsgResp = changeUser(c, cmdMsg)
		case REQ_PLAY_CARD:
			cmdMsgResp = reqPlayCard(c, cmdMsg)
		}
		msg, err := json.Marshal(cmdMsgResp)
		err = c.WriteMessage(mt, msg)
		log.Println("发送的消息：", mt, cmdMsgResp)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

/*
	HTTP 应答公共方法
*/
func writeResp(response interface{}, w http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(string(json))
	fmt.Fprintf(w, string(json))
}

/*
   用户注册处理函数
*/

func singup(w http.ResponseWriter, r *http.Request) {
	log.Println("==============>singup==========>")
	t1 := time.Now()
	var sign SignUp

	var singupResp SignUpResp
	reqBuf, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(reqBuf, &sign); err != nil {
		log.Println("解析JSON错误", err)
		singupResp.ResultCode = "1000"
		singupResp.ResultMsg = "解析JSON错误"
		writeResp(singupResp, w, r)
		return
	}
	defer r.Body.Close()
	log.Println(t1, "=====>", sign)
	user := afuser.New(dbutil.GetDB(), afuser.DEBUG)
	var search afuser.Search
	search.UserName = sign.UserName
	if _, err = user.Get(search); err == nil {
		singupResp.ResultCode = "4000"
		singupResp.ResultMsg = "用户已经存在"
		writeResp(singupResp, w, r)
		return
	}
	var e afuser.AfUser
	e.UserName = sign.UserName
	e.UserPwd = sign.Pwd
	e.Problem = sign.Problem
	e.Answer = sign.Answer
	e.UserImage = sign.HeadImage
	e.InsertDate = time.Now().Unix()
	e.UserId = time.Now().Unix()
	user.InsertEntity(e, nil)

	singupResp.ResultCode = "0000"
	singupResp.ResultMsg = "注册成功！"
	writeResp(singupResp, w, r)

	return
}

/*
   重置密码函数
*/

func resetpwd(w http.ResponseWriter, r *http.Request) {
	t1 := time.Now()
	var resetpwd ResetPwd
	var resetpwdResp ResetPwdResp
	reqBuf, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(reqBuf, &resetpwd); err != nil {
		log.Println("解析JSON错误", err)
		resetpwdResp.ResultCode = "1000"
		resetpwdResp.ResultMsg = "解析JSON错误"
		writeResp(resetpwdResp, w, r)
		return
	}
	defer r.Body.Close()
	log.Println(t1, "=====>", resetpwd)
	user := afuser.New(dbutil.GetDB(), afuser.DEBUG)
	var search afuser.Search
	search.UserName = resetpwd.UserName
	u, err := user.Get(search)
	if err != nil {
		resetpwdResp.ResultCode = "2000"
		resetpwdResp.ResultMsg = "用户不存在"
		writeResp(resetpwdResp, w, r)
		return
	}
	if u.Problem != resetpwd.Problem && u.Answer != resetpwd.Answer {
		resetpwdResp.ResultCode = "2001"
		resetpwdResp.ResultMsg = "回答问题答案错误"
		writeResp(resetpwdResp, w, r)
		return
	}
	onlineMap := map[string]interface{}{
		"user_pwd": resetpwd.NewPwd}
	err = user.UpdateMap(fmt.Sprintf("%d", u.AutoId), onlineMap, nil)

	resetpwdResp.ResultCode = "0000"
	resetpwdResp.ResultMsg = "修改密码成功"
	writeResp(resetpwdResp, w, r)

	return
}

/*
   查询用户信息
*/

func getUser(w http.ResponseWriter, r *http.Request) {
	t1 := time.Now()
	var getuser GetUser
	var getresp GetUserResp
	reqBuf, err := ioutil.ReadAll(r.Body)
	if err = json.Unmarshal(reqBuf, &getuser); err != nil {
		log.Println("解析JSON错误", err)
		getresp.ResultCode = "1000"
		getresp.ResultMsg = "解析JSON错误"
		writeResp(getresp, w, r)
		return
	}
	defer r.Body.Close()
	log.Println(t1, "=====>", getuser)
	user := afuser.New(dbutil.GetDB(), afuser.DEBUG)
	var search afuser.Search
	search.UserName = getuser.UserName
	u, err := user.Get(search)
	if err != nil {
		getresp.ResultCode = "2000"
		getresp.ResultMsg = "用户不存在"
		writeResp(getresp, w, r)
		return
	}
	getresp.Problem = u.Problem
	getresp.Answer = u.Answer
	getresp.HeadImage = u.UserImage
	getresp.ResultCode = "0000"
	getresp.ResultMsg = "查询用户成功"
	writeResp(getresp, w, r)

	return
}

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)
	log.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   "ArmyFight.log",
		MaxSize:    500, // megabytes
		MaxBackups: 50,
		MaxAge:     90, //days
	}))
	envConf := flag.String("env", "config-ci.json", "select a environment config file")
	flag.Parse()
	log.Println("config file ==", *envConf)
	c, err := goconf.New(*envConf)
	if err != nil {
		log.Fatalln("读配置文件出错", err)
	}

	//填充配置文件
	c.Get("/config/LISTEN_PORT", &listenPort)
	c.Get("/config/DB_URL", &dbUrl)
	c.Get("/config/OPEN_CONNS", &openConns)
	c.Get("/config/IDLE_CONNS", &idleConns)
}

func main() {
	fmt.Println("====>ArmyFight Starting....===>")

	dbutil.InitDB(dbUrl, idleConns, openConns)

	http.HandleFunc("/echo", gameHandle)
	http.HandleFunc("/army/api/getuser", getUser)
	http.HandleFunc("/army/api/signup", singup)
	http.HandleFunc("/army/api/resetpwd", resetpwd)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
