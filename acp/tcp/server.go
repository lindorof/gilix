package tcp

import (
	"fmt"
	"strings"

	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"

	"context"
	"net"
)

type tcpServer struct {
	sot       interface{}
	addr      string
	cnnSyncer *util.Syncer

	listener net.Listener
	listenq  bool

	zapt    *util.Zapt
	seqHook acp.SeqHook
}

func CreateServer(sot interface{}, addr string) *tcpServer {
	srv := &tcpServer{
		sot:       sot,
		addr:      addr,
		cnnSyncer: util.CreateSyncer(context.Background()),
	}

	mod := fmt.Sprintf("acptcp-%s", strings.ReplaceAll(addr, ":", "-"))
	srv.zapt = util.ZaptByCfg(0, mod, "tcpServer")
	srv.zapt.Infof("[%p] CreateServer [%s]", srv, addr)

	return srv
}

func (srv *tcpServer) SetSeqHook(sh acp.SeqHook) {
	srv.seqHook = sh
}

func (srv *tcpServer) LoopSync() {
	var err error = nil

	srv.listener, err = net.Listen("tcp", srv.addr)
	srv.zapt.Infof("[%p] net.Listen(%s) return [%v]", srv, srv.addr, err)
	if err != nil {
		return
	}

	for {
		cnn, err := srv.listener.Accept()
		if err != nil {
			srv.zapt.Errorf("[%p] srv.listener.Accept return [%t][%v]", srv, srv.listenq, err)
			if srv.listenq {
				break
			} else {
				continue
			}
		}
		go srv.tcpHandler(cnn)
	}

	srv.cnnSyncer.WaitRelease(util.SyncerWaitModeCancel)
	srv.zapt.Infof("[%p] srv.cnnSyncer.SyncerWaitModeCancel end", srv)
}

func (srv *tcpServer) LoopBreak() {
	srv.zapt.Infof("[%p] srv.listener.Close begin", srv)
	srv.listenq = true
	err := srv.listener.Close()
	srv.zapt.Infof("[%p] srv.listener.Close end [%v]", srv, err)
}

func (srv *tcpServer) tcpHandler(cnn net.Conn) {
	sys := util.CreateSysrw(cnn, 1024*10)
	defer cnn.Close()

	chw := make(chan []byte, 1025)
	seq := srv.seqHook(chw)
	defer seq.Clean()

	freader := func() {
		for {
			_, msg, err := sys.Read()
			if err != nil {
				srv.zapt.Errorf("[%p] [%p] util.SysRead err : %s", srv, cnn, err.Error())
				return
			}
			srv.zapt.Debugf("[%p] [%p] util.SysRead : %s", srv, cnn, string(msg))
			seq.Putr(msg)
		}
	}

	fwriter := func() {
		for msg := range chw {
			if msg == nil {
				srv.zapt.Infof("[%p] [%p] cnn.chw.PeekMessage : nil to exit", srv, cnn)
				return
			}
			srv.zapt.Debugf("[%p] [%p] cnn.chw.PeekMessage : %s", srv, cnn, string(msg))
			if err := sys.Write("", msg); err != nil {
				srv.zapt.Errorf("[%p] [%p] util.SysWrite err : %s", srv, cnn, err.Error())
				return
			}
		}
	}

	rwSyncer := srv.cnnSyncer.DeriveSyncer()

	rwSyncer.Async(freader, func() { cnn.Close() })
	rwSyncer.Async(fwriter, func() { chw <- nil })

	rwSyncer.WaitRelease(util.SyncerWaitModeAny)
}
