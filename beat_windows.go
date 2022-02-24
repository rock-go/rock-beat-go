package beat

import (
	"github.com/rock-go/rock-beat-go/windows/event"
	"github.com/rock-go/rock-beat-go/windows/registry"
	"github.com/rock-go/rock-beat-go/windows/wmi"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xbase"
)

func LuaInjectApi(env *xbase.EnvT) {
	win := lua.NewUserKV()
	event.Inject(env, win)
	wmi.Inject(env, win)
	registry.Inject(env, win)

	env.Global("win", win)
}
