package config

import (
	"github.com/bytedance/gopkg/util/logger"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	k    = koanf.New(".")
	path = "config.yaml"
	Conf Config
)

type Config struct {
	HostName              string `yaml:"HostName"`
	AllowInsecureClientId bool   `yaml:"AllowInsecureClientId"`
}

func Init() {
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		logger.Fatalf("error loading config: %v", err)
	}
	if err := k.Unmarshal("", &Conf); err != nil {
		logger.Fatalf("error unmarshalling config: %v", err)
	}
}
