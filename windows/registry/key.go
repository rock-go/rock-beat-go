package registry

import (
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xreflect"
	"golang.org/x/sys/windows/registry"
)

type rKey struct {
	name  string
	r     *registryGo
	value registry.Key
}

func newLuaKey(L *lua.LState, key registry.Key, name string, r *registryGo) lua.LValue {
	rk := &rKey{value: key, name: name, r: r}
	return L.NewAnyData(rk)
}

func (rk *rKey) LClose(L *lua.LState) int {
	rk.value.Close()
	return 0
}

func (rk *rKey) e(L *lua.LState, err error) int {
	L.Push(lua.LNil)
	L.Pushf("%v", err)
	return 2
}

func (rk *rKey) s(L *lua.LState) int {
	key := L.CheckString(1)
	data, _, err := rk.value.GetStringValue(key)
	if err != nil {
		return rk.e(L, err)
	}

	L.Push(lua.S2L(data))
	return 1
}

func (rk *rKey) n(L *lua.LState) int {
	key := L.CheckString(1)
	data, _, err := rk.value.GetIntegerValue(key)
	if err != nil {
		return rk.e(L, err)
	}

	L.Push(lua.LNumber(data))
	return 1
}

func (rk *rKey) b(L *lua.LState) int {
	key := L.CheckString(1)
	data, _, err := rk.value.GetBinaryValue(key)
	if err != nil {
		return rk.e(L, err)
	}

	L.Push(lua.B2L(data))
	return 1
}

func (rk *rKey) LSetQwordValue(L *lua.LState) int {
	name := L.CheckString(1)
	val := L.CheckNumber(2)
	if err := rk.value.SetQWordValue(name, uint64(val)); err != nil {
		L.Pushf("%v", err)
		return 1
	}
	return 0
}

func (rk *rKey) LSetDwordValue(L *lua.LState) int {
	name := L.CheckString(1)
	val := L.CheckNumber(2)
	if err := rk.value.SetDWordValue(name, uint32(val)); err != nil {
		L.Pushf("%v", err)
		return 1
	}
	return 0
}

func (rk *rKey) LSetMulString(L *lua.LState) int {
	key := L.CheckString(1)
	tab := L.CheckTable(2)
	var arr []string
	auxlib.LTab2S(tab, lua.LTString, func(lv lua.LValue) {
		arr = append(arr, lv.String())
	})
	rk.value.SetStringsValue(key, arr)
	return 0
}

func (rk *rKey) LSetString(L *lua.LState) int {
	key := L.CheckString(1)
	val := L.CheckString(2)
	rk.value.SetStringValue(key, val)
	return 0
}

func (rk *rKey) LCountHelper(L *lua.LState, sub bool) int {
	stat, err := rk.value.Stat()
	if err != nil {
		L.Push(lua.LNumber(0))
		L.Pushf("%v")
		return 2
	}

	if sub {
		L.Push(lua.LNumber(stat.SubKeyCount))
	} else {
		L.Push(lua.LNumber(stat.ValueCount))
	}

	return 1
}

func (rk *rKey) subCountL(L *lua.LState) int {
	return rk.LCountHelper(L, true)
}

func (rk *rKey) MaxValueLen() uint32 {
	stat, err := rk.value.Stat()
	if err != nil {
		logger.Errorf("registry %s\\%s stat error %v", rk.r.name, rk.name, err)
		return 64
	}
	return stat.MaxValueLen
}

func (rk *rKey) LKeysHelper(L *lua.LState, sub bool) int {
	var n int
	var err error
	var names []string

	if L.GetTop() == 0 {
		stat, err := rk.value.Stat()
		if err != nil {
			return rk.e(L, err)
		}

		if sub {
			n = int(stat.SubKeyCount)
		} else {
			n = int(stat.ValueCount)
		}

	} else {
		n = L.CheckInt(1)
	}

	if sub {
		names, err = rk.value.ReadSubKeyNames(n)
	} else {
		names, err = rk.value.ReadValueNames(n)
	}

	if err != nil {
		L.Push(lua.LNil)
		L.Pushf("%v", err)
		return 2
	}

	lv := xreflect.ToLValue(names, L)
	L.Push(lv)
	return 1
}

func (rk *rKey) subKeysL(L *lua.LState) int {
	return rk.LKeysHelper(L, true)
}

func (rk *rKey) keyL(L *lua.LState) int {
	return rk.LKeysHelper(L, false)
}

func (rk *rKey) countL(L *lua.LState) int {
	return rk.LCountHelper(L, false)
}

func (rk *rKey) index(L *lua.LState) int {
	name := L.CheckString(1)
	data := make([]byte, rk.MaxValueLen())
	_, vt, err := rk.value.GetValue(name, data)
	if err != nil {
		return rk.e(L, err)
	}

	lv := convert(L, data, vt)
	L.Push(lv)
	return 1
}

func (rk *rKey) mapL(L *lua.LState) int {

	tab := L.CreateTable(0, 0)
	toTab(L, rk, tab)
	L.Push(tab)
	return 1
}

func (rk *rKey) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "s":
		return L.NewFunction(rk.s)
	case "n":
		return L.NewFunction(rk.n)
	case "b":
		return L.NewFunction(rk.b)
	case "get":
		return L.NewFunction(rk.index)

	case "sub_keys":
		return L.NewFunction(rk.subKeysL)
	case "sub_count":
		return L.NewFunction(rk.subCountL)

	case "keys":
		return L.NewFunction(rk.keyL)
	case "count":
		return L.NewFunction(rk.countL)

	case "map":
		return L.NewFunction(rk.mapL)

	case "close":
		return L.NewFunction(rk.LClose)

	case "set_qword":
		return L.NewFunction(rk.LSetQwordValue)

	case "set_dword":
		return L.NewFunction(rk.LSetDwordValue)

	case "strings":
		return L.NewFunction(rk.LSetMulString)

	case "expand":
		return L.NewFunction(rk.LSetString)

	default:
		return lua.LNil
	}
}

func (rk *rKey) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch val.Type() {

	case lua.LTString:
		rk.value.SetStringValue(key, val.String())
	case lua.LTNumber:
		rk.value.SetDWordValue(key, uint32(val.(lua.LNumber)))
	case lua.LTTable:
		arr := val.(*lua.LTable).Array()
		n := len(arr)
		if n == 0 {
			rk.value.SetBinaryValue(key, []byte{})
			return
		}

		var ret []byte
		for i := 0; i < n; i++ {
			item := arr[i]
			switch item.Type() {
			case lua.LTNumber:
				v := uint32(item.(lua.LNumber))
				if v < 256 && v > 0 {
					ret = append(ret, uint8(v))
				}

			case lua.LTInt:
				v := uint32(item.(lua.LInt))
				if v < 256 && v > 0 {
					ret = append(ret, uint8(v))
				}

			default:
			}
		}
		rk.value.SetBinaryValue(key, ret)
	}
}
