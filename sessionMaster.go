package nev

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type SessionMaster struct {
	isWebSocket bool
	sessions sync.Map
	keys sync.Map
}

func NewSessionMaster() *SessionMaster{
	return &SessionMaster{}
}

// session管理器单例
var smInstance *SessionMaster = nil
func SMInstance() *SessionMaster{
	if smInstance == nil {
		smInstance = NewSessionMaster()
	}
	return smInstance
}

func (this *SessionMaster) GetSessionById(id uint32) *Session {
	value, ok := this.sessions.Load(id)
	if ok {
		if sess, ok := value.(*Session); ok {
			return sess
		}
	}

	return nil
}

func (this *SessionMaster) SetSession(fd uint32, conn *websocket.Conn) {
	sess := NewSession(fd, conn)
	this.sessions.Store(fd, sess)
}

func (this *SessionMaster) SetKey(sess *Session, key string) {
	sess.SetKey(key)
	this.keys.Store(key, sess.Id)
}

func (this *SessionMaster) getIdByKey(key string) uint32{
	id, ok := this.keys.Load(key)
	if !ok {
		return 0
	}
	return id.(uint32)
}

func (this *SessionMaster) DelSessionById(id uint32) {
	value, ok := this.sessions.Load(id)
	if ok {
		if sess, ok := value.(*Session); ok {
			sess.close()
		}
	}
	this.sessions.Delete(id)
}

// 一起发送
func (this *SessionMaster) SendToAll(msg string) {
	// TODO msg打包处理

	this.sessions.Range(func(k, v interface{}) bool {
		if v, ok := v.(*Session); ok {
			if err := v.write(msg); err != nil {
				this.DelSessionById(k.(uint32))
				log.Println(err)
			}
		}
		return true
	})
}


// 单个发送消息
func (this *SessionMaster) SendById(id uint32, msg string) bool {

	sess := this.GetSessionById(id)
	if sess != nil {
		if err := sess.write(msg); err == nil {
			return true
		}
	}

	this.DelSessionById(id)
	return false
}

func (this *SessionMaster) SendByKey(key string, msg string) bool {
	id := this.getIdByKey(key)
	if id == 0 {
		return false
	}
	return this.SendById(id, msg)
}


// 心跳检测
func (this *SessionMaster) HeartBeat(num int64) {
	for {
		time.Sleep(time.Second)
		this.sessions.Range(func(k, v interface{}) bool {
			tem, ok := v.(*Session)
			if !ok {
				return true
			}

			if time.Now().Unix() - tem.times > num {
				this.DelSessionById(k.(uint32))
			}
			return true
		})
	}
}