// -------------------------------------------
// @file      : main.go
// @author    : 蔡波
// @contact   : caibo923@gmail.com
// @time      : 2024/11/15 下午4:56
// -------------------------------------------

package main

import (
	"fmt"
	"net"
	"showme/network"
	"strconv"
	"sync"
	"time"
)

const (
	controlAddr = "0.0.0.0:8009"
	tunnelAddr  = "0.0.0.0:8008"
	visitAddr   = "0.0.0.0:8007"
)

var (
	clientConn         *net.TCPConn
	connectionPool     map[string]*ConnMatch
	connectionPoolLock sync.Mutex
)

type ConnMatch struct {
	addTime time.Time
	accept  *net.TCPConn
}

func main() {
	connectionPool = make(map[string]*ConnMatch, 32)
	go createControlChannel()
	go acceptUserRequest()
	go acceptClientRequest()
	cleanConnectionPool()
}

// 创建一个控制通道,用于传递控制消息,如:心跳,创建新连接
func createControlChannel() {
	tcpListener, err := network.CreateTCPListener(controlAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[已监听客户端连接] %s\n", controlAddr)
	for {
		var tcpConn *net.TCPConn
		tcpConn, err = tcpListener.AcceptTCP()
		if err != nil {
			fmt.Printf("[客户端连接接收失败] %s\n", err)
			continue
		}
		fmt.Printf("[新客户端连接] %s\n", tcpConn.RemoteAddr().String())
		// 如果当前已经有一个客户端存在,则丢弃这个连接
		if clientConn != nil {
			_ = tcpConn.Close()
			continue
		}
		clientConn = tcpConn
		go keepAlive()
	}
}

// 和客户端保持一个心跳连接
func keepAlive() {
	for {
		if clientConn == nil {
			return
		}
		_, err := clientConn.Write([]byte(network.KeepAlive + "\n"))
		if err != nil {
			fmt.Printf("[客户端心跳失败] %s %s\n", clientConn.RemoteAddr(), err)
			clientConn = nil
			return
		}
		time.Sleep(time.Second * 5)
	}
}

// 监听来自用户的请求
func acceptUserRequest() {
	tcpListener, err := network.CreateTCPListener(visitAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[用户已监听] %s\n", visitAddr)
	defer func() {
		_ = tcpListener.Close()
	}()
	for {
		var tcpConn *net.TCPConn
		tcpConn, err = tcpListener.AcceptTCP()
		if err != nil {
			fmt.Printf("[用户接收失败] %s\n", err)
			continue
		}
		fmt.Printf("[新用户连接] %s\n", tcpConn.RemoteAddr().String())
		addConn2Pool(tcpConn)
		sendMessage(network.NewConnection + "\n")
	}
}

// 将用户连接放入连接池
func addConn2Pool(conn *net.TCPConn) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()
	now := time.Now()
	connectionPool[strconv.FormatInt(now.UnixNano(), 10)] = &ConnMatch{
		addTime: now,
		accept:  conn,
	}
}

// 给客户端发送消息
func sendMessage(msg string) {
	if clientConn == nil {
		fmt.Printf("[客户端未连接] %s\n", msg)
		return
	}
	_, err := clientConn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("[消息发送失败] %s %s\n", msg, err)
	}
}

// 接收客户端来的请求并建立隧道
func acceptClientRequest() {
	tcpListener, err := network.CreateTCPListener(tunnelAddr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[已监听客户端请求] %s\n", tunnelAddr)
	defer func() {
		_ = tcpListener.Close()
	}()
	for {
		var tcpConn *net.TCPConn
		tcpConn, err = tcpListener.AcceptTCP()
		if err != nil {
			fmt.Printf("[客户端请求接收失败] %s\n", err)
			continue
		}
		fmt.Printf("[新客户端请求] %s\n", tcpConn.RemoteAddr().String())
		go establishTunnel(tcpConn)
	}
}

func establishTunnel(tunnel *net.TCPConn) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()
	for key, connMatch := range connectionPool {
		if connMatch.accept != nil {
			go network.Join2Conn(connMatch.accept, tunnel)
			delete(connectionPool, key)
			return
		}
	}
}

func cleanConnectionPool() {
	for {
		connectionPoolLock.Lock()
		for key, connMatch := range connectionPool {
			if time.Since(connMatch.addTime) > time.Second*10 {
				_ = connMatch.accept.Close()
				delete(connectionPool, key)
			}
		}
		connectionPoolLock.Unlock()
		time.Sleep(5 * time.Second)
	}
}
