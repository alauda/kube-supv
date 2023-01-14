package ping

import (
	"time"
)

func ConnectTCPs(timeoutSeconds int, addresses ...string) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	return NewWorker(ConnectTCP, deadline, addresses...).Run()
}

func PingTCPs(timeoutSeconds int, addresses ...string) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	return NewWorker(PingTCP, deadline, addresses...).Run()
}
