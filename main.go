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
*/
func signIn(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SignIn============>")
	var cmdMsgResp CommandMsgResp

	GId2ConnMap.Store(cmdMsg.Id, Player{CurrConn: c, SignInTime: time.Now()})
	GConn2IdMap.Store(c, cmdMsg.Id)

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
	玩家发送消息处理
*/
func sendMsg(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SendMsg============>")
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
		//c.SetReadDeadline(time.Now().Add(5 * time.Second))
		mt, message, err := c.ReadMessage()
		log.Println("Recv Origin Messgaeg====>", string(message))
		if err != nil {
			log.Println("read:", err)
			break
		}
		err = json.Unmarshal(message, &cmdMsg)
		if err != nil {
			log.Println("Unmarshal:", err)
		}
		switch cmdMsg.Type {
		case PLAY_CARD:
			cmdMsgResp = playCard(cmdMsg)
		case SIGN_IN:
			cmdMsgResp = signIn(cmdMsg)
		case SEND_MSG:
			cmdMsgResp = sendMsg(cmdMsg)
		}
		msg, err := json.Marshal(cmdMsgResp)
		err = c.WriteMessage(mt, msg)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	fmt.Println("=====>ArmFight======>")

	http.HandleFunc("/echo", gameHandle)

	log.Fatal(http.ListenAndServe(*addr, nil))
	fmt.Println("hello")

}
