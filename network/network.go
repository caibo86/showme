// -------------------------------------------
// @file      : network.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/15 下午3:57
// -------------------------------------------

package network

import (
	"github.com/caibo86/logger"
	"io"
	"net"
)

const (
	KeepAlive     = "KEEP_ALIVE\n"
	NewConnection = "NEW_CONNECTION\n"
)

func TCPListener(addr string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", tcpAddr)
}

func CreateTCPConn(addr string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, tcpAddr)
}

func Join2Conn(local *net.TCPConn, remote *net.TCPConn) {
	go join2Conn(local, remote)
	go join2Conn(remote, local)
}

func join2Conn(local *net.TCPConn, remote *net.TCPConn) {
	defer func() {
		err := remote.Close()
		if err == nil {
			logger.Infof("close remote %s", remote.RemoteAddr())
		}
		err = local.Close()
		if err == nil {
			logger.Infof("close local %s", local.RemoteAddr())
		}
	}()
	_, err := io.Copy(local, remote)
	if err != nil {
		return
	}
}
