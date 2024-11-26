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
	AgentAddr      string `yaml:"agentAddr"`      // 代理监听地址
	TunnelAddr     string `yaml:"tunnelAddr"`     // 隧道监听地址
	ClientAddr     string `yaml:"clientAddr"`     // 客户端监听地址
	MaxClientLimit int    `yaml:"maxClientLimit"` // 最大客户端数
	MaxAgentLimit  int    `yaml:"maxAgentLimit"`  // 最大代理数
}

// GetType implements IConfig
func (config *Config) GetType() string {
	return ServerConfigType
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
