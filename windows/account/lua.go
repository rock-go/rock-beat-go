package account

import (
	"github.com/rock-go/rock/lua"
)

func getAccountByLua(L *lua.LState) int {
	accounts := GetAccounts()
	L.Push(L.NewAnyData(&accounts))
	return 1
}

func getGroupByLua(L *lua.LState) int {
	groups := GetGroups()
	L.Push(L.NewAnyData(&groups))
	return 1
}

func Inject(kv lua.UserKV) {
	kv.Set("account", lua.NewFunction(getAccountByLua))
	kv.Set("group", lua.NewFunction(getGroupByLua))
}
