package dns

import (
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/pipe"
	"github.com/rock-go/rock/xbase"
)

var xEnv *xbase.EnvT

func (m *monitor) pipeL(L *lua.LState) int {
	pv := pipe.LValue(L.Get(1))
	if pv == nil {
		return 0
	}

	m.cfg.pipe = append(m.cfg.pipe, pv)
	return 0
}

func (m *monitor) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "pipe":
		return L.NewFunction(m.pipeL)
	}
	return lua.LNil
}

func constructor(L *lua.LState) int {
	cfg := newConfig(L)
	proc := L.NewProc(cfg.name, typeof)
	if proc.IsNil() {
		proc.Set(newM(cfg))
	} else {
		m := proc.Data.(*monitor)
		xEnv.Free(m.cfg.co)

		m.cfg = cfg
	}

	L.Push(proc)
	return 1
}

/*
	local dns = linux.dns{
		name = "monitor_dns",
		bind = "udp://0.0.0.0:53",
    }
	dns.pipe(_(tx) end)


*/

func Inject(env *xbase.EnvT, x lua.UserKV) {
	xEnv = env
	x.Set("dns", lua.NewFunction(constructor))
}
