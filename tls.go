package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/elazarl/goproxy"
	"math/big"
	"time"
)

type GenTLSConfig func(serverName string) (*tls.Config, error)

func TLSConfigFormCA(cert *x509.Certificate, key *rsa.PrivateKey) GenTLSConfig {
	return func(serverName string) (*tls.Config, error) {
		subConf := goproxy.NewProxyHttpServer()
		//subConf.Tr.DisableCompression = true

		return goproxy.TLSConfigFromCA(&tls.Certificate{
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  key,
		})(serverName, &goproxy.ProxyCtx{
			Proxy: subConf,
		})
	}
}

func TLSConfigFromSelfSigned() GenTLSConfig {
	return func(serverName string) (*tls.Config, error) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
		if err != nil {
			return nil, err
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName: serverName,
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(3 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA:                  true,
			DNSNames:              []string{serverName},
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
		if err != nil {
			return nil, err
		}

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, err
		}

		return &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{cert.Raw},
					PrivateKey:  key,
				},
			},
		}, nil
	}
}

func TLSConfigFrom(cert *x509.Certificate, key any) GenTLSConfig {
	return func(serverName string) (*tls.Config, error) {
		return &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{cert.Raw},
					PrivateKey:  key,
				},
			},
		}, nil
	}
}
