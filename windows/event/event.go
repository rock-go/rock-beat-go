package event

import (
	"context"
	"fmt"
	"github.com/rock-go/rock-beat-go/windows/event/watch"
	"github.com/rock-go/rock/audit"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/node"
	"github.com/rock-go/rock/thread"
	"github.com/rock-go/rock/xcall"
	"time"
)


type winEv struct {
	lua.Super

	cfg       *config
	ctx       context.Context
	stop      context.CancelFunc
	watcher   *watch.WinLogWatcher

	channel   []string
	vm        *lua.LState
}

func newWinEv(cfg *config) *winEv {
	w := &winEv{ cfg: cfg	}
	w.V(lua.INIT , time.Now())
	return w
}

func inPass(pass []uint64, id uint64) bool {
	for _ , v := range pass {
		if v == id {
			return true
		}
	}

	return false
}

func (wv *winEv) bookmark(evt *watch.WinLogEvent) {
	key := fmt.Sprintf("%s_last" , evt.SubscribedChannel)

	err := node.Put(winEvBucket, []byte(key) , lua.S2B(evt.Bookmark))
	if err != nil {
		audit.NewEvent("win-log" ,
			audit.Subject("bbolt db save fail"),
			audit.From(wv.vm.CodeVM()),
			audit.Msg("windows event log save last fail"),
			audit.E(err)).Log().Put()
	}
}

func (wv *winEv) require(id uint64) *lua.LFunction {
	val := wv.cfg.chains.Get(auxlib.ToString(id))
	if val == lua.LNil || val == nil {
		return wv.cfg.hook
	}

	return val.(*lua.LFunction)
}

func (wv *winEv) call(evt *watch.WinLogEvent) {
	fn := wv.require(evt.EventId)
	if fn == nil {
		return
	}

	co := lua.Clone(wv.vm)
	err := xcall.CallByParam(co , lua.P{
		Fn: fn,
		NRet: 0,
		Protect: true,
	} , xcall.Rock , evt.ToLValue(co))

	lua.FreeState(co)

	if err != nil {
		logger.Errorf("%s call ev_%d error %v" , wv.vm.CodeVM() ,evt.EventId , err)
	}
}

func (wv *winEv) send(evt *watch.WinLogEvent) {
	if wv.cfg.transport == nil {
		return
	}
	_ , err := wv.cfg.transport.Write(evt.Bytes())
	if err != nil {
		logger.Errorf("transport write %v" , err)
		return
	}
}

func (wv *winEv) accpet() {

	delay := 50 * time.Millisecond
	max := 1 * time.Second

	for {
		select {

		case <-wv.ctx.Done():
			return
		case evt := <- wv.watcher.Event():
			delay = 0
			wv.bookmark(evt)
			wv.send(evt)

			if inPass(wv.cfg.pass , evt.EventId) {
				continue
			}
			wv.call(evt)
		case err := <-wv.watcher.Error():
			audit.NewEvent("beat-windows-log",
					audit.Subject("windows event log fail"),
				audit.From(wv.vm.CodeVM()),
				audit.Msg("windows 系统日志获取失败"),
				audit.E(err)).Log().Put()

		default:
			delay += 10 * time.Millisecond
			if delay >= max {
				<- time.After(max)
			} else {
				<-time.After(delay)
			}
		}
	}
}

func (wv *winEv) Start() error {

	watcher , err := watch.New()
	if err != nil {
		return err
	}

	ctx , stop := context.WithCancel(context.Background())

	wv.ctx = ctx
	wv.stop = stop
	wv.watcher = watcher
	thread.Spawn(0 , wv.accpet)
	return nil
}

func (wv *winEv) Reload() error {
	for _ , ch := range  wv.channel {
		wv.watcher.RemoveSubscription(ch)
	}
	return nil
}


func (wv *winEv) Close() error {
	wv.stop()
	wv.watcher.Shutdown()
	return nil
}

func (wv *winEv) Name() string {
	return wv.cfg.name
}

func (wv *winEv) Type() string {
	return winEvTypeOf
}