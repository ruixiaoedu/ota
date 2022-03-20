package config

import (
	"gopkg.in/ini.v1"
	"os"
)

type Config struct {
	Keyfile string `ini:"keyfile"` // 密钥地址
}

func NewConfig(filename string) (*Config, error) {
	var cfg Config
	if err := ini.MapTo(&cfg, filename); err != nil {
		if os.IsNotExist(err) {
			// 不存在，设置为空
			return &cfg, nil
		}
		return nil, err
	}
	return &cfg, nil
}
