package util

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
)

type pack struct {
	fun  string
	data []byte
	err  error
}

type Sysrw struct {
	cnn   net.Conn
	rcap  int
	buf   []byte
	packs chan *pack
}

func CreateSysrw(cnn net.Conn, rcap int) *Sysrw {
	return &Sysrw{
		cnn:   cnn,
		rcap:  rcap,
		buf:   nil,
		packs: make(chan *pack, 1025),
	}
}

func (sys *Sysrw) Write(fun string, data []byte) error {
	for buf := append([]byte(fmt.Sprintf("%08X%s^", len(data), fun)), data...); ; {
		n, err := sys.cnn.Write(buf)
		if err != nil {
			return err
		}
		if n < len(buf) {
			buf = buf[n:]
			continue
		}
		return nil
	}
}

func (sys *Sysrw) Read() (string, []byte, error) {
	for {
		select {
		case msg := <-sys.packs:
			return msg.fun, msg.data, msg.err
		default:
			sys.read()
		}
	}
}

func (sys *Sysrw) read() {
	quit := false

READ:
	if quit {
		return
	}

	r := make([]byte, sys.rcap)
	n, err := sys.cnn.Read(r)
	if err != nil {
		sys.packs <- &pack{"", nil, err}
		return
	}

	if len(sys.buf) <= 0 {
		sys.buf = r[:n]
	} else {
		sys.buf = append(sys.buf, r[:n]...)
	}

PARSE:
	sep := bytes.IndexByte(sys.buf, '^')
	if sep < 8 {
		goto READ
	}

	data := sys.buf[sep+1:]
	datalen, _ := strconv.ParseInt(string(sys.buf[:8]), 16, 0)
	if len(data) < int(datalen) {
		goto READ
	}

	sys.packs <- &pack{string(sys.buf[8:sep]), data[:datalen], nil}
	sys.buf = data[datalen:]

	quit = true
	goto PARSE
}
