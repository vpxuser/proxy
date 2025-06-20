package proxy

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/elazarl/goproxy"
	"math/big"
	"time"
)

type TLSConfig interface {
	From(string) (*tls.Config, error)
}

type TLSConfigFn func(string) (*tls.Config, error)

func (f TLSConfigFn) From(san string) (*tls.Config, error) { return f(san) }

func FromCA(cert *x509.Certificate, privateKey *rsa.PrivateKey) TLSConfigFn {
	return func(san string) (*tls.Config, error) {
		return goproxy.TLSConfigFromCA(&tls.Certificate{
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  privateKey,
		})(san, &goproxy.ProxyCtx{
			Proxy: goproxy.NewProxyHttpServer(),
		})
	}
}

func FromSelfSigned() TLSConfigFn {
	return func(san string) (*tls.Config, error) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
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
				CommonName: san,
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(3 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA:                  true,
			DNSNames:              []string{san},
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
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
					PrivateKey:  privateKey,
				},
			},
		}, nil
	}
}

func From(cert *x509.Certificate, privateKey crypto.PrivateKey) TLSConfigFn {
	return func(san string) (*tls.Config, error) {
		return &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{cert.Raw},
					PrivateKey:  privateKey,
				},
			},
		}, nil
	}
}
