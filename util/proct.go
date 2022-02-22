package util

import (
	"context"
	"os/exec"
	"sync"
)

type ProctCb func(*ProctCmd)

type ProctCmd struct {
	C *exec.Cmd
	E error
	f ProctCb
}

type Proct struct {
	wg   *sync.WaitGroup
	ctx  context.Context
	cbs  chan *ProctCmd
	done chan bool
	once sync.Once
}

func NewProct(ctx context.Context) *Proct {
	proct := &Proct{
		wg:   &sync.WaitGroup{},
		ctx:  ctx,
		cbs:  make(chan *ProctCmd, 1024),
		done: make(chan bool, 1),
	}

	go func() {
		for cmd := range proct.cbs {
			if cmd == nil {
				break
			}
			if cmd.f != nil {
				cmd.f(cmd)
			}
		}
		proct.done <- true
	}()

	return proct
}

func (p *Proct) AddCmd(f ProctCb, path string, args ...string) {
	cmd := &ProctCmd{
		C: exec.CommandContext(p.ctx, path, args...),
		E: nil,
		f: f,
	}

	if cmd.E = cmd.C.Start(); cmd.E != nil {
		p.cbs <- cmd
		return
	}

	p.wg.Add(1)

	go func() {
		cmd.E = cmd.C.Wait()
		p.cbs <- cmd

		p.wg.Done()
	}()
}

func (p *Proct) Wait() {
	p.wg.Wait()

	p.once.Do(func() {
		p.cbs <- nil
		<-p.done
	})
}
