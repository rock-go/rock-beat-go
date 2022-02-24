package beat

import (
	"github.com/rock-go/rock-beat-go/linux/dns"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xbase"
)

func LuaInjectApi(env *xbase.EnvT) {
	linux := lua.NewUserKV()
	dns.Inject(env, linux)
	env.Global("linux", linux)
}
