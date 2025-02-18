package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func LoadCert(path string) (*x509.Certificate, error) {
	certRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certRaw)
	if block != nil {
		certPEM, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		return certPEM, nil
	}
	certDER, err := x509.ParseCertificate(certRaw)
	if err != nil {
		return nil, err
	}
	return certDER, nil
}

func LoadKey(path string) (*rsa.PrivateKey, error) {
	keyRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyRaw)
	if block != nil {
		keyPEM, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return keyPEM, nil
	}
	keyDER, err := x509.ParsePKCS1PrivateKey(keyRaw)
	if err != nil {
		return nil, err
	}
	return keyDER, nil
}

func ReadPriKey(path string, parse func(der []byte) (any, error)) (*rsa.PrivateKey, error) {
	priKeyRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(priKeyRaw)
	if block != nil {
		priKeyPEM, err := parse(block.Bytes)
		if err != nil {
			return nil, err
		}
		return priKeyPEM.(*rsa.PrivateKey), nil
	}
	priKeyDER, err := parse(priKeyRaw)
	if err != nil {
		return nil, err
	}
	return priKeyDER.(*rsa.PrivateKey), nil
}
