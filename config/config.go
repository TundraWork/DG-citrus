package config

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
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
	Port                  string `yaml:"Port"`
	UseSecureWebsocket    bool   `yaml:"UseSecureWebsocket"`
	AllowInsecureClientId bool   `yaml:"AllowInsecureClientId"`
}

func Init() {
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		hlog.Fatalf("error loading config: %v", err)
	}
	if err := k.Unmarshal("", &Conf); err != nil {
		hlog.Fatalf("error unmarshalling config: %v", err)
	}
}
