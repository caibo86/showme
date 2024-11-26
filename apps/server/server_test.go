// -------------------------------------------
// @file      : server_test.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/25 下午3:05
// -------------------------------------------

package main

import (
	"fmt"
	"testing"
)

func TestServer_addAgentConn(t *testing.T) {
	a := "aaaa\n"
	b := "aaaa" + "\n"
	fmt.Println(a == b)
}
