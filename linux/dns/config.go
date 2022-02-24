package dns

import (
	"fmt"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/pipe"
	"github.com/rock-go/rock/region"
)

//udp://0.0.0.0/?port=53,5353
type config struct {
	name   string
	region *region.Region
	bind   auxlib.URL
	pipe   []pipe.Pipe
	co     *lua.LState
}

func newConfig(L *lua.LState) *config {
	val := L.Get(1)
	cfg := &config{co: xEnv.Clone(L)}

	switch val.Type() {
	case lua.LTString:
		cfg.name = val.String()

	case lua.LTTable:
		val.(*lua.LTable).Range(func(key string, val lua.LValue) {
			switch key {
			case "name":
				cfg.name = val.String()

			case "bind":
				cfg.bind = auxlib.CheckURL(val, L)

			case "region":
				cfg.region = region.CheckRegionSdk(L, val)
			}
		})

	default:
		L.RaiseError("invalid config type %v", val.Type().String())
	}

	return cfg
}

func (cfg *config) net() string {
	if cfg.bind.V6() {
		return "ip6:udp"
	}

	return "ip4:udp"
}

func (cfg *config) valid() error {
	if e := auxlib.Name(cfg.name); e != nil {
		return e
	}

	if cfg.bind.IsNil() {
		return fmt.Errorf("not found bind")
	}

	if !cfg.bind.V4() && !cfg.bind.V6() {
		return fmt.Errorf("not found hostname must ipv4 or ipv6")
	}

	if cfg.bind.Port() == 0 && len(cfg.bind.Ports()) == 0 {
		return fmt.Errorf("not found listen port")
	}

	switch cfg.bind.Scheme() {
	case "udp", "tcp":
		//todo
	default:
		return fmt.Errorf("not found listen %s", cfg.bind.Scheme())
	}

	return nil
}
