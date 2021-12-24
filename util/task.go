package util

type Worker func(data interface{}, ctx interface{})

type Task struct {
	w Worker
	t chan *job
	q chan bool
	d chan bool
}

type job struct {
	data interface{}
	ctx  interface{}
}

func CreateTask(w Worker) *Task {
	return &Task{
		w: w,
		t: make(chan *job, 1025),
		q: make(chan bool, 1),
		d: make(chan bool, 1),
	}
}

func CreateTaskStart(w Worker) *Task {
	task := CreateTask(w)
	task.Start()
	return task
}

func (task *Task) Start() {
	go func() {
	LOOP:
		for {
			select {
			case <-task.q:
				break LOOP
			case j := <-task.t:
				task.w(j.data, j.ctx)
			}
		}
		task.d <- true
	}()
}

func (task *Task) Stop() {
	task.q <- true
	<-task.d
}

func (task *Task) Put(data interface{}, ctx interface{}) {
	j := &job{data, ctx}
	task.t <- j
}
