package util

import (
	"context"
	"fmt"
	"testing"
)

func TestProct(t *testing.T) {
	cases := []struct {
		path string
		args []string
	}{
		{"nc", []string{"-u", "-l", "8881"}},
		{"nc", []string{"-u", "-l", "8882"}},
		{"nc", []string{"-u", "-l", "8883"}},
		{"nc", []string{"-u", "-l", "8884"}},
	}

	syncer := CreateSyncerWithSig(context.Background())
	proct := NewProct(syncer.Ctx())

	for _, c := range cases {
		n := fmt.Sprintf("%s-%v", c.path, c.args)
		proct.AddCmd(n, func(cmd *ProctCmd) {
			t.Logf("%s-[%d][%v]", cmd.N, cmd.I, cmd.E)
		}, c.path, c.args...)
	}

	syncer.Sync(func() { proct.Wait() }, func() {})
}
