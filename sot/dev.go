package sot

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/lindorof/gilix"
	"github.com/lindorof/gilix/util"
)

type device struct {
	se    *sotEngine
	phy   string
	reqs  chan *sotReq
	seqs  map[*session]bool
	tasks *list.List
	pollb bool

	zapt *util.Zapt

	dev   gilix.Dev
	polli int
	polls []gilix.Callee
	pollc gilix.PollCache

	currwg  sync.WaitGroup
	curreq  *sotReq
	curlck  *session
	curlcki gilix.ID
}

func createDev(se *sotEngine, phy string) *device {
	d := &device{
		se:    se,
		phy:   phy,
		reqs:  make(chan *sotReq, 1025),
		seqs:  make(map[*session]bool, 256),
		tasks: list.New(),
		pollb: true,
	}

	mod := fmt.Sprintf("sotDev-%s", phy)
	d.zapt = util.ZaptByCfg(0, mod, phy)

	se.devsSyncer.Async(d.loopSync, func() { d.reqs <- nil })
	return d
}

func (d *device) loopSync() {
	d.dev = gilix.CBS.DevInit(d.phy, d)
	d.polli = d.dev.PollInterval()
	d.polls = d.dev.PollFuncs()
	d.pollc = d.poll()

	d.zapt.Infof("dev=%p , polli=%d", d.dev, d.polli)

	var ticker *time.Ticker = nil
	var pollt <-chan time.Time = nil
	if d.polli > 0 {
		ticker = time.NewTicker(time.Millisecond * time.Duration(d.polli))
		pollt = ticker.C
	}

LOOP:
	for {
		select {
		case <-pollt:
			if d.pollb {
				d.putq(newReq(srtPollStart, nil, nil, nil))
			}
		case req := <-d.reqs:
			if req == nil {
				break LOOP
			} else if req.tp == srtSeqCreate {
				d.seqs[req.seq] = true
			} else if req.tp == srtSeqClean {
				delete(d.seqs, req.seq)
				d.taskc(req.seq, gilix.ID_NIL)
				d.unlock(req.seq)
			} else if req.tp == srtReqStart || req.tp == srtPollStart {
				if d.onInvoke(req) {
					d.onInvokeRet(req)
				}
			} else if req.tp == srtReqComplete || req.tp == srtPollComplete {
				d.onInvokeRet(req)
			} else if req.tp == srtEvtPost {
				d.onEvtPost(req)
			}
		}
	}

	if ticker != nil {
		ticker.Stop()
	}
	d.currwg.Wait()
	gilix.CBS.DevFini(d.dev)
}

func (d *device) onInvoke(req *sotReq) bool {
	if req.tp != srtPollStart {
		d.zapt.Infof("hs=%d,rid=%d,type=%d,code=%d,timeout=%d",
			req.req.Hs(), req.req.Id(),
			req.req.Type(), req.req.Code(),
			req.req.Timeout())
	}

	d.invokeMeta(req)

	if req.meta.ivx&invokeCI != 0 {
		d.taskc(req.seq, req.req.Id())
	}
	if req.meta.ivx&invokeCN != 0 {
		d.taskc(req.seq, gilix.ID_NIL)
	}
	if req.meta.ivx&invokeRT != 0 {
		req.meta.ivk()
		return true
	}
	if req.meta.ivx&invokePL != 0 {
		if d.curreq != nil {
			return false
		}
	}
	if req.meta.ivx&invokeDF != 0 {
		d.tasks.PushBack(req)
		d.taskg()
		return false
	}

	return false
}

func (d *device) onInvokeRet(req *sotReq) {
	if req.tp != srtPollComplete {
		rret := 0
		if req.meta.rsp != nil {
			rret = int(req.meta.rsp.Ret())
		}
		d.zapt.Infof("hs=%d,rid=%d,type=%d,code=%d,ret=%d,rspret=%d",
			req.req.Hs(), req.req.Id(),
			req.req.Type(), req.req.Code(),
			req.meta.ret, rret)
	}

	if req.meta.ivx&invokeDF != 0 {
		d.curreq = nil
		d.tasks.Remove(d.tasks.Front())
		d.taskg()
	}
	if req.meta.ivx&invokePL == 0 {
		msg := gilix.CBS.MsgByRsp(req.req, req.meta.ret, req.meta.rsp)
		req.seq.putw(msg)
	}
	if len(req.meta.pc) > 0 {
		if req.meta.chg && len(d.pollc) > 0 {
			d.dev.PollChange(d.pollc, req.meta.pc)
		}
		d.pollc = req.meta.pc
	}
}

func (d *device) onEvtPost(req *sotReq) {
	d.zapt.Infof("type=%d,code=%d", req.evt.Type(), req.evt.Code())

	ercv, ehsu := d.dev.OnEvt(d.pollc, req.evt)

	for seq := range d.seqs {
		sent := d.dev.EvtFilter(seq.usr, req.evt)
		if sent == false {
			continue
		}

		erhi := d.erhi(ercv, ehsu, seq)
		if erhi.rcv == nil || erhi.hs == gilix.HS_NIL {
			continue
		}

		msgs := gilix.CBS.MsgByEvt(erhi, erhi.rcv.usr, req.evt)
		for _, msg := range msgs {
			erhi.rcv.putw(msg)
		}
	}
}

func (d *device) taskg() {
	if d.curreq != nil || d.tasks.Len() <= 0 {
		return
	}
	d.curreq = d.tasks.Front().Value.(*sotReq)
	go func() {
		d.currwg.Add(1)

		d.curreq.meta.ivk()
		d.curreq.tpr()
		d.putq(d.curreq)

		d.currwg.Done()
	}()
}

func (d *device) taskc(seq *session, id gilix.ID) {
	for e := d.tasks.Front(); e != nil; e = e.Next() {
		task := e.Value.(*sotReq)
		if seq != task.seq {
			continue
		}
		if id == gilix.ID_NIL {
			task.ctxc()
			continue
		}
		if id == task.req.Id() {
			task.ctxc()
			return
		}
	}
}

func (d *device) putq(req *sotReq) {
	d.reqs <- req
}

func (d *device) PostEvt(evt gilix.Evt) {
	d.putq(newReq(srtEvtPost, nil, nil, evt))
}

func (d *device) PollSwitch(w bool) {
	d.pollb = w
}
