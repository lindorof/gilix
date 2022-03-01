package ws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lindorof/gilix/acp"
	"github.com/lindorof/gilix/util"

	"github.com/gorilla/websocket"
)

func TestC1WR1(t *testing.T) {
	cnn := wscnn()
	defer cnn.Close()

	w := []byte("TestCnnRWClose")

	if err := wswrite(cnn, w); err != nil {
		t.Fatalf("wswrite err %v", err)
	}
	if err := wsread(cnn, w); err != nil {
		t.Fatalf("wsread err %v", err)
	}
}

func BenchmarkC1WRx(b *testing.B) {
	cnn := wscnn()
	defer cnn.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := []byte(fmt.Sprintf("BenchmarkC1WRx%d", i))
		if err := wswrite(cnn, w); err != nil {
			b.Errorf("wswrite err %v", err)
		}
		if err := wsread(cnn, w); err != nil {
			b.Errorf("wsread err %v", err)
		}
	}

	b.StopTimer()
}

func BenchmarkC1WxRxA(b *testing.B) {
	cnn := wscnn()
	defer cnn.Close()

	w := []byte(fmt.Sprintf("BenchmarkC1WxRxA%d", 8))

	b.ResetTimer()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < b.N; i++ {
			if err := wswrite(cnn, w); err != nil {
				b.Errorf("wswrite err %v", err)
			}
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			if err := wsread(cnn, w); err != nil {
				b.Errorf("wsread err %v", err)
			}
		}
		wg.Done()
	}()
	wg.Wait()

	b.StopTimer()
}

func BenchmarkCxWR1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cnn := wscnn()
		w := []byte(fmt.Sprintf("BenchmarkCxWR1%d", i))

		if err := wswrite(cnn, w); err != nil {
			b.Errorf("wswrite err %v", err)
		}
		if err := wsread(cnn, w); err != nil {
			b.Errorf("wsread err %v", err)
		}

		cnn.Close()
	}
}

func BenchmarkCxWR1P(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for i := 0; p.Next(); i++ {
			cnn := wscnn()
			if cnn == nil {
				b.Errorf("wscnn nil")
				continue
			}
			w := []byte(fmt.Sprintf("BenchmarkCxWR1P%d", i))

			if err := wswrite(cnn, w); err != nil {
				b.Errorf("wswrite err %v", err)
			}
			if err := wsread(cnn, w); err != nil {
				b.Errorf("wsread err %v", err)
			}

			cnn.Close()
		}
	})
}

func TestMain(m *testing.M) {
	syncer := setup()
	code := m.Run()
	teardown(syncer)

	os.Exit(code)
}

func setup() *util.Syncer {
	var srv acp.Acceptor = CreateServer(nil, ":8808", "/")
	srv.SetSeqHook(acp.SeqHookMock)

	syncer := util.CreateSyncer(context.Background())
	syncer.Async(srv.LoopSync, srv.LoopBreak)

	fmt.Printf("@@@ setup ok, waiting for 3 seconds...\n")
	time.Sleep(3 * time.Second)

	return syncer
}

func teardown(syncer *util.Syncer) {
	syncer.WaitRelease(util.SyncerWaitModeCancel)
}

func wscnn() *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial("ws://:8808", nil)
	if err != nil {
		fmt.Printf("websocket.DefaultDialer.Dial %v\n", err)
		return nil
	}
	return ws
}

func wsread(cnn *websocket.Conn, exp []byte) error {
	_, msg, err := cnn.ReadMessage()
	if err != nil {
		return err
	}
	if !bytes.Equal(exp, msg) {
		return errors.New("msg not equal")
	}
	return nil
}

func wswrite(cnn *websocket.Conn, msg []byte) error {
	if err := cnn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	return nil
}
