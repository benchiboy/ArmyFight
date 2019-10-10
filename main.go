// ArmFight project ArmFight.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"time"

	"github.com/gorilla/websocket"
)

var (
	GId2ConnMap = &sync.Map{}
	GConn2IdMap = &sync.Map{}
)

var addr = flag.String("addr", "172.17.0.3:9080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

/*
	玩家签到处理
	1:从数据库查询用户的信息 同步到内存中

*/
func signIn(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
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
	GId2ConnMap.Store(cmdMsg.NickName, Player{CurrConn: c, SignInTime: time.Now(),
		NickName: cmdMsg.NickName, Status: STATUS_ONLIN_IDLE})
	GConn2IdMap.Store(c, cmdMsg.NickName)

	cmdMsgResp.FromId = cmdMsg.FromId
	cmdMsgResp.Message = "Sign In Success!"
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
	玩家发牌对比结果
*/

func queryResult(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>queryResult============>")
	sScore, sCard := getCard(cmdMsg.FromId)
	mScore, mCard := getCard(cmdMsg.ToId)

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

	cmdMsgResp.Role = "M"

	proxyMsgResp(cmdMsg.ToId, cmdMsgResp)

	cmdMsgResp.Role = "S"
	cmdMsgResp.Message = sCard
	cmdMsgResp.AnotherMsg = mCard
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
	得到用户列表
*/
func getUsers(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>getUsers============>")
	uList := make([]User, 0)
	var cmdMsgResp CommandMsgResp
	GId2ConnMap.Range(func(k, v interface{}) bool {
		p, _ := v.(Player)
		var u User
		if cmdMsg.NickName != p.NickName {
			u.UserId = fmt.Sprintf("%s", k)
			u.NickName = p.NickName
			u.Avatar = p.Avatar
			u.Candy = p.Candy
			u.Decoration = p.Decoration
			u.Icecream = p.Icecream
			u.LoginTime = p.SignInTime
			u.Memo = p.Memo
			u.Status = p.Status
			uList = append(uList, u)
		}
		return true
	})
	if userBuf, err := json.Marshal(uList); err != nil {
		log.Println(err)
	} else {
		cmdMsgResp.Message = string(userBuf)
		cmdMsgResp.Success = true
	}
	cmdMsgResp.Type = GET_USERS_RESP
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
	}
	toPlayer, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
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
	log.Println("=========>setRole============>", playerId)
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
	log.Println("=========>setStatus============>")
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
	log.Println("=========>setStatus============>", playerId)
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

	setRole(cmdMsg.FromId, "S")
	setRole(cmdMsg.ToId, "M")

	cmdMsgResp.Type = REQ_PALY_YES_RESP
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

	cmdMsgResp.Type = REQ_PALY_NO_RESP
	return cmdMsgResp
}

/*
	登录成功后初始化数据
*/
func initData(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>initData============>")
	var cmdMsgResp CommandMsgResp
	cardMap := map[string]int{"工兵": 3, "排长": 2}
	mjson, _ := json.Marshal(cardMap)
	mString := string(mjson)
	fmt.Printf("print mString:%s", mString)
	cmdMsgResp.Message = mString

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
	}
	playId, ret := playIdObj.(string)
	if !ret {
		log.Println("类型断言错误")
		return
	}
	//删除SN对应的缓存
	GId2ConnMap.Delete(playId)
	GConn2IdMap.Delete(c)
	log.Println("处理断开之后的清理")
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
		log.Println(string(message))
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
			cmdMsgResp = signIn(c, cmdMsg)
		case SEND_MSG:
			cmdMsgResp = sendMsg(c, cmdMsg)
		case SEND_VOICE:
			cmdMsgResp = sendVoice(c, cmdMsg)
		case GET_USERS:
			cmdMsgResp = getUsers(c, cmdMsg)
		case REQ_PLAY:
			cmdMsgResp = reqPlay(c, cmdMsg)
		case REQ_PALY_YES:
			cmdMsgResp = reqPlayYes(c, cmdMsg)
		case REQ_PALY_NO:
			cmdMsgResp = reqPlayNo(c, cmdMsg)
		case REQ_INIT_DATA:
			cmdMsgResp = initData(c, cmdMsg)
		case REQ_GIVEUP:
			cmdMsgResp = giveUp(c, cmdMsg)
		case START_GAME:
			cmdMsgResp = startGame(c, cmdMsg)
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

func main() {
	fmt.Println("=====>ArmFight======>")

	http.HandleFunc("/echo", gameHandle)

	http.HandleFunc("/getUsers", gameHandle)

	log.Fatal(http.ListenAndServe(*addr, nil))
	fmt.Println("hello")

}
