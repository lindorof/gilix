package ws

import (
	"fmt"
	"strings"

	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"

	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

type wsServer struct {
	sot       interface{}
	cnnSyncer *util.Syncer

	upgrader *websocket.Upgrader
	handler  *http.ServeMux
	server   *http.Server

	zapt    *util.Zapt
	seqHook acp.SeqHook
}

func CreateServer(sot interface{}, addr string, url ...string) *wsServer {
	if len(url) <= 0 {
		url = []string{"/"}
	}

	srv := &wsServer{
		sot:       sot,
		cnnSyncer: util.CreateSyncer(context.Background()),
	}

	srv.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024 * 10,
		WriteBufferSize: 1024 * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	srv.handler = &http.ServeMux{}
	for _, u := range url {
		srv.handler.HandleFunc(u, srv.wsHandler)
	}

	srv.server = &http.Server{
		Addr:    addr,
		Handler: srv.handler,
	}

	mod := fmt.Sprintf("acpws-%s", strings.ReplaceAll(addr, ":", "-"))
	srv.zapt = util.ZaptByCfg(0, mod, "wsServer")
	srv.zapt.Infof("[%p] CreateServer [%s]", srv, addr)

	return srv
}

func (srv *wsServer) SetSeqHook(sh acp.SeqHook) {
	srv.seqHook = sh
}

func (srv *wsServer) LoopSync() {
	srv.zapt.Infof("[%p] server.ListenAndServe begin", srv)
	err := srv.server.ListenAndServe()
	srv.zapt.Infof("[%p] server.ListenAndServe end [%v]", srv, err)

	srv.cnnSyncer.WaitRelease(util.SyncerWaitModeCancel)
	srv.zapt.Infof("[%p] srv.cnnSyncer.SyncerWaitModeCancel end", srv)
}

func (srv *wsServer) LoopBreak() {
	srv.zapt.Infof("[%p] server.Shutdown begin", srv)
	err := srv.server.Shutdown(context.Background())
	srv.zapt.Infof("[%p] server.Shutdown end [%v]", srv, err)
}

func (srv *wsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	cnn, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		srv.zapt.Infof("[%p] [%p] wsUpgrade : %s", srv, cnn, err.Error())
		return
	}
	defer cnn.Close()

	chw := make(chan []byte, 1025)
	seq := srv.seqHook(chw)
	defer seq.Clean()

	freader := func() {
		for {
			msgType, msg, err := cnn.ReadMessage()
			if err != nil {
				srv.zapt.Errorf("[%p] [%p] cnn.ReadMessage err : %s", srv, cnn, err.Error())
				return
			}
			srv.zapt.Debugf("[%p] [%p] cnn.ReadMessage : [%d]%s", srv, cnn, msgType, string(msg))
			switch {
			case msgType == websocket.CloseMessage:
				return
			case msgType == websocket.TextMessage || msgType == websocket.BinaryMessage:
				seq.Putr(msg)
			}
		}
	}

	fwriter := func() {
		for msg := range chw {
			if msg == nil {
				srv.zapt.Infof("[%p] [%p] cnn.chw.PeekMessage : nil to exit", srv, cnn)
				return
			}
			srv.zapt.Debugf("[%p] [%p] cnn.chw.PeekMessage : %s", srv, cnn, string(msg))
			if err := cnn.WriteMessage(websocket.TextMessage, msg); err != nil {
				srv.zapt.Errorf("[%p] [%p] cnn.WriteMessage err : %s", srv, cnn, err.Error())
				return
			}
		}
	}

	rwSyncer := srv.cnnSyncer.DeriveSyncer()

	rwSyncer.Async(freader, func() { cnn.Close() })
	rwSyncer.Async(fwriter, func() { chw <- nil })

	rwSyncer.WaitRelease(util.SyncerWaitModeAny)
}
