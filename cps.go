package gilix

import "github.com/lindorof/gilix/acp"

type DevCp interface {
	PostEvt(Evt)
	PollSwitch(bool)
}

type Xcps interface {
	SotLoopSync()
	SotLoopBreak()
	SubmitAcp(acp acp.Acceptor)
}

// CPS Getter
var NewCPS func() Xcps = nil
