package acp

type SeqMock struct {
	chw chan<- []byte
}

func SeqHookMock(chw chan<- []byte) Session {
	return &SeqMock{chw: chw}
}

func (seq *SeqMock) Putr(msg []byte) {
	seq.chw <- msg
}

func (seq *SeqMock) Clean() {

}
