package tcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lindorof/gilix/util"
)

func TestInvoke(t *testing.T) {
	caller := caller()

	cases := []struct {
		name     string
		idx, exp int
	}{
		{"ok0", 0, 0},
		{"ok1", 1, 1},
		{"ok2", 2, 2},
		{"err1", 8, 9},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := invoke(caller, c.idx, c.exp); err != nil {
				t.Errorf("%v\n", err)
			}
		})
	}

	caller.Fini()
}

func BenchmarkC1Ix(b *testing.B) {
	caller := caller()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := invoke(caller, i, i); err != nil {
			b.Errorf("[%d] %v\n", i, err)
		}
	}

	b.StopTimer()
	caller.Fini()

	b.Logf("BenchmarkC1Ix Complete [%d]\n", b.N)
}

func BenchmarkC1IxP(b *testing.B) {
	caller := caller()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			if err := invoke(caller, 0, 0); err != nil {
				b.Errorf("[%d] %v\n", 0, err)
			}
		}
	})

	b.StopTimer()
	caller.Fini()

	b.Logf("BenchmarkC1IxP Complete [%d]\n", b.N)
}

func BenchmarkC1I1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		caller := caller()
		if err := invoke(caller, i, i); err != nil {
			b.Errorf("[%d] %v\n", i, err)
		}
		caller.Fini()
	}

	b.Logf("BenchmarkC1I1 Complete [%d]\n", b.N)
}

func TestMain(m *testing.M) {
	mockSyncer := setup()
	code := m.Run()
	teardown(mockSyncer)

	os.Exit(code)
}

func setup() *util.Syncer {
	mockSrv := createServerMock(":8808", "Phy1", false)
	mockSyncer := util.CreateSyncer(context.Background())
	mockSyncer.Async(mockSrv.loopSync, mockSrv.loopBreak)

	fmt.Printf("@@@ setup ok, waiting for 3 seconds...\n")
	time.Sleep(3 * time.Second)

	return mockSyncer
}

func teardown(mockSyncer *util.Syncer) {
	mockSyncer.WaitRelease(util.SyncerWaitModeCancel)
}

func caller() *tcpCaller {
	return CreateCaller(":8808", 5*time.Second)
}

func invoke(caller *tcpCaller, idx int, exp int) error {
	type para struct {
		Name string
		Data int
		exp  int
	}

	in := &para{"Track1", idx, exp}
	out := &para{}

	ret, err := caller.Invoke("ReadTrack", in, out)
	if err != nil {
		return err
	}
	if ret != 0 {
		return errors.New("ret not 0")
	}
	if out.Name != in.Name || out.Data != in.exp {
		return errors.New("out result not expected")
	}

	return nil
}
