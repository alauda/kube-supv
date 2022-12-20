package mock

import (
	"net"

	"github.com/alauda/kube-supv/pkg/log"
)

func Echo(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Errorf(`read from "%s", error: %v`, conn.RemoteAddr(), err)
		return
	}
	conn.Write(buf[:n])
}
