//go:build windows
// +build windows

package wmi

import (
	"github.com/go-ole/go-ole"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/xreflect"
	"time"
	"unsafe"
)

type ant struct {
	err  error
	data *ole.VARIANT
}

func newAnt(v *ole.VARIANT) *ant {
	return &ant{data: v}
}

func (a *ant) toLV(L *lua.LState) *lua.AnyData {
	ud := L.NewAnyData(a)
	ud.Meta("s", lua.NewFunction(a.s))
	ud.Meta("n", lua.NewFunction(a.n))
	ud.Meta("b", lua.NewFunction(a.b))
	ud.Meta("t", lua.NewFunction(a.t))
	return ud
}

func (a *ant) toArr(L *lua.LState, item *ole.VARIANT) *lua.LTable {

	safeArr := item.ToArray()
	if safeArr == nil {
		return L.CreateTable(0, 0)
	} else {
		arr := safeArr.ToValueArray()
		n := len(arr)
		tab := L.CreateTable(n, 0)
		for i := 0; i < n; i++ {
			tab.RawSetInt(i, xreflect.ToLValue(arr[i], L))
		}
		return tab
	}
}

func (a *ant) s(L *lua.LState) int {
	key := L.CheckString(1)
	item, err := a.data.ToIDispatch().GetProperty(key)
	if err == nil {
		switch item.VT {
		case ole.VT_BSTR:
			L.Push(lua.S2L(item.ToString()))
			return 1
		case ole.VT_DATE, ole.VT_NULL:
			date, _ := ole.GetVariantDate(uint64(item.Val))
			L.Push(lua.S2L(date.Format(time.RFC3339Nano)))
			return 1
		}
	}

	L.Push(lua.LString(""))
	return 1
}

func (a *ant) n(L *lua.LState) int {
	key := L.CheckString(1)
	item, err := a.data.ToIDispatch().GetProperty(key)
	if err != nil {
		L.Push(lua.LNumber(0))
		return 0
	}
	var lv lua.LValue

	switch item.VT {
	case ole.VT_I1:
		lv = lua.LInt(int8(item.Val))
	case ole.VT_UI1:
		lv = lua.LInt(uint8(item.Val))
	case ole.VT_I2:
		lv = lua.LInt(int16(item.Val))
	case ole.VT_UI2:
		lv = lua.LInt(uint16(item.Val))
	case ole.VT_I4:
		lv = lua.LInt(int32(item.Val))
	case ole.VT_UI4:
		lv = lua.LNumber(uint32(item.Val))
	case ole.VT_I8:
		lv = lua.LNumber(int64(item.Val))
	case ole.VT_UI8:
		lv = lua.LNumber(uint64(item.Val))
	case ole.VT_INT:
		lv = lua.LInt(int(item.Val))
	case ole.VT_UINT:
		lv = lua.LNumber(item.Val)
	case ole.VT_R4:
		lv = lua.LNumber(*(*float32)(unsafe.Pointer(&item.Val)))
	case ole.VT_R8:
		lv = lua.LNumber(*(*float64)(unsafe.Pointer(&item.Val)))
	default:
		lv = lua.LNumber(0)

	}

	L.Push(lv)
	return 1
}

func (a *ant) b(L *lua.LState) int {
	key := L.CheckString(1)
	item, err := a.data.ToIDispatch().GetProperty(key)
	if err == nil {
		if item.VT == ole.VT_BOOL {
			L.Push(lua.LBool(!(int(item.Val) == 0)))
			return 1
		}
	}

	L.Push(lua.LBool(false))
	return 1
}

func (a *ant) t(L *lua.LState) int {
	key := L.CheckString(1)

	item, err := a.data.ToIDispatch().GetProperty(key)
	if err == nil {
		if item.VT == ole.VT_ARRAY {
			L.Push(a.toArr(L, item))
			return 1
		}
	}

	L.Push(L.CreateTable(0, 0))
	return 1
}

func (a *ant) Index(L *lua.LState, key string) lua.LValue {
	item, err := a.data.ToIDispatch().GetProperty(key)
	if err != nil {
		return lua.LNil
	}
	var lv lua.LValue

	switch item.VT {
	case ole.VT_ARRAY:
		lv = a.toArr(L, item)
	default:
		lv = xreflect.ToLValue(item.Value(), L)
	}
	return lv
}
