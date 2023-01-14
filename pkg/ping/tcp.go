package ping

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/pkg/errors"
)

func ConnectTCP(addr string, deadline time.Time) error {
	dialer := net.Dialer{
		Timeout:  time.Second * 2,
		Deadline: deadline,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, `tcp dial %s`, addr)
	}
	conn.Close()
	return nil
}

func PingTCP(addr string, deadline time.Time) error {
	dialer := net.Dialer{
		Timeout:  time.Second * 2,
		Deadline: deadline,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, `tcp dial %s`, addr)
	}
	defer conn.Close()

	playload := makePayload()
	wsize, err := conn.Write([]byte(playload))
	if err != nil {
		return errors.Wrapf(err, `tcp write %s %v`, addr, string(playload))
	}
	buf := make([]byte, 1024)

	rsize, err := conn.Read(buf)
	if err != nil {
		return errors.Wrapf(err, `tcp write %s %v`, addr, string(playload))
	}
	received := string(buf[:rsize])
	if wsize != rsize || received != playload {
		return fmt.Errorf("tcp received unexcepted data %s != %s", string(playload), string(buf[:rsize]))
	}
	return nil
}

func makePayload() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int())
}
