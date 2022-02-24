package registry

import (
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"golang.org/x/sys/windows/registry"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

func convert(L *lua.LState, data []byte, vt uint32) lua.LValue {
	switch vt {
	case registry.NONE:
		return lua.LSNull

	case registry.SZ, registry.EXPAND_SZ:
		if len(data) == 0 {
			return lua.LSNull
		} else {
			u := (*[1 << 29]uint16)(unsafe.Pointer(&data[0]))[: len(data)/2 : len(data)/2]
			v := syscall.UTF16ToString(u)
			return lua.S2L(v)
		}

	case registry.QWORD:
		if len(data) != 4 {
			return lua.LNumber(0)
		} else {
			var val32 uint32
			copy((*[4]byte)(unsafe.Pointer(&val32))[:], data)
			return lua.LNumber(val32)
		}

	case registry.DWORD:
		if len(data) != 8 {
			return lua.LNumber(0)
		} else {
			var val64 uint64
			copy((*[8]byte)(unsafe.Pointer(&val64))[:], data)
			return lua.LNumber(val64)
		}

	case registry.BINARY:
		n := len(data)
		tab := L.CreateTable(n, 0)
		for i := 0; i < n; i++ {
			tab.RawSetInt(i+1, lua.LNumber(data[i]))
		}
		return tab

	case registry.MULTI_SZ:
		tab := L.CreateTable(0, 0)
		if len(data) == 0 {
			return tab
		}

		p := (*[1 << 29]uint16)(unsafe.Pointer(&data[0]))[: len(data)/2 : len(data)/2]
		if len(p) == 0 {
			return tab
		}
		if p[len(p)-1] == 0 {
			p = p[:len(p)-1] // remove terminating null
		}

		k := 1
		from := 0
		for i, c := range p {
			if c == 0 {
				tab.RawSetInt(k, lua.S2L(string(utf16.Decode(p[from:i]))))
				k++
				from = i + 1
			}
		}
		return tab

	default:
		return lua.LSNull
	}
}

func toTab(L *lua.LState, rk *rKey, tab *lua.LTable) {
	stat, err := rk.value.Stat()
	if err != nil {
		logger.Errorf("got registry %s stat error %v", rk.name, err)
		return
	}

	if stat.ValueCount != 0 {
		names, e := rk.value.ReadValueNames(int(stat.ValueCount))
		if e != nil {
			logger.Errorf("got registry %s names error %v", rk.name, e)
		} else {
			data := make([]byte, stat.MaxValueLen)
			for _, key := range names {
				path := rk.r.name + "\\" + rk.name + "\\" + key

				_, vt, e2 := rk.value.GetValue(key, data)
				if e2 != nil {
					logger.Errorf("%s got value error %v", path, e2)
					continue
				}
				tab.RawSetString(path, convert(L, data, vt))
			}
		}

	}

	if stat.SubKeyCount != 0 {
		names, e := rk.value.ReadSubKeyNames(int(stat.SubKeyCount))
		if e != nil {
			logger.Errorf("got registry %s sub names error %v", rk.name, e)
			return
		}

		for _, name := range names {
			path := rk.name + "\\" + name
			perm := registry.QUERY_VALUE | registry.ENUMERATE_SUB_KEYS
			item, e2 := registry.OpenKey(rk.r.handleKey, path, uint32(perm))
			if e2 != nil {
				logger.Errorf("got registry %s\\%s open error %v", rk.r.name, path, e2)
				continue
			}

			rk2 := &rKey{
				r:     rk.r,
				name:  path,
				value: item,
			}

			toTab(L, rk2, tab)
			item.Close()
		}
	}

}
