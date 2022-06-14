package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"
)

type httpServer struct {
	sot       interface{}
	cnnSyncer *util.Syncer

	handler *http.ServeMux
	server  *http.Server

	zapt    *util.Zapt
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

	mod := fmt.Sprintf("acphttp-%s", strings.ReplaceAll(addr, ":", "-"))
	srv.zapt = util.ZaptByCfg(0, mod, "httpServer")
	srv.zapt.Infof("[%p] CreateServer [%s][idlet:%d]", srv, addr, idlet)

	return srv
}

func (srv *httpServer) SetSeqHook(sh acp.SeqHook) {
	srv.seqHook = sh
}

func (srv *httpServer) LoopSync() {
	srv.zapt.Infof("[%p] server.ListenAndServe begin", srv)
	err := srv.server.ListenAndServe()
	srv.zapt.Infof("[%p] server.ListenAndServe end [%v]", srv, err)

	srv.cnnSyncer.WaitRelease(util.SyncerWaitModeCancel)
	srv.zapt.Infof("[%p] srv.cnnSyncer.SyncerWaitModeCancel end", srv)
}

func (srv *httpServer) LoopBreak() {
	srv.zapt.Infof("[%p] server.Shutdown begin", srv)
	err := srv.server.Shutdown(context.Background())
	srv.zapt.Infof("[%p] server.Shutdown end [%v]", srv, err)
}

func (srv *httpServer) httpHandler(w http.ResponseWriter, r *http.Request) {
	chw := make(chan []byte, 1025)
	seq := srv.seqHook(chw)
	defer seq.Clean()

	msg := bytes.NewBuffer(make([]byte, 1024*10))
	_, err := io.Copy(msg, r.Body)
	if err != nil {
		srv.zapt.Errorf("[%p] io.Copy(req.body) err : %s", srv, err.Error())
		return
	}
	srv.zapt.Debugf("[%p] io.Copy(req.body) : %s", srv, string(msg.Bytes()))

	seq.Putr(msg.Bytes())
	srv.cnnSyncer.Sync(func() {
		msg, ok := <-chw
		if !ok || msg == nil {
			srv.zapt.Infof("[%p] chw.PeekMessage : nil to exit", srv)
			return
		}
		if _, err := w.Write(msg); err != nil {
			srv.zapt.Errorf("[%p] resp.Write err : %s", srv, err.Error())
		}
	}, func() { chw <- nil })
}
