package mock

import (
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
)

func ValidatePorts(ports ...int) error {
	for _, port := range ports {
		if port <= 0 || port > 65535 {
			return fmt.Errorf(`port %d is out of range between 1 and 65535`, port)
		}
	}
	return nil
}

func ListenTCP(durationSeconds int, ports ...int) error {
	if err := ValidatePorts(ports...); err != nil {
		return err
	}

	var servers []io.Closer
	defer func() {
		for _, s := range servers {
			s.Close()
		}
	}()

	t := time.NewTimer(time.Duration(durationSeconds) * time.Second)

	for _, port := range ports {
		s, err := NewTCPServer(fmt.Sprintf(":%d", port), Echo)
		if err != nil {
			return errors.Wrapf(err, `start tcp server at port: %d`, port)
		}
		servers = append(servers, s)
	}
	<-t.C
	return nil
}
