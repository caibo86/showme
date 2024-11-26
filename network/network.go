// -------------------------------------------
// @file      : network.go
// @author    : 蔡波
// @contact   : caibo923@gmail.com
// @time      : 2024/11/15 下午3:57
// -------------------------------------------

package network

import (
	"io"
	"log"
	"net"
)

const (
	KeepAlive     = "KEEP_ALIVE\n"
	NewConnection = "NEW_CONNECTION\n"
)

func CreateTCPListener(addr string) (*net.TCPListener, error) {
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
	go joinConn(local, remote)
	go joinConn(remote, local)
}

func joinConn(local *net.TCPConn, remote *net.TCPConn) {
	defer func() {
		_ = remote.Close()
		_ = local.Close()
	}()
	_, err := io.Copy(local, remote)
	if err != nil {
		log.Println("copy failed:", err)
		return
	}
}
