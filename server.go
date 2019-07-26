package nev

import (
	"encoding/json"
	"log"
	"flag"
	"github.com/gorilla/websocket"
	"net/http"
)

//var addr = flag.String("addr", "localhost:712", "ws addr")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool{ return true },
}

var fd uint32 = 1000
func handler(w http.ResponseWriter, r *http.Request){
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer conn.Close()

	// 生成session
	fd ++
	SMInstance().SetSession(fd, conn)
	if RMInstance().OnOpen(fd ,conn) == false {
		log.Println("启动失败...")
		return
	}

	connHandler(SMInstance().GetSessionById(fd))
}


func ListenAndServe(addr string) {
	flag.Parse()
	//log.SetFlags(0)
	http.HandleFunc("/", handler)
	//http.HandleFunc("/http", httpHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func connHandler(sess *Session) {
	defer func() {
		SMInstance().DelSessionById(sess.Id)
		RMInstance().OnClose(sess.Id)
	}()

	// TODO 心跳检测

	for {
		_, msg, err := sess.Conn.ReadMessage()
		if err != nil {
			log.Println("Read:", err)
			break
		}
		log.Printf("recv: %s", msg)
		sess.updateTime()

		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Println("connHandler err:", err)
			continue
		}

		// 调用路由
		RMInstance().OnMessage(sess.Id, string(msg))
		if RMInstance().Hook(sess.Id, data) == false {
			return
		}
	}
}