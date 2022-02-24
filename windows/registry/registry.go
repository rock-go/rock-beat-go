package registry

import (
	"github.com/rock-go/rock/lua"
	"golang.org/x/sys/windows/registry"
)

type registryGo struct {
	name      string
	handleKey registry.Key
}

func newRegistryGo(name string) *registryGo {
	var key registry.Key
	switch name {
	case "USERS":
		key = registry.USERS
	case "LOCAL_MACHINE":
		key = registry.LOCAL_MACHINE
	case "CLASSES_ROOT":
		key = registry.CLASSES_ROOT
	case "CURRENT_CONFIG":
		key = registry.CURRENT_CONFIG
	case "CURRENT_USER":
		key = registry.CURRENT_USER
	case "PERFORMANCE_DATA":
		key = registry.PERFORMANCE_DATA
	}

	return &registryGo{name: name, handleKey: key}
}

func registryL(name string) lua.LValue {
	obj := newRegistryGo(name)
	ud := lua.NewAnyData(obj)

	ud.Meta("set", lua.NewFunction(obj.setL))
	ud.Meta("open", lua.NewFunction(obj.openL))
	ud.Meta("pipe" , lua.NewFunction(obj.pipeL))
	//ud.Meta("delete_key" , lua.NewFunction(obj.create))
	//ud.Meta("open_key"   , lua.NewFunction(obj.create))
	return ud
}

func CheckPerm(L *lua.LState, idx int) uint32 {
	perm := L.CheckString(idx)
	switch perm {
	case "ALL_ACCESS":
		return registry.ALL_ACCESS
	case "CREATE_LINK":
		return registry.CREATE_LINK
	case "CREATE_SUB_KEY":
		return registry.CREATE_SUB_KEY
	case "EXECUTE":
		return registry.EXECUTE
	case "NOTIFY":
		return registry.NOTIFY
	case "QUERY_VALUE":
		return registry.QUERY_VALUE
	case "READ":
		return registry.READ
	case "SET_VALUE":
		return registry.SET_VALUE
	case "WOW65_32KEY":
		return registry.WOW64_32KEY
	case "WOW65_64KEY":
		return registry.WOW64_64KEY
	case "WRITE":
		return registry.WRITE
	case "ENUMERATE_SUB_KEYS":
		return registry.ENUMERATE_SUB_KEYS

	default:
		L.RaiseError("invalid type perm")
		return 0
	}
}

func (r *registryGo) setL(L *lua.LState) int {
	name := L.CheckString(1)
	perm := CheckPerm(L, 2)

	key, exists, err := registry.CreateKey(r.handleKey, name, perm)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LFalse)
		L.Pushf("%v", err)
	} else {
		L.Push(newLuaKey(L, key, name, r))
		L.Push(lua.LBool(exists))
		L.Push(lua.LNil)
	}

	return 3

}

func (r *registryGo) openL(L *lua.LState) int {
	name := L.CheckString(1)
	key, err := registry.OpenKey(r.handleKey, name, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		L.Push(lua.LNil)
		L.Pushf("%v", err)
		return 2
	}

	L.Push(newLuaKey(L, key, name, r))
	return 1
}


