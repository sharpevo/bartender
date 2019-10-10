package workerpool

type Request struct {
	Data   interface{}
	Errorc chan error
}

type launchWorker func(int, chan Request)

type dispatcher struct {
	inputc    chan Request
	workerNum int
}

func NewDispatcher(queueNum int, workerNum int, lw launchWorker) *dispatcher {
	d := &dispatcher{
		inputc: make(chan Request, queueNum),
	}
	for i := 0; i < workerNum; i++ {
		d.addWorker(i, lw)
	}
	return d
}

func (d *dispatcher) addWorker(id int, lw launchWorker) {
	go lw(id, d.inputc)
}

func (d *dispatcher) AddRequest(r Request) {
	d.inputc <- r
}

func (d *dispatcher) Stop() {
	close(d.inputc)
}
