package tcp

import (
	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"

	"context"
	"log"
	"net"
)

type tcpServer struct {
	sot       interface{}
	addr      string
	cnnSyncer *util.Syncer

	listener net.Listener
	listenq  bool

	seqHook acp.SeqHook
}

func CreateServer(sot interface{}, addr string) *tcpServer {
	srv := &tcpServer{
		sot:       sot,
		addr:      addr,
		cnnSyncer: util.CreateSyncer(context.Background()),
	}

	log.Printf("[%p] CreateServer\n", srv)
	return srv
}

func (srv *tcpServer) SetSeqHook(sh acp.SeqHook) {
	srv.seqHook = sh
}

func (srv *tcpServer) LoopSync() {
	var err error = nil

	srv.listener, err = net.Listen("tcp", srv.addr)
	log.Printf("[%p] net.Listen(%s) return [%v]\n", srv, srv.addr, err)
	if err != nil {
		return
	}

	for {
		cnn, err := srv.listener.Accept()
		if err != nil {
			log.Printf("[%p] srv.listener.Accept return [%t][%v]\n", srv, srv.listenq, err)
			if srv.listenq {
				break
			} else {
				continue
			}
		}
		go srv.tcpHandler(cnn)
	}

	srv.cnnSyncer.WaitRelease(util.SyncerWaitModeCancel)
	log.Printf("[%p] srv.cnnSyncer.SyncerWaitModeCancel return\n", srv)
}

func (srv *tcpServer) LoopBreak() {
	log.Printf("[%p] srv.listener.Close begin\n", srv)
	srv.listenq = true
	err := srv.listener.Close()
	log.Printf("[%p] srv.listener.Close end [%v]\n", srv, err)
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
				log.Printf("[%p] [%p] util.SysRead : %s\n", srv, cnn, err.Error())
				return
			}
			seq.Putr(msg)
		}
	}

	fwriter := func() {
		for msg := range chw {
			if msg == nil {
				log.Printf("[%p] [%p] cnn.chw.PeekMessage : nil\n", srv, cnn)
				return
			}
			if err := sys.Write("", msg); err != nil {
				log.Printf("[%p] [%p] util.SysWrite : %s\n", srv, cnn, err.Error())
				return
			}
		}
	}

	rwSyncer := srv.cnnSyncer.DeriveSyncer()

	rwSyncer.Async(freader, func() { cnn.Close() })
	rwSyncer.Async(fwriter, func() { chw <- nil })

	rwSyncer.WaitRelease(util.SyncerWaitModeAny)
}
