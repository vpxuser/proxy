package setting

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/kataras/golog"
	yaklog "github.com/yaklang/yaklang/common/log"
	"gopkg.in/yaml.v2"
	"os"
	"tool/utils"
)

const CONFIG_PATH = "config/config.yaml"

var (
	Config Configure
	Cert   *x509.Certificate
	Key    *rsa.PrivateKey
)

type Configure struct {
	Log    Log    `yaml:"log"`
	Proxy  Proxy  `yaml:"proxy"`
	CA     CA     `yaml:"ca"`
	Switch Switch `yaml:"switch"`
	TLS    TLS    `yaml:"tls"`
}

type Log struct {
	Level golog.Level `yaml:"level"`
}

type Proxy struct {
	Host            string `yaml:"host"`
	ManualPort      string `yaml:"manualPort"`
	TransparentPort string `yaml:"transparentPort"`
	ReplayPort      string `yaml:"replayPort"`
	Threads         int    `yaml:"threads"`
	Upstream        string `yaml:"upstream"`
}

type CA struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type Switch struct {
	HTTP      bool `yaml:"http"`
	WebSocket bool `yaml:"websocket"`
	TCP       bool `yaml:"tcp"`
}

type TLS struct {
	DefaultSNI string `yaml:"defaultSNI"`
}

func init() {
	file, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		yaklog.Fatalf("read config failed - %v", err)
	}

	if err = yaml.Unmarshal(file, &Config); err != nil {
		yaklog.Fatalf("parse config failed - %v", err)
	}

	yaklog.Infof("load base config success")

	cert, err := utils.LoadCert(Config.CA.Cert)
	if err != nil {
		yaklog.Fatal(err)
	}
	key, err := utils.LoadKey(Config.CA.Key)
	if err != nil {
		yaklog.Fatal(err)
	}

	Cert, Key = cert, key

	yaklog.Infof("load root ca certificate and privateKey success")
}
