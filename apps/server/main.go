// -------------------------------------------
// @file      : main.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/15 下午4:56
// -------------------------------------------

package main

import (
	"github.com/caibo86/cberrors"
	"github.com/caibo86/config"
	"github.com/caibo86/logger"
	"github.com/caibo86/misc"
)

func main() {
	config.Load(misc.GetPathInRootDir("config/server.yaml"), &Config{})
	if GetConfig() == nil {
		cberrors.Panic("config is nil")
		return
	}
	logger.Init()
	defer func() {
		_ = logger.Close()
	}()
	server := NewServer()
	server.run()
}
