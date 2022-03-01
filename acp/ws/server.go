package ws

import (
	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"

	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type wsServer struct {
	sot       interface{}
	cnnSyncer *util.Syncer

	upgrader *websocket.Upgrader
	handler  *http.ServeMux
	server   *http.Server

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

	log.Printf("[%p] CreateServer\n", srv)
	return srv
}

func (srv *wsServer) SetSeqHook(sh acp.SeqHook) {
	srv.seqHook = sh
}

func (srv *wsServer) LoopSync() {
	log.Printf("[%p] server.ListenAndServe begin\n", srv)
	err := srv.server.ListenAndServe()
	log.Printf("[%p] server.ListenAndServe end [%v]\n", srv, err)

	srv.cnnSyncer.WaitRelease(util.SyncerWaitModeCancel)
	log.Printf("[%p] srv.cnnSyncer.SyncerWaitModeCancel return\n", srv)
}

func (srv *wsServer) LoopBreak() {
	log.Printf("[%p] server.Shutdown begin\n", srv)
	err := srv.server.Shutdown(context.Background())
	log.Printf("[%p] server.Shutdown end [%v]\n", srv, err)
}

func (srv *wsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	cnn, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%p] [%p] wsUpgrade : %s\n", srv, cnn, err.Error())
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
				//log.Printf("[%p] [%p] cnn.ReadMessage : %s\n", srv, cnn, err.Error())
				return
			}
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
				//log.Printf("[%p] [%p] cnn.chw.PeekMessage : nil\n", srv, cnn)
				return
			}
			if err := cnn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[%p] [%p] cnn.WriteMessage : %s\n", srv, cnn, err.Error())
				return
			}
		}
	}

	rwSyncer := srv.cnnSyncer.DeriveSyncer()

	rwSyncer.Async(freader, func() { cnn.Close() })
	rwSyncer.Async(fwriter, func() { chw <- nil })

	rwSyncer.WaitRelease(util.SyncerWaitModeAny)
}
