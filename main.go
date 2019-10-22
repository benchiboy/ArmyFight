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

	"ArmFight/afplay"
	"ArmFight/afplaydetail"
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

var (
	serverAddr = flag.String("addr", "127.0.0.1:9080", "http service address")
	upgrader   = websocket.Upgrader{} // use default options
)

/*
	玩家签到处理
	1:从数据库查询用户的信息 同步到内存中
*/
func signIn(c *websocket.Conn, playerType int, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SignIn============>")
	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Type = SIGN_IN_RESP
	cmdMsgResp.Success = true
	_, ok := GId2ConnMap.Load(cmdMsg.FromId)
	if ok {
		log.Println(cmdMsg.FromId + "用户已经在线")
		cmdMsgResp.Success = false
		cmdMsgResp.Message = "用户已经在线"
		return cmdMsgResp
	}
	user := afuser.New(dbutil.GetDB(), afuser.DEBUG)
	var search afuser.Search
	search.UserName = cmdMsg.FromId
	e, err := user.Get(search)
	if err != nil {
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
	GId2ConnMap.Store(cmdMsg.FromId, Player{
		CurrConn:   c,
		SignInTime: time.Now(),
		Status:     STATUS_ONLIN_IDLE,
		PlayerType: playerType,
		Avatar:     e.UserImage,
		NickName:   cmdMsg.FromId,
		Coins:      e.CoinCnt,
		Medals:     e.MedalCnt})

	GConn2IdMap.Store(c, cmdMsg.FromId)
	cmdMsgResp.FromId = SYSTEM_NAME
	cmdMsgResp.ToId = cmdMsg.FromId
	cmdMsgResp.Message = e.UserImage
	return cmdMsgResp
}

/*
	玩家发牌处理
*/

func playCard(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========PlayCard======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	cmdMsg.Role = getRole(cmdMsg.FromId)
	proxyMsg(c, cmdMsg)

	//插入明细表
	playdt := afplaydetail.New(dbutil.GetDB(), afuser.DEBUG)
	var e afplaydetail.AfPlayDetail
	e.BatchNo = cmdMsg.BatchNo
	e.PlayNo = cmdMsg.PlayNo
	e.PlayCard = cmdMsg.Message
	e.Player = cmdMsg.FromId
	e.InsertDate = time.Now().Unix()
	playdt.InsertEntity(e, nil)

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
	log.Println("==========ReqPlayCard======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	proxyMsg(c, cmdMsg)
	var cmdMsgResp CommandMsgResp

	cmdMsgResp.Message = cmdMsg.Message
	cmdMsgResp.Type = REQ_PLAY_CARD_RESP
	return cmdMsgResp
}

/*
	功能：更新决战的结果
*/

func updateOver(batchNo string, endType string, winner string) {
	log.Println("==========UpdateResult======>")
	rrr := afplay.New(dbutil.GetDB(), afuser.DEBUG)
	playMap := map[string]interface{}{FIELD_ENDTYPE: endType, FIELD_WINNER: winner, FIELD_UPDATE_TIME: time.Now()}
	if err := rrr.UpdateMap(batchNo, playMap, nil); err != nil {
		log.Println("更新对战批次失败", err)
	}
}

/*
	功能：更新决战的结果
*/
func updateResult(batchNo string, playNo int64, fromId string, toId string, winner string, isEnd string) {
	log.Println("==========UpdateResult======>", batchNo, playNo, fromId, toId, winner)
	//更新明细表
	r := afplaydetail.New(dbutil.GetDB(), afuser.DEBUG)
	tr, err := r.DB.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	if winner == WINNER_BOTH {
		detailMap := map[string]interface{}{FIELD_RESULT: RESULT_EQUAL, FIELD_UPDATE_TIME: time.Now()}
		if err = r.UpdateMap(batchNo, playNo, fromId, detailMap, tr); err != nil {
			log.Println("更新对战明细失败", err)
			tr.Rollback()
		}
		detailMap = map[string]interface{}{FIELD_RESULT: RESULT_EQUAL, FIELD_UPDATE_TIME: time.Now()}
		if err = r.UpdateMap(batchNo, playNo, toId, detailMap, tr); err != nil {
			log.Println("更新对战明细失败", err)
			tr.Rollback()
		}

		rr := afuser.New(dbutil.GetDB(), afuser.DEBUG)
		userMap := map[string]interface{}{FIELD_COIN_CNT: 11, FIELD_UPDATE_TIME: time.Now()}
		if err = rr.UpdateMap(fromId, userMap, tr); err != nil {
			log.Println("更新用户失败", err)
			tr.Rollback()
		}

		userMap = map[string]interface{}{FIELD_COIN_CNT: 11, FIELD_UPDATE_TIME: time.Now()}
		if err = rr.UpdateMap(toId, userMap, tr); err != nil {
			log.Println("更新用户失败", err)
			tr.Rollback()
		}
	}
	if winner == fromId {
		detailMap := map[string]interface{}{FIELD_RESULT: RESULT_WINNER, FIELD_UPDATE_TIME: time.Now()}
		if err = r.UpdateMap(batchNo, playNo, fromId, detailMap, tr); err != nil {
			log.Println("更新对战明细失败", err)
			tr.Rollback()
		}
		detailMap = map[string]interface{}{FIELD_RESULT: RESULT_LOSER, FIELD_UPDATE_TIME: time.Now()}
		if err = r.UpdateMap(batchNo, playNo, toId, detailMap, tr); err != nil {
			log.Println("更新对战明细失败", err)
			tr.Rollback()
		}

		rr := afuser.New(dbutil.GetDB(), afuser.DEBUG)
		userMap := map[string]interface{}{FIELD_COIN_CNT: 11, FIELD_UPDATE_TIME: time.Now()}
		if err = rr.UpdateMap(fromId, userMap, tr); err != nil {
			log.Println("更新用户失败", err)
			tr.Rollback()
		}
	}
	if winner == toId {
		detailMap := map[string]interface{}{FIELD_RESULT: RESULT_WINNER, FIELD_UPDATE_TIME: time.Now()}
		if err = r.UpdateMap(batchNo, playNo, toId, detailMap, tr); err != nil {
			log.Println("更新对战明细失败", err)
			tr.Rollback()
		}
		detailMap = map[string]interface{}{FIELD_RESULT: RESULT_LOSER, FIELD_UPDATE_TIME: time.Now()}
		if err = r.UpdateMap(batchNo, playNo, fromId, detailMap, tr); err != nil {
			log.Println("更新对战明细失败", err)
			tr.Rollback()
		}

		rr := afuser.New(dbutil.GetDB(), afuser.DEBUG)
		userMap := map[string]interface{}{FIELD_COIN_CNT: 11, FIELD_UPDATE_TIME: time.Now()}
		if err = rr.UpdateMap(toId, userMap, tr); err != nil {
			log.Println("更新用户失败", err)
			tr.Rollback()
		}

	}
	//更新批次信息
	if isEnd == "E" {
		rrr := afplay.New(dbutil.GetDB(), afuser.DEBUG)
		playMap := map[string]interface{}{FIELD_STATUS: "e", FIELD_ENDTYPE: END_NORMAL, FIELD_WINNER: winner, FIELD_UPDATE_TIME: time.Now()}
		if err = rrr.UpdateMap(batchNo, playMap, tr); err != nil {
			log.Println("更新对战批次失败", err)
			tr.Rollback()
		}
	}
	tr.Commit()
}

func compareCard(fromPlayer string, toPlayer string, fromScore int, toScore int) (string, string) {
	//出工兵和地雷的情况
	if fromScore == CARD_ENGINEER && toScore == CARD_MINE {
		return GAME_DOING, fromPlayer
	}
	if toScore == CARD_ENGINEER && fromScore == CARD_MINE {
		return GAME_DOING, toPlayer
	}
	//出炸弹和地雷的情况
	if toScore == CARD_MINE && fromScore == CARD_BOMB {
		return GAME_DOING, WINNER_BOTH
	}
	if fromScore == CARD_MINE && toScore == CARD_BOMB {
		return GAME_DOING, WINNER_BOTH
	}
	if fromScore > toScore {
		//出现炸弹，地址 和人员相碰的情况
		if fromScore == CARD_BOMB || fromScore == CARD_MINE {
			if toScore == CARD_TIP {
				return GAME_DOING, toPlayer
			} else {
				return GAME_DOING, WINNER_BOTH
			}
		} else {
			if toScore == CARD_TIP {
				return GAME_END, fromPlayer
			} else {
				return GAME_DOING, fromPlayer
			}
		}
	} else if fromScore == toScore {
		return GAME_DOING, WINNER_BOTH
	} else {
		//出现炸弹，地雷 和人员相碰的情况
		if toScore == CARD_BOMB || toScore == CARD_MINE {
			if fromScore == CARD_TIP {
				return GAME_END, fromPlayer
			} else {
				return GAME_DOING, WINNER_BOTH
			}
		} else {
			if fromScore == CARD_TIP {
				return GAME_END, toPlayer
			} else {
				return GAME_DOING, toPlayer
			}
		}
	}
}

/*
	玩家发牌对比结果
*/

func queryResult(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========queryResult======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	fromScore, fromCard := getCard(cmdMsg.FromId)
	toScore, toCard := getCard(cmdMsg.ToId)
	endTag, winner := compareCard(cmdMsg.FromId,
		cmdMsg.ToId, fromScore, toScore)
	cmdMsg.Message = fromCard
	cmdMsg.Winner = winner
	cmdMsg.Status = endTag
	proxyMsg(c, cmdMsg)
	cmdMsgResp := CommandMsgResp{Type: QUERY_RESULT_RESP, ToId: cmdMsg.ToId,
		FromId: cmdMsg.FromId, Winner: winner, Status: endTag, Message: toCard}
	log.Println(cmdMsg)
	updateResult(cmdMsg.BatchNo, cmdMsg.PlayNo, cmdMsg.FromId, cmdMsg.ToId, winner, endTag)

	return cmdMsgResp
}

/*
	玩家发送消息处理
*/
func sendMsg(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========SendMsg======>", cmdMsg.FromId, "===>", cmdMsg.ToId)

	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = SEND_MSG_RESP
	return cmdMsgResp
}

/*
	玩家发送语音
*/
func sendVoice(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========SendVoice======>", cmdMsg.FromId, "===>", cmdMsg.ToId)

	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = SEND_VOICE_RESP
	return cmdMsgResp
}

/*
	功能：得到人类对手类别
*/
func getUsers(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========GetUsers======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	uList := make([]Player, 0)
	var cmdMsgResp CommandMsgResp
	GId2ConnMap.Range(func(k, v interface{}) bool {
		p, _ := v.(Player)
		if cmdMsg.FromId != p.NickName && p.PlayerType == HUMAN_TYPE {
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
	log.Println("==========GetRobots======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	uList := make([]Player, 0)
	var cmdMsgResp CommandMsgResp
	GId2ConnMap.Range(func(k, v interface{}) bool {
		p, _ := v.(Player)
		if cmdMsg.FromId != p.NickName && p.PlayerType == ROBOT_TYPE {
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
	功能：转发命令到toId节点
*/

func proxyMsg(c *websocket.Conn, cmdMsg CommandMsg) {
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
	功能：玩家切换了对手，需要通知原对手
		 并把原对手的在线状态修改为空闲
*/
func changeUser(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========ChangeUser======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp
	setStatus(cmdMsg.ToId, STATUS_ONLIN_IDLE)
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = CHANGE_USER_RESP
	return cmdMsgResp
}

/*
	请求玩家
*/
func reqPlay(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========ReqPlay======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = REQ_PLAY_RESP
	return cmdMsgResp
}

/*
	开始游戏
*/
func startGame(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========StartGame======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	proxyMsg(c, cmdMsg)

	play := afplay.New(dbutil.GetDB(), afuser.DEBUG)

	var e afplay.AfPlay
	e.BatchNo = cmdMsg.BatchNo
	e.FromPlayer = cmdMsg.FromId
	e.ToPlayer = cmdMsg.ToId
	play.InsertEntity(e, nil)

	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Type = START_GAME_RESP
	return cmdMsgResp
}

/*
	功能说明：修改玩家的角色
	入参：M：主角色 S：从角色
*/
func setRole(playerId string, role string) {
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到", playerId)
		return
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return
	}
	player.Role = role
	GId2ConnMap.Store(playerId, player)
}

/*
	功能说明：得到玩家的角色
	返回：M：主角色 S：从角色
*/
func getRole(playerId string) string {
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到", playerId)
		return ""
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return ""
	}
	return player.Role
}

/*
	功能说明：获取玩家的类型
	返回：1：机器人 2：
*/
func getPlayerType(playerId string) int {
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
		return 0
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return 0
	}
	return player.PlayerType
}

/*
	功能说明：保存玩家的出牌
	参数：玩家ID，牌的分值,牌的ID
*/
func setCard(playerId string, score int, card string) {
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "setCard缓存信息没有获取到")
		return
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("setCard类型断言错误")
		return
	}
	player.CurrSCore = score
	player.CurrCard = card
	GId2ConnMap.Store(playerId, player)
}

/*
	功能说明：得到玩家的出牌
	入参：玩家ID
	出参：分值、牌ID
*/
func getCard(playerId string) (int, string) {
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
	修改玩家状态
*/
func setStatus(playerId string, status int) {
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
		return
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return
	}
	player.Status = status
	GId2ConnMap.Store(playerId, player)
}

/*
	存储另一玩家的信息
*/
func setNickname(playerId string, nikeName string) {
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.ToNickName = nikeName
	GId2ConnMap.Store(playerId, player)
}

/*
	另一玩家答应请求
*/
func reqPlayYes(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========ReqPlayYes======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp
	cmdMsg.Message = "对方同意了"

	proxyMsg(c, cmdMsg)
	//初始玩家双方的状态
	setStatus(cmdMsg.FromId, STATUS_ONLIE_DONG)
	setStatus(cmdMsg.ToId, STATUS_ONLIE_DONG)

	setNickname(cmdMsg.FromId, cmdMsg.ToId)
	setNickname(cmdMsg.ToId, cmdMsg.FromId)

	setRole(cmdMsg.FromId, ROLE_SLAVE)
	setRole(cmdMsg.ToId, ROLE_MASTER)

	cmdMsgResp.Type = REQ_PLAY_YES_RESP
	return cmdMsgResp
}

/*
	另一玩家拒绝请求
*/
func reqPlayNo(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========ReqPlayNo======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
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
	log.Println("==========InitData======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
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
	cards, _ := json.Marshal(cardMap)
	cardStr := string(cards)
	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Message = cardStr
	cmdMsgResp.FromId = SYSTEM_NAME
	cmdMsgResp.ToId = cmdMsg.FromId
	cmdMsgResp.Type = REQ_INIT_DATA_RESP
	return cmdMsgResp
}

/*
	发起认输
*/
func giveUp(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========GiveUp======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Message = "放弃认输"
	cmdMsgResp.Type = REQ_GIVEUP_RESP
	return cmdMsgResp
}

/*
	发起求和
*/
func draw(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("==========Draw======>", cmdMsg.FromId, "===>", cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Message = "请求平局"
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
		setStatus(cmdMsg.ToId, STATUS_ONLIN_IDLE)
		proxyMsg(nil, cmdMsg)
	}

}

/*
   功能：处理手机端的请求指令
*/

func gameHandle(w http.ResponseWriter, r *http.Request) {
	log.Println("==========GiveUp======>")
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
			log.Println("读消息出现错误:ReadMessage Error", err)
			disconnClear(c)
			break
		}
		log.Printf("服务端收到消息: %d %+v \n", mt, string(message))
		if err = json.Unmarshal(message, &cmdMsg); err != nil {
			log.Println("解析消息出现错误:", err)
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
		cmdMsgResp.BatchNo = cmdMsg.BatchNo
		msg, err := json.Marshal(cmdMsgResp)
		err = c.WriteMessage(mt, msg)
		log.Printf("服务端应答消息: %d %+v \n", mt, cmdMsgResp)
		if err != nil {
			log.Println("发送应答消息出错:", err)
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
	log.Println("==============>singup============>")
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
	log.Println("==============>resetpwd============>")
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
	log.Println("==============>getUser============>")
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
	c.Get("/config/LISTEN_IP", serverAddr)
}

func main() {
	fmt.Println("====>ArmyFight Starting....===>")
	fmt.Println("====>Listen Addr===>", *serverAddr)

	dbutil.InitDB(dbUrl, idleConns, openConns)

	http.HandleFunc("/echo", gameHandle)
	http.HandleFunc("/army/api/getuser", getUser)
	http.HandleFunc("/army/api/signup", singup)
	http.HandleFunc("/army/api/resetpwd", resetpwd)

	log.Fatal(http.ListenAndServe(*serverAddr, nil))
}
