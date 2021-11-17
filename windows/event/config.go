package event

import (
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/lua"
)

type config struct {
	name      string
	chains    lua.UserKV
	hook      *lua.LFunction
	pass      []uint64
	begin     bool
	transport lua.Writer
}

func normal() *config {
	return &config{ chains: lua.NewUserKV() }
}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cfg := normal()
	
	tab.Range(func(key string, val lua.LValue) {
		switch key {
		case "name":
			cfg.name = val.String()

		case "begin":
			cfg.begin = lua.CheckBool(L , val)

		case "transport":
			cfg.transport = auxlib.CheckTransport(L , val)

		case "hook":
			cfg.hook = lua.CheckFunction(L , val)

		case "pass":
			switch val.Type() {
			case lua.LTNumber:
				cfg.pass = append(cfg.pass , uint64(val.(lua.LNumber)))
			case lua.LTTable:
				arr := val.(*lua.LTable)
				cfg.pass = append(cfg.pass , auxlib.LTab2SUI64(arr)...)
			}

		default:
			L.RaiseError("%s config not found %s field", winEvTypeOf, key)
		}
	})

	if e := cfg.valid() ; e != nil {
		L.RaiseError("%v" , e)
		return nil
	}

	return cfg
}

func (cfg *config) valid() error {
	return nil
}