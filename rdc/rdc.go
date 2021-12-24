package rdc

type Rdc interface {
	Invoke(fun string, in interface{}, out interface{}) (int, error)
	Fini()
}
