// -------------------------------------------
// @file      : server.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/21 下午6:35
// -------------------------------------------

package main

import (
	"fmt"
	"github.com/caibo86/cberrors"
	"github.com/caibo86/logger"
	"io"
	"net"
	"showme/network"
	"sync"
	"time"
)

// ClientConn 用户连接
type ClientConn struct {
	addTime time.Time    // 创建时间
	conn    *net.TCPConn // tcp连接
}

// Server 公网服务端
type Server struct {
	Config      *Config // 配置
	AgentConns  map[string]*net.TCPConn
	AgentLock   sync.Mutex
	TunnelConns map[string]*net.TCPConn
	TunnelLock  sync.Mutex
	ClientConns map[string]*ClientConn
}

// NewServer 创建服务器
func NewServer() *Server {
	return &Server{
		Config:      GetConfig(),
		AgentConns:  make(map[string]*net.TCPConn),
		TunnelConns: make(map[string]*net.TCPConn),
		ClientConns: make(map[string]*ClientConn),
	}
}

// 启动服务器
func (server *Server) run() {
	logger.Infof("server start")
	go server.keepAgentAlive()
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
	listener, err := network.CreateTCPListener(server.Config.AgentAddr)
	if err != nil {
		cberrors.PanicWrap(err)
		return
	}
	logger.Infof("start listen agents on %s", server.Config.AgentAddr)
	defer func() {
		_ = listener.Close()
	}()
	for {
		fmt.Println("进来了吗")
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			logger.Errorf("agent accept err %s", err)
			continue
		}
		fmt.Println("有连接进来吗")
		logger.Infof("agent accepted %s", conn.RemoteAddr())
		server.addAgentConn(conn)
	}
}

// 添加代理连接
func (server *Server) addAgentConn(conn *net.TCPConn) bool {
	server.AgentLock.Lock()
	defer server.AgentLock.Unlock()
	if len(server.AgentConns) >= server.Config.MaxAgentLimit {
		logger.Errorf("agent connection limit reached %d, abandon %s",
			server.Config.MaxAgentLimit, conn.RemoteAddr())
		_ = conn.Close()
		return false
	}
	server.AgentConns[conn.RemoteAddr().String()] = conn
	return true
}

// 启动客户端监听
func (server *Server) createClientChannel() {
	listener, err := network.CreateTCPListener(server.Config.ClientAddr)
	if err != nil {
		cberrors.PanicWrap(err)
		return
	}
	logger.Infof("start listen clients on %s", server.Config.ClientAddr)
	for {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			logger.Errorf("agent accept err %s", err)
			continue
		}
		logger.Infof("agent accepted %s", conn.RemoteAddr())
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
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, err := io.Copy(client, tunnel)
		if err != nil {
			logger.Errorf("copy tunnel to client err %s", err)
		}
		wg.Done()
	}()
	go func() {
		_, err := io.Copy(tunnel, client)
		if err != nil {
			logger.Errorf("copy client to tunnel err %s", err)
		}
		wg.Done()
	}()
	wg.Wait()
	logger.Infof("close tunnel client %s to %s", client.RemoteAddr(), tunnel.RemoteAddr())
	_ = client.Close()
	server.addTunnelConn(tunnel)
}

// 启动代理隧道监听
func (server *Server) createTunnelChannel() {
	listener, err := network.CreateTCPListener(server.Config.TunnelAddr)
	if err != nil {
		cberrors.PanicWrap(err)
		return
	}
	logger.Infof("start listen tunnels on %s", server.Config.TunnelAddr)
	for {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			logger.Errorf("agent accept err %s", err)
			continue
		}
		logger.Infof("agent accepted %s", conn.RemoteAddr())
		server.addTunnelConn(conn)
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
	var conn *net.TCPConn
	defer func() {
		// 如果有连接,清空一下
		if conn != nil {
			clearTCPConn(conn)
		}
	}()
	server.TunnelLock.Lock()
	defer server.TunnelLock.Unlock()
	for _, c := range server.TunnelConns {
		conn = c
		break
	}
	return conn
}

// 清空一下TCPConn
func clearTCPConn(conn *net.TCPConn) {
	buffer := make([]byte, 1024)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			break
		}
	}
	logger.Infof("clear tcp conn %s", conn.RemoteAddr())
}
