package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/lindorof/gilix/util"
)

type invocation struct {
	done chan bool

	fun string
	in  interface{}
	out interface{}
	ret int64
	err error
}

type tcpCaller struct {
	fini chan bool
	ivks chan *invocation

	addr   string
	dialdt time.Duration

	zapt *util.Zapt
}

func CreateCaller(addr string, dialdt time.Duration) *tcpCaller {
	caller := &tcpCaller{
		fini: make(chan bool, 1),
		ivks: make(chan *invocation, 1025),

		addr:   addr,
		dialdt: dialdt,
	}

	mod := fmt.Sprintf("gilix/rdctcp-%s", strings.ReplaceAll(addr, ":", "-"))
	caller.zapt = util.ZaptByCfg(mod, "tcpCaller")
	caller.zapt.Infof("[%p] CreateCaller [%s][dialdt:%d]", caller, addr, dialdt)

	go caller.loop()
	return caller
}

func (caller *tcpCaller) Invoke(fun string, in interface{}, out interface{}) (int, error) {
	ivk := &invocation{
		done: make(chan bool, 1),

		fun: fun,
		in:  in,
		out: out,
		ret: 0,
		err: nil,
	}

	caller.ivks <- ivk
	<-ivk.done

	return int(ivk.ret), ivk.err
}

func (caller *tcpCaller) Fini() {
	caller.zapt.Infof("[%p] caller.Fini begin", caller)
	caller.ivks <- nil
	<-caller.fini
	caller.zapt.Infof("[%p] caller.Fini end", caller)
}

func (caller *tcpCaller) loop() {
	cnn, cerr := net.DialTimeout("tcp", caller.addr, caller.dialdt)
	sys := util.CreateSysrw(cnn, 1024*10)

	for ivk := range caller.ivks {
		if ivk == nil {
			caller.zapt.Infof("[%p] range caller.ivks : nil to exit", caller)
			break
		}
		caller.invoke(ivk, sys, cerr)
	}

	if cerr == nil {
		cnn.Close()
	}
	caller.fini <- true
}

func (caller *tcpCaller) invoke(ivk *invocation, sys *util.Sysrw, cerr error) {
	defer func() { ivk.done <- true }()
	var rets string = ""
	var data []byte = nil

	ivk.err = cerr
	if ivk.err != nil {
		caller.zapt.Errorf("[%p] connect err : %v", caller, ivk.err)
		return
	}

	data, ivk.err = json.MarshalIndent(ivk.in, "", "  ")
	if ivk.err != nil {
		caller.zapt.Errorf("[%p] json.MarshalIndent ivk.in err : %v", caller, ivk.err)
		return
	}

	caller.zapt.Debugf("[%p] json.MarshalIndent ivk.in : %s", caller, string(data))

	ivk.err = sys.Write(ivk.fun, data)
	if ivk.err != nil {
		caller.zapt.Errorf("[%p] sys.Write err : %v", caller, ivk.err)
		return
	}

	rets, data, ivk.err = sys.Read()
	if ivk.err != nil {
		caller.zapt.Errorf("[%p] sys.Read err : %v", caller, ivk.err)
		return
	}

	caller.zapt.Debugf("[%p] sys.Read : [%s]%s", caller, rets, string(data))

	ivk.ret, ivk.err = strconv.ParseInt(rets, 10, 0)
	if ivk.err != nil {
		caller.zapt.Errorf("[%p] strconv.ParseInt rets err : %v", caller, ivk.err)
		return
	}

	ivk.err = json.Unmarshal(data, ivk.out)
	if ivk.err != nil {
		caller.zapt.Errorf("[%p] json.Unmarshal ivk.out err : %v", caller, ivk.err)
		return
	}
}
