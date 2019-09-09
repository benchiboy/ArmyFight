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

var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

/*
	玩家签到处理
	1:从数据库查询用户的信息 同步到内存中

*/
func signIn(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SignIn============>")
	GId2ConnMap.Store(cmdMsg.NickName, Player{CurrConn: c, SignInTime: time.Now(),
		NickName: cmdMsg.NickName, Status: STATUS_ONLIN_IDLE})
	GConn2IdMap.Store(c, cmdMsg.NickName)
	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Type = SIGN_IN
	cmdMsgResp.Success = true
	cmdMsgResp.Message = "Sign In Success!"

	return cmdMsgResp
}

/*
	玩家发牌处理
*/

func playCard(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>PlayCard============>")
	var cmdMsgResp CommandMsgResp
	return cmdMsgResp
}

/*
	玩家发牌对比结果
*/

func playCardResult(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>playCardResult============>")
	var cmdMsgResp CommandMsgResp
	return cmdMsgResp
}

/*
	玩家发送消息处理
*/
func sendMsg(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SendMsg============>")
	var cmdMsgResp CommandMsgResp
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
		cmdMsgResp.Type = GET_USERS
		cmdMsgResp.Success = true
	}
	return cmdMsgResp
}

/*
	请求玩家
*/
func reqPlay(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlay============>")
	var cmdMsgResp CommandMsgResp
	log.Println(cmdMsg.FromId, cmdMsg.ToId)

	return cmdMsgResp
}

/*
	另一玩家确认请求
*/
func reqPlayOK(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayOK============>")
	var cmdMsgResp CommandMsgResp
	return cmdMsgResp
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
			log.Println("read:", err)
			break
		}
		if err = json.Unmarshal(message, &cmdMsg); err != nil {
			log.Println("Unmarshal:", err)
		}
		switch cmdMsg.Type {
		case PLAY_CARD:
			cmdMsgResp = playCard(c, cmdMsg)
		case PLAY_CARD_RESULT:
			cmdMsgResp = playCardResult(c, cmdMsg)
		case SIGN_IN:
			cmdMsgResp = signIn(c, cmdMsg)
		case SEND_MSG:
			cmdMsgResp = sendMsg(c, cmdMsg)
		case GET_USERS:
			cmdMsgResp = getUsers(c, cmdMsg)
		case REQ_PLAY:
			cmdMsgResp = reqPlay(c, cmdMsg)
		case REQ_PALYOK:
			cmdMsgResp = reqPlayOK(c, cmdMsg)
		}
		msg, err := json.Marshal(cmdMsgResp)
		err = c.WriteMessage(mt, msg)
		log.Println(mt, cmdMsgResp)
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
