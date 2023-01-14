package ping

import (
	"sync"
	"time"

	"github.com/alauda/kube-supv/pkg/errarr"
)

type WorkerFun func(addr string, deadline time.Time) error

type Worker struct {
	deadline time.Time
	adds     []string
	fun      WorkerFun
}

func NewWorker(fun WorkerFun, deadline time.Time, adds ...string) *Worker {
	return &Worker{
		deadline: deadline,
		fun:      fun,
		adds:     adds,
	}
}

func (w *Worker) Run() error {
	errs := errarr.NewErrors()
	errCh := make(chan error)
	finish := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(len(w.adds))

	for _, addr := range w.adds {
		go w.runOne(addr, errCh, &wg, w.fun)
	}

	go w.collector(errs, errCh, finish)

	wg.Wait()
	close(errCh)
	<-finish
	return errs.AsError()
}

func (w *Worker) runOne(addr string, errCh chan<- error, wg *sync.WaitGroup, fun WorkerFun) {
	err := fun(addr, w.deadline)
	if err != nil {
		errCh <- err
	}
	wg.Done()
}

func (w *Worker) collector(errs *errarr.Errors, errCh <-chan error, finish chan<- bool) {
	for {
		err := <-errCh
		if err == nil {
			close(finish)
			break
		}
		errs.Append(err)
	}
}
