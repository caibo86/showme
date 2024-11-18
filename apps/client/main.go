// -------------------------------------------
// @file      : main.go
// @author    : 蔡波
// @contact   : caibo923@gmail.com
// @time      : 2024/11/15 下午4:01
// -------------------------------------------

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"showme/network"
)

var (
	localServerAddr   = "127.0.0.1:8080"   // 本地需要暴露的服务端口
	remoteIP          = "10.23.50.144"     // 远端的IP地址
	remoteControlAddr = remoteIP + ":8009" // 远端的服务控制通道,用来传递控制信息,如出现新连接和心跳
	remoteServerAddr  = remoteIP + ":8008" // 远端服务端口,用来建立隧道
)

func main() {
	tcpConn, err := network.CreateTCPConn(remoteControlAddr)
	if err != nil {
		fmt.Printf("[连接失败] %s %s\n", remoteControlAddr, err)
		return
	}
	fmt.Printf("[连接成功] %s\n", remoteControlAddr)
	reader := bufio.NewReader(tcpConn)
	for {
		var s string
		s, err = reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		// 当有新连接信号出现时,新建一个tcp连接
		if s == network.NewConnection+"\n" {
			go connectLocalAndRemote()
		}
	}
	fmt.Printf("[连接断开] %s\n", remoteControlAddr)
}

func connectLocalAndRemote() {
	local := connectLocal()
	if local == nil {
		return
	}
	remote := connectRemote()
	if remote == nil {
		_ = local.Close()
		return
	}
	network.Join2Conn(local, remote)
}

func connectLocal() *net.TCPConn {
	conn, err := network.CreateTCPConn(localServerAddr)
	if err != nil {
		fmt.Printf("[连接本地服务失败] %s\n", err)
		return nil
	}
	return conn
}

func connectRemote() *net.TCPConn {
	conn, err := network.CreateTCPConn(remoteServerAddr)
	if err != nil {
		fmt.Printf("[连接远端服务失败] %s\n", err)
		return nil
	}
	return conn
}
