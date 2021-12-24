package sot

type matchDevRet struct {
	phy string
	ret chan *device
}

type devRep struct {
	se *sotEngine

	devm    map[string]*device
	matcher chan *matchDevRet
}

func createDevRep(se *sotEngine) *devRep {
	return &devRep{
		se: se,

		devm:    make(map[string]*device, 256),
		matcher: make(chan *matchDevRet, 1025),
	}
}

func (dr *devRep) matchDev(phy string) <-chan *device {
	ret := make(chan *device, 1)
	dr.matcher <- &matchDevRet{phy, ret}
	return ret
}

func (dr *devRep) loopSync() {
	for m := range dr.matcher {
		if dev, ok := dr.devm[m.phy]; ok {
			m.ret <- dev
		} else {
			dev := createDev(dr.se, m.phy)
			dr.devm[m.phy] = dev
			m.ret <- dev
		}
	}
}

func (dr *devRep) loopBreak() {
	close(dr.matcher)
}
