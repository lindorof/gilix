package util

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lindorof/gilix"
)

type ZaptDayMode int

const (
	FilesOfDay = 0
	DirsOfDay  = 1
)

type Zapt struct {
	log *zap.SugaredLogger
}

func ZaptByCfg(mod string, file string) *Zapt {
	path, mode, purge, lelvel := gilix.CBS.ZaptCfg()
	return CreateZapt(path, mod, file, mode, purge, lelvel)
}

func CreateZapt(path string, mod string, file string, xday string, purge int, level string) *Zapt {
	core := zapcore.NewCore(
		encoder(),
		zapcore.AddSync(writer(path, mod, file, str2xday(xday), purge)),
		str2lvl(level))
	zapt := &Zapt{zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.DPanicLevel)).Sugar()}

	zapt.log.Infof("================================")
	zapt.log.Infof(filepath.Join(mod, file))
	zapt.log.Infof("================================")

	return zapt
}

func (zapt *Zapt) Debugf(template string, args ...interface{}) {
	zapt.log.Debugf(template, args...)
}

func (zapt *Zapt) Infof(template string, args ...interface{}) {
	zapt.log.Infof(template, args...)
}

func (zapt *Zapt) Warnf(template string, args ...interface{}) {
	zapt.log.Warnf(template, args...)
}

func (zapt *Zapt) Errorf(template string, args ...interface{}) {
	zapt.log.Errorf(template, args...)
}

func (zapt *Zapt) DPanicf(template string, args ...interface{}) {
	zapt.log.DPanicf(template, args...)
}

func (zapt *Zapt) Panicf(template string, args ...interface{}) {
	zapt.log.Panicf(template, args...)
}

func (zapt *Zapt) Fatalf(template string, args ...interface{}) {
	zapt.log.Fatalf(template, args...)
}

func writer(path string, mod string, file string, xday ZaptDayMode, purge int) io.Writer {
	lp := ""
	rp := ""
	fp := ""
	tm := time.Now()

	if xday == DirsOfDay {
		lp = filepath.Join(path, fmt.Sprintf("%04d%02d%02d", tm.Year(), tm.Month(), tm.Day()))
		rp = path
		fp = mod2file(mod, file)
	} else {
		lp = filepath.Join(path, mod)
		rp = lp
		fp = file
	}

	er := os.MkdirAll(lp, 0744)
	if er != nil {
		panic(er)
	}

	wt, er := os.OpenFile(
		filepath.Join(lp, fmt.Sprintf("%s-%04d%02d%02d.log", fp, tm.Year(), tm.Month(), tm.Day())),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if er != nil {
		panic(er)
	}

	rotate(rp, fp, xday, purge)
	return wt
}

func encoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stack",
		FunctionKey:      "",
		ConsoleSeparator: " ",
		LineEnding:       "\r\n",
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("%-6s", l.CapitalString()))
		},
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(
				fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d:%03d",
					t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/int(time.Millisecond)))
		},
		EncodeCaller: func(e zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("[%s:%d/%s]", ecfile(e), e.Line, ecfunc(e)))
		},
	})
}

func ecfile(ec zapcore.EntryCaller) string {
	fn := ec.File
	if i := strings.LastIndexByte(fn, '/'); i != -1 {
		fn = fn[i+1:]
	}
	if i := strings.LastIndexByte(fn, '.'); i != -1 {
		fn = fn[:i]
	}
	return fn
}

func ecfunc(ec zapcore.EntryCaller) string {
	fn := ec.Function
	if i := strings.LastIndexByte(fn, '.'); i != -1 {
		fn = fn[i+1:]
	}
	return fn
}

func str2xday(s string) ZaptDayMode {
	s = strings.ToUpper(s)
	switch s {
	case "FILESOFDAY":
		return FilesOfDay
	case "DIRSOFDAY":
		return DirsOfDay
	default:
		return FilesOfDay
	}
}

func str2lvl(s string) zapcore.Level {
	s = strings.ToUpper(s)
	switch s {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "DPANIC":
		return zapcore.DPanicLevel
	case "PANIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func rotate(path string, file string, xday ZaptDayMode, purge int) {
	arr, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	fis := []fs.FileInfo{}
	for _, a := range arr {
		if a.IsDir() && xday == DirsOfDay {
			if _, err := strconv.ParseInt(a.Name(), 10, 0); err == nil {
				fis = append(fis, a)
			}
		}
		if !a.IsDir() && xday != DirsOfDay {
			if strings.HasPrefix(a.Name(), file) {
				fis = append(fis, a)
			}
		}
	}

	sort.SliceStable(fis, func(i, j int) bool {
		return strings.Compare(fis[i].Name(), fis[j].Name()) > 0
	})

	if purge > 0 && len(fis) > purge {
		for _, a := range fis[purge:] {
			os.RemoveAll(filepath.Join(path, a.Name()))
		}
	}
}

func mod2file(mod string, file string) string {
	mod = strings.ReplaceAll(mod, "/", "-")
	mod = strings.ReplaceAll(mod, "\\", "-")
	return mod + "--" + file
}
