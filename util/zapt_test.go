package util

import (
	"os"
	"testing"
)

func TestFilesOfDay(t *testing.T) {
	zapt := zapt("FilesOfDay")
	logOne(zapt)
}

func TestDirsOfDay(t *testing.T) {
	zapt := zapt("DirsOfDay")
	logOne(zapt)
}

func BenchmarkFilesOfDay(b *testing.B) {
	zapt := zapt("FilesOfDay")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logOne(zapt)
	}
}

func BenchmarkDirsOfDay(b *testing.B) {
	zapt := zapt("DirsOfDay")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logOne(zapt)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func zapt(xday string) *Zapt {
	return newZapt("./gilix_trace", "util/zapt", "zapt", xday, 2, "DEBUG")
}

func logOne(zapt *Zapt) {
	zapt.Debugf("this is debug %f", 3.1415926)
	zapt.Infof("this is info %d %s", 88, "hehe")
	zapt.Warnf("this is warn 0x%08X", 0xABCDEF)
	zapt.Errorf("this is error %08b", '1')
}
