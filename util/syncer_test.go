package util

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func wm2s(wm SyncerWaitMode) string {
	switch wm {
	case SyncerWaitModeIdle:
		return "Idle"
	case SyncerWaitModeCancel:
		return "Cancel"
	case SyncerWaitModeAny:
		return "Any"
	}
	return "---"
}

func syncf(s string, m int, l func(string)) (func(), func()) {
	c := make(chan bool, 1)
	t := time.NewTicker(1 * time.Second)
	return func() {
			for i := 1; i <= m; i++ {
				select {
				case <-c:
					l(fmt.Sprintf("%s !!!cancel", s))
					return
				case <-t.C:
					l(fmt.Sprintf("%s sleep %ds", s, i))
				}
			}
			l(fmt.Sprintf("%s done", s))
		}, func() {
			c <- true
		}
}

func TestSyncerAsync(t *testing.T) {
	cases := []struct {
		wm SyncerWaitMode
		nt []int
	}{
		{SyncerWaitModeIdle, []int{5}},
		{SyncerWaitModeIdle, []int{3, 6}},
		{SyncerWaitModeCancel, []int{6}},
		{SyncerWaitModeCancel, []int{2, 4}},
		{SyncerWaitModeAny, []int{5, 9}},
		{SyncerWaitModeAny, []int{3, 5, 7, 8, 9}},
	}

	for _, c := range cases {
		n := fmt.Sprintf("%s_%v", wm2s(c.wm), c.nt)
		t.Run(n, func(t *testing.T) {
			syncer := CreateSyncerWithSig(context.Background())
			for _, i := range c.nt {
				f, b := syncf(fmt.Sprintf("%s_%d", n, i), i, func(s string) { t.Log(s) })
				syncer.Async(f, b)
			}
			syncer.WaitRelease(c.wm)
		})
	}
}

func TestDeriveSyncerAsync(t *testing.T) {
	cases := []struct {
		wm SyncerWaitMode
		pt int
		nt []int
	}{
		{SyncerWaitModeIdle, 3, []int{5}},
		{SyncerWaitModeIdle, 8, []int{3, 6}},
		{SyncerWaitModeCancel, 5, []int{3}},
		{SyncerWaitModeCancel, 6, []int{2, 4}},
		{SyncerWaitModeCancel, 2, []int{5, 8}},
		{SyncerWaitModeAny, 8, []int{3, 6}},
		{SyncerWaitModeAny, 2, []int{5, 9}},
	}

	for _, c := range cases {
		n := fmt.Sprintf("%s<%d>_%v", wm2s(c.wm), c.pt, c.nt)
		t.Run(n, func(t *testing.T) {
			syncer := CreateSyncerWithSig(context.Background())
			pf, pb := syncf(n, c.pt, func(s string) { t.Log(s) })
			syncer.Async(pf, pb)

			for _, i := range c.nt {
				f, b := syncf(fmt.Sprintf("%s_%d", n, i), i, func(s string) { t.Log(s) })
				ds := syncer.DeriveSyncer()
				ds.Async(f, b)
			}

			syncer.WaitRelease(c.wm)
		})
	}
}

func TestSyncerSync(t *testing.T) {
	cases := []struct {
		st int
	}{
		{3},
		{6},
	}

	for _, c := range cases {
		n := fmt.Sprintf("sync<%d>", c.st)
		t.Run(n, func(t *testing.T) {
			syncer := CreateSyncerWithSig(context.Background())
			f, b := syncf(n, c.st, func(s string) { t.Log(s) })
			syncer.Sync(f, b)
		})
	}
}
