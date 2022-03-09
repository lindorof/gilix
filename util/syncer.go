package util

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SyncerWaitMode int

const (
	SyncerWaitModeCancel = 1
	SyncerWaitModeIdle   = 2
	SyncerWaitModeAny    = 3
)

type Syncer struct {
	parent *Syncer
	ctxs   context.CancelFunc

	wg    *sync.WaitGroup
	ctx   context.Context
	ctxc  context.CancelFunc
	once  *sync.Once
	onced chan bool
}

func newSyncer(ctx context.Context, parent *Syncer, ctxs context.CancelFunc) *Syncer {
	syncer := &Syncer{
		parent: parent,
		ctxs:   ctxs,
	}

	syncer.wg = &sync.WaitGroup{}
	syncer.ctx, syncer.ctxc = context.WithCancel(ctx)
	syncer.once = &sync.Once{}
	syncer.onced = make(chan bool, 1)

	return syncer
}

func (syncer *Syncer) increase() {
	if syncer.parent != nil {
		syncer.parent.increase()
	}
	syncer.wg.Add(1)
}

func (syncer *Syncer) decrease() {
	if syncer.parent != nil {
		syncer.parent.decrease()
	}
	syncer.once.Do(func() { syncer.onced <- true })
	syncer.wg.Done()
}

func CreateSyncerWithSig(ctx context.Context) *Syncer {
	sigs := []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTERM}
	ctx, ctxs := signal.NotifyContext(ctx, sigs...)

	return newSyncer(ctx, nil, ctxs)
}

func CreateSyncer(ctx context.Context) *Syncer {
	return newSyncer(ctx, nil, nil)
}

func CreateSyncerGroup(ctx context.Context, syncers ...**Syncer) {
	for _, syncer := range syncers {
		*syncer = newSyncer(ctx, nil, nil)
	}
}

func WaitReleaseSyncerGroup(act SyncerWaitMode, syncers ...*Syncer) {
	for _, syncer := range syncers {
		syncer.WaitRelease(act)
	}
}

func (syncer *Syncer) DeriveSyncer() *Syncer {
	return newSyncer(syncer.ctx, syncer, nil)
}

func (syncer *Syncer) WaitRelease(act SyncerWaitMode) {
	switch act {
	case SyncerWaitModeAny:
		<-syncer.onced
		syncer.ctxc()
	case SyncerWaitModeCancel:
		syncer.ctxc()
	case SyncerWaitModeIdle:
	}
	syncer.wg.Wait()

	syncer.fini()
}

func (syncer *Syncer) Async(fSync func(), fBreak func()) {
	syncer.increase()
	go func() {
		syncer.sync(fSync, fBreak)
		syncer.decrease()
	}()
}

func (syncer *Syncer) Sync(fSync func(), fBreak func()) {
	syncer.sync(fSync, fBreak)

	syncer.fini()
}

func (syncer *Syncer) sync(fSync func(), fBreak func()) {
	done := make(chan bool, 1)
	go func() {
		fSync()
		done <- true
	}()

	for {
		select {
		case <-done:
			return
		case <-syncer.ctx.Done():
			fBreak()
			<-done
			return
		}
	}
}

func (syncer *Syncer) fini() {
	syncer.ctxc()
	if syncer.ctxs != nil {
		syncer.ctxs()
	}
}

func (syncer *Syncer) Ctx() context.Context {
	return syncer.ctx
}
