package beat

import (
	"github.com/rock-go/rock-beat-go/windows/account"
	"github.com/rock-go/rock-beat-go/windows/event"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xcall"
)

func LuaInjectApi(env xcall.Env) {
	ukv := lua.NewUserKV()
	event.Inject(ukv)
	account.Inject(ukv)
	env.Set("beat" , ukv)
}