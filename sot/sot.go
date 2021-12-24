package sot

import (
	"context"

	"gitee.com/lindorof/gilix"
	"gitee.com/lindorof/gilix/acp"
	"gitee.com/lindorof/gilix/util"
)

func init() {
	se := &sotEngine{}
	se.devr = createDevRep(se)
	se.breaker = make(chan bool, 1)
	util.CreateSyncerGroup(context.Background(), &se.acpSyncer, &se.devrSyncer, &se.devsSyncer)

	gilix.CPS = se
}

type sotEngine struct {
	devr       *devRep
	breaker    chan bool
	acpSyncer  *util.Syncer
	devrSyncer *util.Syncer
	devsSyncer *util.Syncer
}

func (se *sotEngine) SotLoopSync() {
	se.devrSyncer.Async(se.devr.loopSync, se.devr.loopBreak)
	<-se.breaker
	util.WaitReleaseSyncerGroup(util.SYNCER_WAIT_MODE_CANCEL, se.acpSyncer, se.devrSyncer, se.devsSyncer)
}

func (se *sotEngine) SotLoopBreak() {
	se.breaker <- true
}

func (se *sotEngine) SubmitAcp(a acp.Acceptor) {
	a.SetSeqHook(func(chw chan<- []byte) acp.Session {
		return newSession(se, a, chw)
	})
	se.acpSyncer.Async(a.LoopSync, a.LoopBreak)
}
