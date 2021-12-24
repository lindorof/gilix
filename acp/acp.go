package acp

type SeqHook func(chw chan<- []byte) Session

type Session interface {
	Putr(msg []byte)
	Clean()
}

type Acceptor interface {
	SetSeqHook(sh SeqHook)
	LoopSync()
	LoopBreak()
}
