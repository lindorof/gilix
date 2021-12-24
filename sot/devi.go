package sot

import (
	"context"
	"time"

	"github.com/lindorof/gilix"
	"github.com/lindorof/gilix/util"
)

const (
	invokeCI = 0x000001
	invokeCN = 0x000010
	invokeRT = 0x000100
	invokePL = 0x001000
	invokeDF = 0x010000
)

type invokeCur struct {
	req  *sotReq
	done chan bool
}

type invokeMeta struct {
	ivx int
	ivk func()

	ret gilix.RET
	rsp gilix.Rsp
	pc  gilix.PollCache
	chg bool
}

func (d *device) invokeMeta(req *sotReq) {
	req.meta = &invokeMeta{}

	if req.tp == srtPollStart {
		req.meta.ivx = invokePL | invokeDF
		req.meta.ivk = func() { d.invoke_poll(req) }
		return
	}

	qut, cee, pci, chk, chg := d.dev.OnReq(req.req, req.seq.usr)

	switch req.req.Type() {
	case gilix.TYPE_INF:
		if qut == gilix.QUEUET_DF {
			req.meta.ivx = invokeDF
		} else {
			req.meta.ivx = invokeRT
		}
		req.meta.ivk = func() { d.invoke_inf(req, cee, pci) }
	case gilix.TYPE_CMD:
		if qut == gilix.QUEUET_RT {
			req.meta.ivx = invokeRT
		} else {
			req.meta.ivx = invokeDF
		}
		req.meta.ivk = func() { d.invoke_cmd(req, cee, chk, chg) }
	case gilix.TYPE_OPEN:
		req.meta.ivx = invokeRT
		req.meta.ivk = func() { d.invoke_open(req) }
	case gilix.TYPE_REG:
		req.meta.ivx = invokeRT
		req.meta.ivk = func() { d.invoke_reg(req) }
	case gilix.TYPE_DEREG:
		req.meta.ivx = invokeRT
		req.meta.ivk = func() { d.invoke_dereg(req) }
	case gilix.TYPE_LOCK:
		req.meta.ivx = invokeDF
		req.meta.ivk = func() { d.invoke_lock(req) }
	case gilix.TYPE_UNLOCK:
		req.meta.ivx = invokeDF
		req.meta.ivk = func() { d.invoke_unlock(req) }
	case gilix.TYPE_CLOSE:
		req.meta.ivx = invokeCN | invokeDF
		req.meta.ivk = func() { d.invoke_close(req) }
	case gilix.TYPE_CANCEL:
		req.meta.ivx = invokeCI
		req.meta.ivk = nil
	}
}

func (d *device) invoke_open(req *sotReq) {
}

func (d *device) invoke_close(req *sotReq) {
}

func (d *device) invoke_reg(req *sotReq) {
}

func (d *device) invoke_dereg(req *sotReq) {
}

func (d *device) invoke_unlock(req *sotReq) {
	req.meta.ret = d.unlock(req.seq)
}

func (d *device) invoke_lock(req *sotReq) {
	if req.meta.ret = util.ContextErr(req.ctx); req.meta.ret != gilix.RET_SUCCESS {
		return
	}
	if d.curlck == req.seq {
		req.meta.ret = gilix.RET_ALREADY_LOCKED
		return
	}
	if d.lock(req) {
		return
	}

	tried := false
	util.TickWorker(req.ctx, time.Millisecond*50,
		func() bool {
			if !tried {
				d.dev.OnLockTry()
				tried = true
			}
			return d.lock(req)
		},
		func(ctxErr gilix.RET) {
			req.meta.ret = ctxErr
		})
}

func (d *device) invoke_inf(req *sotReq, cee gilix.Callee, pci int) {
	if req.meta.ret = util.ContextErr(req.ctx); req.meta.ret != gilix.RET_SUCCESS {
		return
	}
	if pci >= 0 && pci < len(d.pollc) {
		req.meta.rsp = d.pollc[pci]
		return
	}
	if cee == nil {
		req.meta.ret = gilix.RET_UNSUPP_CATEGORY
		return
	}

	req.meta.rsp = cee(req.ctx, req.req.Para())
}

func (d *device) invoke_cmd(req *sotReq, cee gilix.Callee, chk bool, chg bool) {
	if req.meta.ret = util.ContextErr(req.ctx); req.meta.ret != gilix.RET_SUCCESS {
		return
	}
	if d.curlck != nil && d.curlck != req.seq {
		req.meta.ret = gilix.RET_LOCKED
		return
	}
	if cee == nil {
		req.meta.ret = gilix.RET_UNSUPP_COMMAND
		return
	}

	req.meta.chg = chg
	req.meta.rsp = cee(req.ctx, req.req.Para())
	if chk {
		req.meta.pc = d.poll()
	}
}

func (d *device) invoke_poll(req *sotReq) {
	req.meta.pc = d.poll()
	req.meta.chg = true
}

func (d *device) poll() gilix.PollCache {
	pc := make(gilix.PollCache, len(d.polls))
	for i, cee := range d.polls {
		if cee != nil {
			pc[i] = cee(context.Background(), nil)
		}
	}
	return pc
}

func (d *device) unlock(seq *session) gilix.RET {
	if d.curlck != seq {
		return gilix.RET_NOT_LOCKED_YET
	}

	d.curlck = nil
	d.curlcki = gilix.ID_NIL

	return gilix.RET_SUCCESS
}

func (d *device) lock(req *sotReq) bool {
	if d.curlck != nil {
		return false
	}

	d.curlck = req.seq
	d.curlcki = req.req.Id()

	return true
}

/*
1，CLOSE之前，要tryUnlock
6，tryLock的时候，考虑如何同时控制好curlck和curlcki
3，在invoke_req/dereg的时候调用Usr接口
//2，SOT_REQ_TYPE_SESSION_CLEAN 的时候要tryUnlock
//3，seq要有HS，与DEV和USR的初始化时机一致
//4，增加curlcki
//5，tryUnlock里同时控制好curlck和curlcki
*/
