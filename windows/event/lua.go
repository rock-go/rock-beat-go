package event

import (
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/pipe"
	"github.com/rock-go/rock/xbase"
	"reflect"
	"strings"
)

var (
	xEnv *xbase.EnvT
	//instance *winEv
	winEvBucketOffset = "windows_event_record_offset"
	winLoginBucket    = "windows_access_log"
	winEvTypeOf       = reflect.TypeOf((*winEv)(nil)).String()
)


func (wv *winEv) subscribeL(L *lua.LState) int {
	name := L.CheckString(1)
	query := L.CheckString(2)

	if !wv.inChannel(name) {
		wv.cfg.channel = append(wv.cfg.channel, channel{name, query})
	}
	return 0
}

func (wv *winEv) pipeL(L *lua.LState) int {
	pv := pipe.LValue(L.Get(1))
	if pv != nil {
		wv.cfg.pipe = append(wv.cfg.pipe, pv)
	}

	return 0
}

func (wv *winEv) toL(L *lua.LState) int {
	wv.cfg.sdk = auxlib.CheckWriter(L.Get(1) , L)
	return 0
}


func (wv *winEv) Index(L *lua.LState, key string) lua.LValue {

	switch key {
	case "subscribe":
		return L.NewFunction(wv.subscribeL)

	case "pipe":
		return L.NewFunction(wv.pipeL)

	case "to":
		return L.NewFunction(wv.toL)

	default:
		//todo
	}

	return lua.LNil
}

func (wv *winEv) NewIndex(L *lua.LState, key string, val lua.LValue) {
	if strings.HasPrefix(key, "ev_") {
		wv.cfg.chains.Set(key[3:], lua.CheckFunction(L, val))
	}
}

func constructor(L *lua.LState) int {
	cfg := newConfig(L)
	proc := L.NewProc(cfg.name, winEvTypeOf)
	if proc.IsNil() {
		proc.Set(newWinEv(cfg))
	} else {
		proc.Data.(*winEv).cfg = cfg
	}
	L.Push(proc)
	return 1
}

func Inject(env *xbase.EnvT, ukv lua.UserKV) {
	xEnv = env
	ukv.Set("event", lua.NewFunction(constructor))
}
