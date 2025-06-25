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

// TLSConfig defines an interface for dynamically generating a *tls.Config
// based on the provided Server Name (SAN).
// TLSConfig 定义了一个接口，用于根据传入的服务名称（SAN）动态生成 *tls.Config。
type TLSConfig interface {
	From(string) (*tls.Config, error)
}

// TLSConfigFn is a function adapter that implements the TLSConfig interface.
// TLSConfigFn 是一个函数适配器，实现了 TLSConfig 接口。
type TLSConfigFn func(string) (*tls.Config, error)

// From calls the function itself to generate a *tls.Config.
// From 会调用函数本体来生成 *tls.Config。
func (f TLSConfigFn) From(san string) (*tls.Config, error) { return f(san) }

////////////////////////////////////////////////////////////////////////////////

// FromCA creates a TLSConfigFn using a CA certificate and private key,
// which generates leaf certificates signed by the CA for the specified SAN.
// FromCA 使用 CA 证书和私钥返回一个 TLSConfigFn，
// 会为指定的 SAN 签发由 CA 签名的服务端证书。
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

////////////////////////////////////////////////////////////////////////////////

// FromSelfSigned generates a self-signed certificate for the given SAN and returns
// a tls.Config containing that certificate.
// FromSelfSigned 会为指定 SAN 创建一个自签名证书，并返回包含该证书的 tls.Config。
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

////////////////////////////////////////////////////////////////////////////////

// From returns a static tls.Config using the given certificate and private key.
// From 使用传入的证书和私钥构造并返回一个固定的 tls.Config。
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
