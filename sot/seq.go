package sot

import (
	"log"

	"github.com/lindorof/gilix"
	"github.com/lindorof/gilix/acp"
)

type session struct {
	se  *sotEngine
	acp acp.Acceptor
	chw chan<- []byte

	dev *device
	usr gilix.Usr
	hs  gilix.HS
}

func newSession(se *sotEngine, acp acp.Acceptor, chw chan<- []byte) *session {
	seq := &session{
		se:  se,
		acp: acp,
		chw: chw,
	}

	log.Printf("[%p] [%p] createSession\n", acp, seq)
	return seq
}

func (seq *session) Putr(r []byte) {
	msg := gilix.CBS.MsgDecode(r)
	if msg == nil {
		return
	}

	if seq.dev == nil {
		seq.dev = <-seq.se.devr.matchDev(msg.Phyname())
		seq.usr = gilix.CBS.UsrSnap(msg)
		seq.hs = msg.Hs()

		seq.dev.putq(newReq(srtSeqCreate, seq, nil, nil))
	}

	seq.dev.putq(newReq(srtReqStart, seq, msg, nil))
}

func (seq *session) Clean() {
	seq.dev.putq(newReq(srtSeqClean, seq, nil, nil))
}

func (seq *session) putw(msg gilix.Msg) {
	if msg == nil {
		return
	}
	if w := gilix.CBS.MsgEncode(msg); w != nil {
		seq.chw <- w
	}
}
