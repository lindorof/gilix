package gilix

import (
	"context"
)

/* ********************************************************** */
// const
/* ********************************************************** */

type (
	HS      int
	ID      int
	TYPE    int
	CODE    int
	TIMEOUT int
	RET     int
	QUEUET  int
	ERCV    int
	EHSU    int
	PARA    interface{}
)

const (
	HS_NIL = 0
	ID_NIL = 0

	TYPE_OPEN    = 1
	TYPE_CLOSE   = 2
	TYPE_REG     = 3
	TYPE_DEREG   = 4
	TYPE_INF     = 5
	TYPE_CMD     = 6
	TYPE_LOCK    = 7
	TYPE_UNLOCK  = 8
	TYPE_CANCEL  = 50
	TYPE_EVT_USR = 101
	TYPE_EVT_SRV = 102
	TYPE_EVT_EXE = 103
	TYPE_EVT_SYS = 104

	RET_SUCCESS         = 0
	RET_TIMEOUT         = 1
	RET_CANCELLED       = 2
	RET_LOCKED          = 3
	RET_ALREADY_LOCKED  = 4
	RET_NOT_LOCKED_YET  = 5
	RET_UNSUPP_CATEGORY = 6
	RET_UNSUPP_COMMAND  = 7

	QUEUET_AUTO = 0
	QUEUET_RT   = 1
	QUEUET_DF   = 2

	ERCV_CURRENT = 1
	ERCV_LOCKER  = 2
	ERCV_ALL     = 3

	EHSU_CURRENT = 1
	EHSU_LOCKER  = 2
	EHSU_ALL     = 3
)

/* ********************************************************** */
// Msg
/* ********************************************************** */

type HsId interface {
	Hs() HS
	Id() ID
}

type TypeCode interface {
	Type() TYPE
	Code() CODE
}

type ParaX interface {
	Aux() PARA
	Para() PARA
}

type Rsp interface {
	Para() PARA
	Ret() RET
}

type Evt interface {
	TypeCode
	ParaX
}

type Msg interface {
	HsId
	TypeCode
	ParaX
	Timeout() TIMEOUT
	Ret() RET
	Phyname() string
}

/* ********************************************************** */
// Dev
/* ********************************************************** */

type Usr interface{}
type Callee func(context.Context, PARA) Rsp /* for Inf/Cmd only */
type PollCache []Rsp

type Dev interface {
	PollInterval() int
	PollFuncs() []Callee

	OnReq(Msg, Usr) (qut QUEUET, cee Callee, pci int, chk bool, chg bool) /* ret used for Inf/Cmd only */
	OnEvt(cur PollCache, e Evt) (ERCV, EHSU)
	OnLockTry()
	PollChange(old PollCache, new PollCache)
}

/* ********************************************************** */
// CBS
/* ********************************************************** */

type Xcbs interface {
	UsrSnap(Msg) Usr

	MsgEncode(Msg) []byte
	MsgDecode([]byte) Msg
	MsgByRsp(Msg, RET, Rsp) Msg
	MsgByEvt(HsId, Usr, Evt) []Msg

	DevInit(phy string, dc DevCp) Dev
	DevFini(Dev)

	ZaptCfg() (path string, mode string, purge int, lelvel string)
}

// CBS Setter
var CBS Xcbs = nil
