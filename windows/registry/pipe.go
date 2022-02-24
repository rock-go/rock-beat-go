package registry

import (
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/pipe"
	"golang.org/x/sys/windows/registry"
)

type pv struct {
	prefix string
	name   string
	vd     []byte
	vt     uint32
}

func (p pv) ToLValue() lua.LValue {
	return lua.NewAnyData(p)
}

func (p pv) Index(L *lua.LState , key string) lua.LValue {
	switch key {
	case "path":
		return lua.S2L(p.prefix + "\\" + p.name)

	case "prefix":
		return lua.S2L(p.prefix)

	case "root":
		return lua.S2L(p.prefix)

	case "name":
		return lua.S2L(p.name)

	case "value":
		return convert(L , p.vd , p.vt)

	}

	return lua.LNil
}

func (r *registryGo) onReadValue(L *lua.LState , prefix string ,
	root registry.Key , info *registry.KeyInfo ,  pp []pipe.Pipe) {

	defer root.Close()

	if info.ValueCount <= 0 {
		return
	}

	names , err := root.ReadValueNames(int(info.ValueCount))
	if err != nil {
		xEnv.Errorf("read stat fail %v" , err)
		return
	}

	buff := make([]byte , info.MaxValueLen)
	lv := pv{}
	for _ , name := range names {
		n , vt , e2 := root.GetValue(name , buff)
		if e2 != nil {
			xEnv.Errorf("%s got %s value error %v" , r.name , name , e2)
			continue
		}

		lv.vt     = vt
		lv.vd     = buff[:n]
		lv.name   = name
		lv.prefix = prefix

		pipe.Do(pp , lv , L , func(err error) {
			xEnv.Errorf("%s pipe %v" , r.name , err)
		})

	}
}

func (r *registryGo) onReadSubValue(L *lua.LState , prefix string ,
	root registry.Key , info *registry.KeyInfo , pp []pipe.Pipe) {

	if info.SubKeyCount <= 0 {
		return
	}

	names , err := root.ReadSubKeyNames(int(info.SubKeyCount))
	if err != nil {
		xEnv.Errorf("%s pipe sub value fail %v" , r.name , err)
		return
	}

	perm := uint32(registry.QUERY_VALUE | registry.ENUMERATE_SUB_KEYS)
	for _ , name := range names {
		path := prefix + "\\" + name
		root2 , e2 := registry.OpenKey(r.handleKey , path , perm)
		if e2 != nil {
			xEnv.Errorf("%s pipe sub %v" , r.name , e2)
			continue
		}

		i2 , e2 := root2.Stat()
		if e2 != nil {
			xEnv.Errorf("%s pipe sub stat fail %v" , r.name , e2)
			continue
		}

		r.onReadValue(L , path , root2 , i2 , pp)
	}
}
func (r *registryGo) pipeL(L *lua.LState) int {
	name := L.CheckString(1)
	pp := pipe.Check(L)
	if len(pp) == 0 {
		return 0
	}

	root, err := registry.OpenKey(r.handleKey, name, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		L.Push(err)
		return 1
	}

	info , err := root.Stat()
	if err != nil {
		L.Push(err)
		return 1
	}

	prefix := r.name + "\\" + name
	r.onReadValue(L , prefix , root , info , pp)
	r.onReadSubValue(L , prefix , root , info , pp)
	return 0
}
