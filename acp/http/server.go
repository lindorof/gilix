package http

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"gitee.com/lindorof/gilix/acp"
	"gitee.com/lindorof/gilix/util"
)

type httpServer struct {
	sot       interface{}
	cnnSyncer *util.Syncer

	handler *http.ServeMux
	server  *http.Server

	seqHook acp.SeqHook
}

func CreateServer(sot interface{}, addr string, idlet time.Duration, url ...string) *httpServer {
	if len(url) <= 0 {
		url = []string{"/"}
	}

	srv := &httpServer{
		sot:       sot,
		cnnSyncer: util.CreateSyncer(context.Background()),
	}

	srv.handler = &http.ServeMux{}
	for _, u := range url {
		srv.handler.HandleFunc(u, srv.httpHandler)
	}

	srv.server = &http.Server{
		Addr:        addr,
		Handler:     srv.handler,
		IdleTimeout: idlet,
	}

	log.Printf("[%p] CreateServer\n", srv)
	return srv
}

func (srv *httpServer) SetSeqHook(sh acp.SeqHook) {
	srv.seqHook = sh
}

func (srv *httpServer) LoopSync() {
	log.Printf("[%p] server.ListenAndServe begin\n", srv)
	err := srv.server.ListenAndServe()
	log.Printf("[%p] server.ListenAndServe end [%v]\n", srv, err)

	srv.cnnSyncer.WaitRelease(util.SYNCER_WAIT_MODE_CANCEL)
	log.Printf("[%p] srv.cnnSyncer.SYNCER_WAIT_MODE_CANCEL return\n", srv)
}

func (srv *httpServer) LoopBreak() {
	log.Printf("[%p] server.Shutdown begin\n", srv)
	err := srv.server.Shutdown(context.Background())
	log.Printf("[%p] server.Shutdown end [%v]\n", srv, err)
}

func (srv *httpServer) httpHandler(w http.ResponseWriter, r *http.Request) {
	chw := make(chan []byte, 1025)
	seq := srv.seqHook(chw)
	defer seq.Clean()

	msg := bytes.NewBuffer(make([]byte, 1024*10))
	_, err := io.Copy(msg, r.Body)
	if err != nil {
		log.Printf("[%p] io.Copy(req.body) : %s\n", srv, err.Error())
		return
	}

	seq.Putr(msg.Bytes())
	srv.cnnSyncer.Sync(func() {
		msg, ok := <-chw
		if !ok || msg == nil {
			log.Printf("[%p] chw.PeekMessage : nil\n", srv)
			return
		}
		if _, err := w.Write(msg); err != nil {
			log.Printf("[%p] resp.Write : %s\n", srv, err.Error())
		}
	}, func() { chw <- nil })
}
