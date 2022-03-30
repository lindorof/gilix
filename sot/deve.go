package sot

import "github.com/lindorof/gilix"

type eRcvHsId struct {
	rcv *session
	hs  gilix.HS
	id  gilix.ID
}

func (t *eRcvHsId) Hs() gilix.HS {
	return t.hs
}

func (t *eRcvHsId) Id() gilix.ID {
	return t.id
}

func (d *device) erhi(ercv gilix.ERCV, ehsu gilix.EHSU, seq *session) *eRcvHsId {
	t := &eRcvHsId{}

	if ercv == gilix.ERCV_CURRENT && d.curreq != nil && d.curreq.seq == seq {
		t.rcv = seq
	}
	if ercv == gilix.ERCV_LOCKER && d.curlck == seq {
		t.rcv = seq
	}
	if ercv == gilix.ERCV_ALL {
		t.rcv = seq
	}

	if ehsu == gilix.ERCV_CURRENT && d.curreq != nil && d.curreq.req != nil {
		t.hs = d.curreq.req.Hs()
		t.id = d.curreq.req.Id()
	}
	if ehsu == gilix.ERCV_LOCKER && d.curlck != nil {
		t.hs = d.curlck.hs
		t.id = d.curlcki
	}
	if ehsu == gilix.ERCV_ALL {
		t.hs = seq.hs
		t.id = gilix.ID_NIL
	}

	return t
}
