// -------------------------------------------
// @file      : config.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/21 下午6:35
// -------------------------------------------

package main

import (
	"github.com/caibo86/cberrors"
	"github.com/caibo86/config"
)

const ServerConfigType = "server"

// Config 配置
type Config struct {
	AgentPort      int32 `yaml:"agentPort"`      // 代理监听端口
	TunnelPort     int32 `yaml:"tunnelPort"`     // 隧道监听端口
	ClientPort     int32 `yaml:"clientPort"`     // 客户端监听端口
	MaxClientLimit int32 `yaml:"maxClientLimit"` // 最大客户端数
	MaxAgentLimit  int32 `yaml:"maxAgentLimit"`  // 最大代理数
}

// GetType implements IConfig
func (config *Config) GetType() string {
	return "ServerConfigType"
}

// GetConfig 获取配置
func GetConfig() *Config {
	ic := config.Get(ServerConfigType)
	if ic == nil {
		cberrors.Panic("unable to find config:%s", ServerConfigType)
		return nil
	}
	c, ok := ic.(*Config)
	if !ok {
		cberrors.Panic("invalid type config:%s ", ServerConfigType)
		return nil
	}
	return c
}
