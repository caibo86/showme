// -------------------------------------------
// @file      : server.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/21 下午6:35
// -------------------------------------------

package main

import (
	"github.com/caibo86/cberrors"
	"github.com/caibo86/logger"
	"net"
	"showme/network"
	"sync"
	"time"
)

// Server 公网服务端
type Server struct {
	Config      *Config // 配置
	AgentConns  map[string]*net.TCPConn
	AgentLock   sync.Mutex
	TunnelConns map[string]*net.TCPConn
	TunnelLock  sync.Mutex
}

// NewServer 创建服务器
func NewServer() *Server {
	return &Server{
		Config:      GetConfig(),
		AgentConns:  make(map[string]*net.TCPConn),
		TunnelConns: make(map[string]*net.TCPConn),
	}
}

// 启动服务器
func (server *Server) run() {
	logger.Infof("server start")
	go server.keepAgentAlive()
	go server.tunnelStatus()
	go server.createAgentChannel()
	go server.createTunnelChannel()
	server.createClientChannel()
}

func (server *Server) keepAgentAlive() {
	for {
		time.Sleep(time.Second * 30)
		server.AgentLock.Lock()
		logger.Infof("current agent count %d", len(server.AgentConns))
		for k, conn := range server.AgentConns {
			_, err := conn.Write([]byte(network.KeepAlive))
			if err != nil {
				logger.Errorf("agent %s ping err %s", conn.RemoteAddr(), err)
				delete(server.AgentConns, k)
				_ = conn.Close()
			}
		}
		server.AgentLock.Unlock()
	}
}

// 启动代理监听
func (server *Server) createAgentChannel() {
	listener, err := network.TCPListener(server.Config.AgentAddr)
	if err != nil {
		cberrors.PanicWrap(err)
		return
	}
	logger.Infof("start listen agents on %s", server.Config.AgentAddr)
	defer func() {
		_ = listener.Close()
	}()
	for {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			logger.Errorf("agent accept err %s", err)
			continue
		}
		logger.Infof("agent accepted %s", conn.RemoteAddr())
		server.addAgentConn(conn)
	}
}

// 添加代理连接
func (server *Server) addAgentConn(conn *net.TCPConn) bool {
	server.AgentLock.Lock()
	defer server.AgentLock.Unlock()
	if len(server.AgentConns) >= server.Config.AgentLimit {
		logger.Errorf("agent connection limit reached %d, abandon %s",
			server.Config.AgentLimit, conn.RemoteAddr())
		_ = conn.Close()
		return false
	}
	server.AgentConns[conn.RemoteAddr().String()] = conn
	return true
}

// 启动客户端监听
func (server *Server) createClientChannel() {
	listener, err := network.TCPListener(server.Config.ClientAddr)
	if err != nil {
		cberrors.PanicWrap(err)
		return
	}
	logger.Infof("start listen clients on %s", server.Config.ClientAddr)
	for {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			logger.Errorf("client accept err %s", err)
			continue
		}
		logger.Infof("client accepted %s", conn.RemoteAddr())
		go server.tunnelClient(conn)
	}
}

// 给客户端连接选择一条隧道
func (server *Server) tunnelClient(client *net.TCPConn) {
	tunnel := server.getTunnelConn()
	if tunnel == nil {
		logger.Errorf("no tunnel for client %s", client.RemoteAddr())
		_ = client.Close()
		return
	}
	logger.Infof("tunnel client %s to %s", client.RemoteAddr(), tunnel.RemoteAddr())
	network.Join2Conn(client, tunnel)
}

// 启动代理隧道监听
func (server *Server) createTunnelChannel() {
	listener, err := network.TCPListener(server.Config.TunnelAddr)
	if err != nil {
		cberrors.PanicWrap(err)
		return
	}
	logger.Infof("start listen tunnels on %s", server.Config.TunnelAddr)
	for {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			logger.Errorf("tunnel accept err %s", err)
			continue
		}
		logger.Infof("tunnel accepted %s", conn.RemoteAddr())
		go server.addTunnelConn(conn)
	}
}

// 添加隧道连接
func (server *Server) addTunnelConn(conn *net.TCPConn) {
	server.TunnelLock.Lock()
	defer server.TunnelLock.Unlock()
	key := conn.RemoteAddr().String()
	if server.TunnelConns[key] != nil {
		err := server.TunnelConns[key].Close()
		if err != nil {
			logger.Errorf("close tunnel conn %s err %s", key, err)
		}
		delete(server.TunnelConns, key)
	}
	server.TunnelConns[key] = conn
	logger.Infof("add tunnel conn %s", key)
}

// 获取隧道连接
func (server *Server) getTunnelConn() *net.TCPConn {
	logger.Infof("try to get tunnel conn")
	var conn *net.TCPConn
	server.TunnelLock.Lock()
	defer server.TunnelLock.Unlock()
	for key, c := range server.TunnelConns {
		conn = c
		delete(server.TunnelConns, key)
		break
	}
	return conn
}

func (server *Server) tunnelStatus() {
	var i int64
	for {
		time.Sleep(time.Second * 2)
		count := len(server.TunnelConns)
		if i%15 == 0 {
			logger.Infof("current tunnel count %d", count)
		}
		i += 1
		if count < server.Config.TunnelLimit {
			for j := count; j < server.Config.TunnelLimit; j++ {
				server.newClientNtf()
			}
		}

	}
}

// 通知代理有新客户端连接
func (server *Server) newClientNtf() {
	server.AgentLock.Lock()
	defer server.AgentLock.Unlock()
	var conn *net.TCPConn
	for _, c := range server.AgentConns {
		conn = c
		break
	}
	if conn == nil {
		return
	}
	_, err := conn.Write([]byte(network.NewConnection))
	if err != nil {
		logger.Errorf("notify agent new client err %s", err)
	}
}
