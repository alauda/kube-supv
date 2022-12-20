package ping

import (
	"sync"
	"time"

	"github.com/alauda/kube-supv/pkg/errors"
)

func PingTCPs(timeoutSeconds int, addresses ...string) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	errs := errors.NewErrors()
	errCh := make(chan error)
	finish := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(len(addresses))

	for _, addr := range addresses {
		go pingWorker(addr, deadline, errCh, &wg)
	}

	go pingCollector(errs, errCh, finish)

	wg.Wait()
	<-finish

	return errs.AsError()
}

func pingWorker(addr string, deadline time.Time, errCh chan<- error, wg *sync.WaitGroup) {
	err := PingTCP(addr, deadline)
	if err != nil {
		errCh <- err
	}
	wg.Done()
}

func pingCollector(errs *errors.Errors, errCh <-chan error, finish chan<- bool) {
	for {
		err := <-errCh
		if err != nil {
			close(finish)
			break
		}
		errs.Append(err)
	}
}
