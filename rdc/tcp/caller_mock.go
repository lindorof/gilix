package tcp

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/lindorof/gilix/util"
)

type tcpServerMock struct {
	phy       string
	addr      string
	trace     bool
	cnnSyncer *util.Syncer

	task     *util.Task
	listener net.Listener
	listenq  bool
}

func (srv *tcpServerMock) mockLogf(format string, v ...interface{}) {
	if srv.trace {
		s1 := fmt.Sprintf(">> [%s][%p] ", srv.phy, srv)
		s2 := fmt.Sprintf(format, v...)
		log.Print(s1, s2)
	}
}

func createServerMock(addr string, phy string, trace bool) *tcpServerMock {
	srv := &tcpServerMock{
		phy:       phy,
		addr:      addr,
		trace:     trace,
		cnnSyncer: util.CreateSyncer(context.Background()),
	}

	return srv
}

func (srv *tcpServerMock) work(data interface{}, ctx interface{}) {
	chw := ctx.(chan []byte)
	chw <- data.([]byte)
}

func (srv *tcpServerMock) loopSync() {
	srv.task = util.CreateTaskStart(srv.work)
	defer srv.task.Stop()

	ls, err := net.Listen("tcp", srv.addr)
	srv.mockLogf("net.Listen(%s) return [%v]\n", srv.addr, err)
	if err != nil {
		return
	}
	srv.listener = ls

	for {
		cnn, err := srv.listener.Accept()
		if err != nil {
			srv.mockLogf("srv.listener.Accept return [%t][%v]\n", srv.listenq, err)
			if srv.listenq {
				break
			} else {
				continue
			}
		}
		go srv.tcpHandler(cnn)
	}

	srv.cnnSyncer.WaitRelease(util.SYNCER_WAIT_MODE_CANCEL)
	srv.mockLogf("srv.cnnSyncer.SYNCER_WAIT_MODE_CANCEL return\n")
}

func (srv *tcpServerMock) loopBreak() {
	srv.mockLogf("srv.listener.Close begin\n")
	srv.listenq = true
	err := srv.listener.Close()
	srv.mockLogf("srv.listener.Close end [%v]\n", err)
}

func (srv *tcpServerMock) tcpHandler(cnn net.Conn) {
	srv.mockLogf("[%p] tcpHandler Entry\n", cnn)

	sys := util.CreateSysrw(cnn, 1024*10)
	defer cnn.Close()

	chw := make(chan []byte, 1025)

	freader := func() {
		for {
			_, msg, err := sys.Read()
			if err != nil {
				srv.mockLogf("<Exit> [%p] util.SysRead : %s\n", cnn, err.Error())
				return
			}
			//srv.mockLogf("read [%p]\n%s\n", cnn, string(msg))
			srv.task.Put(msg, chw)
		}
	}

	fwriter := func() {
		for msg := range chw {
			if msg == nil {
				srv.mockLogf("<Exit> [%p] cnn.chw.PeekMessage : nil\n", cnn)
				return
			}
			if err := sys.Write("0", msg); err != nil {
				srv.mockLogf("<Exit> [%p] util.SysWrite : %s\n", cnn, err.Error())
				return
			}
			//srv.mockLogf("write [%p]\n%s\n", cnn, string(msg))
		}
	}

	rwSyncer := srv.cnnSyncer.DeriveSyncer()

	rwSyncer.Async(freader, func() { cnn.Close() })
	rwSyncer.Async(fwriter, func() { chw <- nil })

	rwSyncer.WaitRelease(util.SYNCER_WAIT_MODE_ANY)

	srv.mockLogf("[%p] tcpHandler Exit\n", cnn)
}
