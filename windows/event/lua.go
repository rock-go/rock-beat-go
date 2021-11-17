package event

import (
	"github.com/rock-go/rock/audit"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/node"
	"reflect"
	"strings"
)

var (
	instance *winEv
	winEvBucket = "windows_event_log"
	winLoginBucket = "windows_access_log"
	winEvTypeOf = reflect.TypeOf((*winEv)(nil)).String()
)

func init() {
	instance = newWinEv(normal())
}

func (wv *winEv) subscribe(L *lua.LState) int {
	channel := L.CheckString(1)
	query   := L.CheckString(2)

	var bookmark []byte
	var err error
	if wv.cfg.begin {
		wv.watcher.SubscribeFromBeginning(channel , query)
		goto loop
	}

	bookmark , err = node.Get(winEvBucket, channel + "_last")
	if err != nil {
		wv.watcher.SubscribeFromBeginning(channel, query)
		goto loop
	}

	audit.NewEvent("win-log").
		Subject("%s last bookmark" , channel).
		From(wv.vm.CodeVM()).
		Msg("%s" , bookmark).Log().Put()

	wv.watcher.SubscribeFromBookmark(channel ,query , string(bookmark))

loop:
	for _ , item := range wv.channel {
		if item != channel {
			wv.channel = append(wv.channel , item)
		}
	}

	return 0
}

func (wv *winEv) Index(L *lua.LState , key string) lua.LValue {

	switch key {
	case "subscribe":
		return L.NewFunction(wv.subscribe)

	default:
		//todo
	}

	return lua.LNil
}

func (wv *winEv) NewIndex(L *lua.LState , key string , val lua.LValue) {

	switch key {
	case "hook":
		wv.cfg.hook = lua.CheckFunction(L , val)

	default:
		if strings.HasPrefix(key , "ev_") {
			wv.cfg.chains.Set(key[3:] , lua.CheckFunction(L ,val))
		}
	}
}


func newWinLogApi(L *lua.LState) int {
	cfg := newConfig(L)
	proc := L.NewProc(cfg.name , winEvTypeOf)
	if proc.IsNil() {
		instance.cfg = cfg
		proc.Set(instance)
	} else {
		instance.cfg = cfg
	}

	L.Push(proc)
	proc.Value.(*winEv).vm = L
	return 1
}

func Inject(ukv lua.UserKV) {
	ukv.Set("winlog" , lua.NewFunction(newWinLogApi))
}