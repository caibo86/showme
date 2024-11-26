// -------------------------------------------
// @file      : config.go
// @author    : bo cai
// @contact   : caibo923@gmail.com
// @time      : 2024/11/25 下午6:18
// -------------------------------------------

package main

import (
	"github.com/caibo86/cberrors"
	"github.com/caibo86/config"
)

const AgentConfigType = "agent"

type Config struct {
	ServerAddr  string `yaml:"serverAddr"`
	TunnelAddr  string `yaml:"tunnelAddr"`
	ServiceAddr string `yaml:"serviceAddr"`
	TunnelLimit int    `yaml:"tunnelLimit"`
}

func (c *Config) GetType() string {
	return AgentConfigType
}

// GetConfig 获取配置
func GetConfig() *Config {
	ic := config.Get(AgentConfigType)
	if ic == nil {
		cberrors.Panic("unable to find config:%s", AgentConfigType)
		return nil
	}
	c, ok := ic.(*Config)
	if !ok {
		cberrors.Panic("invalid type config:%s ", AgentConfigType)
		return nil
	}
	return c
}
