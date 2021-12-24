package util

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SYNCER_WAIT_MODE int

const (
	SYNCER_WAIT_MODE_CANCEL = 1
	SYNCER_WAIT_MODE_IDLE   = 2
	SYNCER_WAIT_MODE_ANY    = 3
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

func WaitReleaseSyncerGroup(act SYNCER_WAIT_MODE, syncers ...*Syncer) {
	for _, syncer := range syncers {
		syncer.WaitRelease(act)
	}
}

func (syncer *Syncer) DeriveSyncer() *Syncer {
	return newSyncer(syncer.ctx, syncer, nil)
}

func (syncer *Syncer) WaitRelease(act SYNCER_WAIT_MODE) {
	switch act {
	case SYNCER_WAIT_MODE_CANCEL:
		syncer.ctxc()
	case SYNCER_WAIT_MODE_IDLE:
		<-syncer.ctx.Done()
	case SYNCER_WAIT_MODE_ANY:
		<-syncer.onced
		syncer.ctxc()
	default:
	}

	syncer.wg.Wait()
	if syncer.ctxs != nil {
		syncer.ctxs()
	}
}

func (syncer *Syncer) Async(fSync func(), fBreak func()) {
	syncer.increase()
	go func() {
		syncer.Sync(fSync, fBreak)
		syncer.decrease()
	}()
}

func (syncer *Syncer) Sync(fSync func(), fBreak func()) {
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
