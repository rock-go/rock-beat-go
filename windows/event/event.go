package event

import (
	"context"
	"github.com/rock-go/rock-beat-go/windows/event/watch"
	"github.com/rock-go/rock/audit"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/bucket"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/pipe"
	"time"
)

type winEv struct {
	lua.Super

	cfg     *config
	ctx     context.Context
	stop    context.CancelFunc
	watcher *watch.WinLogWatcher
}

func newWinEv(cfg *config) *winEv {
	w := &winEv{cfg: cfg}
	w.V(lua.INIT, winEvTypeOf)
	return w
}

func inPass(pass []uint64, id uint64) bool {
	for _, v := range pass {
		if v == id {
			return true
		}
	}

	return false
}

func (wv *winEv) bookmark(evt *watch.WinLogEvent) {
	if len(wv.cfg.bkt) == 0 {
		return
	}

	err := bucket.Pack(xEnv, wv.cfg.bkt...).Push(evt.Channel, auxlib.S2B(evt.Bookmark), 0)
	if err != nil {
		audit.NewEvent("win-log").
			Subject("bbolt db save fail").
			From(wv.cfg.co.CodeVM()).
			Msg("windows event log save last fail").
			E(err).Log().Put()
	}
}

func (wv *winEv) require(id uint64) pipe.Pipe {
	val := wv.cfg.chains.Get(auxlib.ToString(id))
	if val == lua.LNil || val == nil {
		return nil
	}

	return pipe.LFunc(val.(*lua.LFunction))
}

func (wv *winEv) call(evt *watch.WinLogEvent) {
	pv := wv.require(evt.EventId)
	if pv != nil {
		if e := pv(evt , wv.cfg.co) ; e != nil {
			xEnv.Errorf("%s event id %d pipe call fail %v" , wv.Name() , evt.EventId , e)
			return
		}
	}

	pipe.Do(wv.cfg.pipe , evt , wv.cfg.co , func(err error){
		xEnv.Errorf("%s event %s pipe call fail %v" , wv.Name() , evt.EventId , err)
	})
}

func (wv *winEv) send(evt *watch.WinLogEvent) {
	if wv.cfg.sdk == nil {
		return
	}
	_, err := wv.cfg.sdk.Write(evt.Bytes())
	if err != nil {
		logger.Errorf("transport write %v", err)
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
		case evt := <-wv.watcher.Event():
			delay = 0
			wv.bookmark(evt)
			wv.send(evt)

			if inPass(wv.cfg.pass, evt.EventId) {
				continue
			}
			wv.call(evt)
		case err := <-wv.watcher.Error():
			audit.NewEvent("beat-windows-log",
				audit.Subject("windows event log fail"),
				audit.From(wv.cfg.co.CodeVM()),
				audit.Msg("windows 系统日志获取失败"),
				audit.E(err)).Log().Put()

		default:
			delay += 10 * time.Millisecond
			if delay >= max {
				<-time.After(max)
			} else {
				<-time.After(delay)
			}
		}
	}
}

func (wv *winEv) Start() error {

	watcher, err := watch.New()
	if err != nil {
		return err
	}

	ctx, stop := context.WithCancel(context.Background())
	wv.ctx = ctx
	wv.stop = stop
	wv.watcher = watcher

	for _, item := range wv.cfg.channel {
		wv.subscribe(item.name, item.query)
	}

	xEnv.Spawn(0, wv.accpet)
	return nil
}

func (wv *winEv) Reload() error {
	errs := xEnv.NewERR()
	for _, name := range wv.watcher.Watches() {
		if !wv.inChannel(name) {
			errs.Raise(name, wv.watcher.RemoveSubscription(name))
		}
	}
	return errs.Wrap()
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
