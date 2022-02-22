package util

import (
	"context"
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
		proct.AddCmd(func(cmd *ProctCmd) {
			t.Logf("%s-%v-[%d][%v]", c.path, c.args, cmd.C.Process.Pid, cmd.E)
		}, c.path, c.args...)
	}

	syncer.Sync(func() { proct.Wait() }, func() {})
}
