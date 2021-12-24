package gilix

import "gitee.com/lindorof/gilix/acp"

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
var CPS Xcps = nil
