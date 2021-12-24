package sot

import (
	"container/list"
	"time"

	"gitee.com/lindorof/gilix"
)

type device struct {
	se    *sotEngine
	phy   string
	reqs  chan *sotReq
	seqs  map[*session]bool
	tasks *list.List
	pollb bool

	dev   gilix.Dev
	polli int
	polls []gilix.Callee
	pollc gilix.PollCache

	curivk  *invokeCur
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

	se.devsSyncer.Async(d.loopSync, func() { d.reqs <- nil })
	return d
}

func (d *device) loopSync() {
	d.dev = gilix.CBS.DevInit(d.phy, d)
	d.polli = d.dev.PollInterval()
	d.polls = d.dev.PollFuncs()
	d.pollc = d.poll()

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
	if d.curivk != nil {
		<-d.curivk.done
	}
	gilix.CBS.DevFini(d.dev)
}

func (d *device) onInvoke(req *sotReq) bool {
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
		if d.curivk != nil {
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
	if req.meta.ivx&invokeDF != 0 {
		d.curivk = nil
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
	ercv, ehsu := d.dev.OnEvt(d.pollc, req.evt)

	for seq := range d.seqs {
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
	if d.curivk != nil || d.tasks.Len() <= 0 {
		return
	}
	d.curivk = &invokeCur{
		req:  d.tasks.Front().Value.(*sotReq),
		done: make(chan bool, 1),
	}
	go func() {
		d.curivk.req.meta.ivk()

		d.curivk.req.tpr()
		d.putq(d.curivk.req)

		d.curivk.done <- true
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
