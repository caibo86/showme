// -------------------------------------------
// @file      : agent.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/25 下午6:12
// -------------------------------------------

package main

import (
	"bufio"
	"github.com/caibo86/logger"
	"io"
	"net"
	"showme/network"
)

type Tunnel struct {
	ID     int32        // ID
	Local  *net.TCPConn // 本地服务连接
	Remote *net.TCPConn // 外网服务器连接
}

// Agent 内外代理
type Agent struct {
	*Config // 配置
	Tunnels map[int32]*Tunnel
}

func NewAgent() *Agent {
	ret := &Agent{
		Config:  GetConfig(),
		Tunnels: make(map[int32]*Tunnel),
	}
	return ret
}

func (agent *Agent) run() {
	conn, err := network.CreateTCPConn(agent.Config.ServerAddr)
	if err != nil {
		logger.Errorf("connect to server %s err %s", agent.Config.ServerAddr, err)
		return
	}
	logger.Infof("connect to server %s success", agent.Config.ServerAddr)
	reader := bufio.NewReader(conn)
	for {
		var s string
		s, err = reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		switch s {
		case network.KeepAlive:
			logger.Infof("receive keep alive")
		case network.NewConnection:

		}
	}
	logger.Infof("connect to server %s closed", agent.Config.ServerAddr)
	return
}

func (agent *Agent) createTunnels() {
	for i := 0; i < agent.Config.TunnelLimit; i++ {
		var local, remote *net.TCPConn
		var err error
		local, err = network.CreateTCPConn(agent.Config.ServiceAddr)
		if err != nil {
			logger.Errorf("connect to service %s err %s", agent.Config.ServiceAddr, err)
			return
		}
		remote, err = network.CreateTCPConn(agent.Config.TunnelAddr)
		if err != nil {
			logger.Errorf("connect to server %s err %s", agent.Config.TunnelAddr, err)
			_ = local.Close()
			return
		}
		agent.Tunnels[int32(i)] = &Tunnel{
			ID:     int32(i),
			Local:  local,
			Remote: remote,
		}
		network.Join2Conn(local, remote)
	}
	logger.Infof("create %d tunnels success", len(agent.Tunnels))
}
