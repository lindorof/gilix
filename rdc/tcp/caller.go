package tcp

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"gitee.com/lindorof/gilix/util"
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
}

func CreateCaller(addr string, dialdt time.Duration) *tcpCaller {
	caller := &tcpCaller{
		fini: make(chan bool, 1),
		ivks: make(chan *invocation, 1025),

		addr:   addr,
		dialdt: dialdt,
	}

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
	caller.ivks <- nil
	<-caller.fini
}

func (caller *tcpCaller) loop() {
	cnn, cerr := net.DialTimeout("tcp", caller.addr, caller.dialdt)
	sys := util.CreateSysrw(cnn, 1024*10)

	for ivk := range caller.ivks {
		if ivk == nil {
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
		return
	}

	data, ivk.err = json.MarshalIndent(ivk.in, "", "  ")
	if ivk.err != nil {
		return
	}

	ivk.err = sys.Write(ivk.fun, data)
	if ivk.err != nil {
		return
	}

	rets, data, ivk.err = sys.Read()
	if ivk.err != nil {
		return
	}

	ivk.ret, ivk.err = strconv.ParseInt(rets, 10, 0)
	if ivk.err != nil {
		return
	}

	ivk.err = json.Unmarshal(data, ivk.out)
	if ivk.err != nil {
		return
	}
}
