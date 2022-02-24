package event

import (
	"github.com/rock-go/rock/audit"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/bucket"
)

func (wv *winEv) subscribe(name, query string) {

	var err error
	if wv.cfg.begin {
		wv.watcher.SubscribeFromBeginning(name, query)
		return
	}

	bookmark, err := bucket.Pack(xEnv, winEvBucketOffset).Value(name)
	if err != nil {
		wv.watcher.SubscribeFromBeginning(name, query)
		return
	}

	audit.NewEvent("win-log").
		Subject("%s last bookmark", name).
		From(wv.cfg.co.CodeVM()).
		Msg("%s", bookmark).Log().Put()

	wv.watcher.SubscribeFromBookmark(name, query, auxlib.B2S(bookmark))
}

func (wv *winEv) inChannel(name string) bool {
	for _, item := range wv.cfg.channel {
		if item.name == name {
			return true
		}
	}
	return false
}

