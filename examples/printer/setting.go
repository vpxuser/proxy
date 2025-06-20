package main

import (
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"github.com/vpxuser/proxy"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	Cert  string `yaml:"cert"`
	Key   string `yaml:"key"`
	SAN   string `yaml:"san"`
	Proxy string `yaml:"proxy"`
}

var (
	Cfg  = new(Config)
	Cert *x509.Certificate
	Key  *rsa.PrivateKey
)

func init() {
	yml, err := os.ReadFile("config.yaml")
	if err != nil {
		proxy.Fatal(err)
	}

	if err := yaml.Unmarshal(yml, Cfg); err != nil {
		proxy.Fatal(err)
	}

	certPEM, err := os.ReadFile(Cfg.Cert)
	if err != nil {
		proxy.Fatal(err)
	}

	block, _ := pem.Decode(certPEM)
	Cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		proxy.Fatal(err)
	}

	keyPEM, err := os.ReadFile(Cfg.Key)
	if err != nil {
		proxy.Fatal(err)
	}

	block, _ = pem.Decode(keyPEM)
	Key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		proxy.Fatal(err)
	}

	proxy.Debug("Loaded config file successfully")
}
