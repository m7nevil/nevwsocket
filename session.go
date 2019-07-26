package nev

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Session struct {
	Id    uint32
	Conn  *websocket.Conn
	times int64
	lock  sync.Mutex
	Key   string
}

func NewSession(id uint32, conn *websocket.Conn) *Session {
	return &Session{
		Id:    id,
		Conn:  conn,
		times: time.Now().Unix(),
	}
}

func (this *Session) SetKey(key string){
	this.Key = key
}

func (this *Session) write(msg string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	err := this.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	return err
}

func (this *Session) close() {
	this.Conn.Close()
}

func (this *Session)updateTime(){
	this.times = time.Now().Unix()
}