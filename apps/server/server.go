// -------------------------------------------
// @file      : server.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/21 下午6:35
// -------------------------------------------

package main

import (
	"net"
	"time"
)

// ClientConn 用户连接
type ClientConn struct {
	addTime time.Time    // 创建时间
	conn    *net.TCPConn // tcp连接
}

type Server struct {
	AgentConns  map[string]*net.TCPConn
	TunnelConns map[string]*net.TCPConn
	ClientConns map[string]*ClientConn
}

func (server *Server) run() {

}
