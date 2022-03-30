package sot

import (
	"context"

	"github.com/lindorof/gilix"
	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"
)

func init() {
	se := &sotEngine{}
	se.devr = createDevRep(se)
	se.breaker = make(chan bool, 1)
	util.CreateSyncerGroup(context.Background(), &se.acpSyncer, &se.devrSyncer, &se.devsSyncer)

	se.zapt = util.ZaptByCfg("gilix/sotEngine", "sot")

	gilix.CPS = se
}

type sotEngine struct {
	devr       *devRep
	breaker    chan bool
	acpSyncer  *util.Syncer
	devrSyncer *util.Syncer
	devsSyncer *util.Syncer

	zapt *util.Zapt
}

func (se *sotEngine) SotLoopSync() {
	se.zapt.Infof("entry")
	se.devrSyncer.Async(se.devr.loopSync, se.devr.loopBreak)
	<-se.breaker
	se.zapt.Infof("se.devrSyncer.Async end")

	util.WaitReleaseSyncerGroup(util.SyncerWaitModeCancel, se.acpSyncer, se.devrSyncer, se.devsSyncer)
	se.zapt.Infof("WaitReleaseSyncerGroup end")
}

func (se *sotEngine) SotLoopBreak() {
	se.breaker <- true
}

func (se *sotEngine) SubmitAcp(a acp.Acceptor) {
	se.zapt.Infof("%p", a)

	a.SetSeqHook(func(chw chan<- []byte) acp.Session {
		return newSession(se, a, chw)
	})
	se.acpSyncer.Async(a.LoopSync, a.LoopBreak)
}
