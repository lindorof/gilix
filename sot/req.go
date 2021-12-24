package sot

import (
	"context"
	"time"

	"github.com/lindorof/gilix"
)

const (
	srtSeqCreate    = 1
	srtSeqClean     = 2
	srtReqStart     = 11
	srtReqComplete  = 12
	srtPollStart    = 21
	srtPollComplete = 22
	srtEvtPost      = 101
)

type sotReq struct {
	tp  int
	seq *session
	req gilix.Msg
	evt gilix.Evt

	ctx  context.Context
	ctxc context.CancelFunc

	meta *invokeMeta
}

func newReq(tp int, seq *session, req gilix.Msg, evt gilix.Evt) *sotReq {
	sr := &sotReq{
		tp:  tp,
		seq: seq,
		req: req,
		evt: evt,
	}

	if req != nil {
		if req.Timeout() > 0 {
			dl := time.Now().Add(time.Millisecond * time.Duration(req.Timeout()))
			sr.ctx, sr.ctxc = context.WithDeadline(context.Background(), dl)
		} else {
			sr.ctx, sr.ctxc = context.WithCancel(context.Background())
		}
	}

	return sr
}

func (req *sotReq) tpr() {
	switch req.tp {
	case srtReqStart:
		req.tp = srtReqComplete
	case srtPollStart:
		req.tp = srtPollComplete
	}
}
