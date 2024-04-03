package utils

import (
	"net"
	"strings"
)

const UnixAddrPrefix = "unix://"

func HandlerConn(addr string) (net.Conn, error) {
	if strings.HasPrefix(addr, UnixAddrPrefix) {
		return net.Dial("unix", strings.TrimPrefix(addr, UnixAddrPrefix))
	} else {
		return net.Dial("tcp", addr)
	}
}

func HandlerListen(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, UnixAddrPrefix) {
		return net.Listen("unix", strings.TrimPrefix(addr, UnixAddrPrefix))
	} else {
		return net.Listen("tcp", addr)
	}
}
