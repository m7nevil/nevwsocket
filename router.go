package nev

import (
	"github.com/gorilla/websocket"
	"reflect"
)

const (
	DEFAULT_ACTION = "Default"
	BEFORE_ACTION = "BeforeRequest"
	AFTER_ACTION = "AfterRequest"
)

type Module interface {
	Default(fd uint32, data interface{}) bool
	BeforeRequest(fd uint32, data interface{}) bool
	AfterRequest(fd uint32, data interface{}) bool
}

type Event interface {
	OnOpen(fd uint32, conn *websocket.Conn) bool
	OnClose(fd uint32)
	OnMessage(fd uint32, msg string) bool
}

type RouterMap struct {
	actions map[string] func(uint32, interface{}) bool
	methods map[string] map[string] func(uint32, interface{}) bool
	modules map[string] Module
	events Event
}

func NewRouterMap() *RouterMap {
	return &RouterMap{
		methods: make(map[string] map[string] func(uint32, interface{}) bool),
		actions: make(map[string] func(uint32, interface{}) bool),
		modules: make(map[string] Module),
	}
}

var routerMapInstance *RouterMap
func RMInstance() *RouterMap{
	if routerMapInstance == nil {
		routerMapInstance = NewRouterMap()
	}
	return routerMapInstance
}



func (this *RouterMap) RegisterEvent(e Event){
	this.events = e
}

func (this *RouterMap) RegisterAction(action string, function func(uint32, interface{}) bool) bool{

	if _, exist := this.actions[action]; !exist {
		this.actions[action] = function
		return true
	}
	return false
}

func (this *RouterMap) RegisterModule(moduleName string, module Module) bool {
	if _, exist := this.methods[moduleName]; exist {
		return false
	}

	this.methods[moduleName] = make(map[string] func(uint32, interface{}) bool)
	this.modules[moduleName] = module

	tmpType := reflect.TypeOf(module)
	tmpValue := reflect.ValueOf(module)
	for i := 0; i < tmpType.NumMethod(); i ++ {
		tmp := tmpValue.Method(i).Interface()
		if tmpFunc, ok := tmp.(func(uint32, interface{})bool); ok {
			this.methods[moduleName][tmpType.Method(i).Name] = tmpFunc
		}
	}
	return true
}

func (this *RouterMap) HookAction(funcName string, fd uint32, data interface{}) bool {
	if function, exist := this.actions[funcName]; exist {
		return function(fd, data)
	}
	return false
}

func (this *RouterMap) HookModule(moduleName string, method string, fd uint32, data interface{}) bool {
	if _, exist := this.methods[moduleName]; !exist {
		return false
	}

	if this.methods[moduleName][BEFORE_ACTION](fd, data) == false {
		return false
	}

	if action, exist := this.methods[moduleName][method]; exist {
		if action(fd, data) == false {
			return false
		}
	} else {
		if this.methods[moduleName][DEFAULT_ACTION](fd, data) == false {
			return false
		}
	}

	if this.methods[moduleName][AFTER_ACTION](fd, data) == false{
		return false
	}

	return true
}

func (this *RouterMap) Hook(fd uint32, requestData map[string]interface{}) bool{
	//if this.OnMessage(fd, requestData) == false {
	//	return false
	//}

	if action, exist := requestData["action"].(string); exist {
		if this.HookAction(action, fd, requestData["data"]) == false {
			return false
		}
	} else {
		if this.HookModule(requestData["module"].(string), requestData["method"].(string), fd, requestData["data"]) == false {
			return false
		}
	}

	return true
}

func (this *RouterMap) OnClose(fd uint32) {
	if this.events != nil {
		this.events.OnClose(fd)
	}
}

func (this *RouterMap) OnOpen(fd uint32, conn *websocket.Conn) bool{
	if this.events != nil {
		return this.events.OnOpen(fd, conn)
	}
	return true
}

func (this *RouterMap) OnMessage(fd uint32,  msg string) bool {
	if this.events != nil {
		return this.events.OnMessage(fd, msg)
	}
	return true
}
