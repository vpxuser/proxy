# Proxy Go Package

`proxy` is a Go package designed to create a man-in-the-middle (MITM) proxy tool for intercepting, modifying, and forwarding traffic over multiple protocols including HTTP, HTTPS, TLS, WebSocket, and TCP. This package is ideal for security testing, traffic analysis, and protocol research.

## Features

This package supports the following features:

- **HTTP/HTTPS MITM**: Intercepts and modifies HTTP and HTTPS requests/responses, supporting TLS decryption and encryption.
- **TLS MITM**: Handles TLS handshake, decrypts and re-encrypts traffic for protocol analysis and tampering.
- **WebSocket MITM**: Supports WebSocket connections, intercepting and modifying WebSocket protocol upgrade requests and data transmission.
- **TCP MITM**: Intercepts and modifies raw TCP traffic, useful for any TCP-based protocol.
- **Protocol Support**: MITM attacks for HTTP/HTTPS, WebSocket, TLS, and TCP protocols.

## Installation

Make sure your Go environment is correctly set up, then install the package via the following command:

```shell
go get github.com/vpxuser/proxy
```

## Usage

Here’s a simple example of how to use the package:

### Start HTTP/HTTPS MITM Proxy

```go
package main

import "github.com/vpxuser/proxy"

func main() {
	config := proxy.NewHttpProxy()
	config.Port = "8080"
	config.Serve(&proxy.Manual{})
}
```

### Configure TLS Handshake and Certificates

`proxy` supports using custom certificates for TLS MITM. You can provide your own CA certificate and private key:

```go
func loadCert(config *proxy.HttpProxy) {
	cert, err := os.ReadFile("config/ca.crt")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(cert)

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	config.Cert = certificate
}

func loadPrivateKey(config *proxy.HttpProxy) {
	key, err := os.ReadFile("config/ca.key")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(key)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	config.Key = privateKey
}
```

### HTTP MITM Proxy
You can modify HTTP Request and Response

```go
func snifferHTTP(config *proxy.HttpProxy) {
	config.OnRequest().Do(func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
		reqRaw, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Error(err)
			return req, nil
		}
		log.Debugf("HTTP Request : \n%s", reqRaw)
		return req, nil
	})

	config.OnResponse().Do(func(resp *http.Response, ctx *proxy.Context) *http.Response {
		respRaw, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Error(err)
			return resp
		}
		log.Debugf("HTTP Response : \n%s", respRaw)
		return resp
	})
}
```

### WebSocket MITM Proxy

The package also supports WebSocket protocol MITM. You can easily intercept and modify WebSocket messages.

```go
func snifferWebSocket(config *proxy.HttpProxy) {
	config.OnWebSocket().Do(func(frame ws.Frame, reverse bool, ctx *proxy.Context) ws.Frame {
		log.Debugf("WebSocket Frame Payload : %s", frame.Payload)
		return frame
	})
}
```

### TCP MITM Proxy

For TCP traffic, the `proxy` package allows you to intercept and modify any TCP-based protocol traffic.

```go
func snifferTCPRaw(config *proxy.HttpProxy) {
	config.OnRaw().Do(func(raw []byte, reverse bool, ctx *proxy.Context) []byte {
		log.Debugf("TCP Raw : %s", raw)
		return raw
	})
}
```

## Configuration Options

- **Listen Address and Port**: Use `proxy.HttpProxy.Port` to configure the address and port to listen on.
- **Certificates and Keys**: Use `proxy.HttpProxy.Cert` and `proxy.HttpProxy.Key` to configure the certificates and private key for TLS MITM.
- **Logging**: The package logs events to the console by default, but custom logging can be configured.

## Man-in-the-Middle Attack Workflow

This tool works by acting as a middleman between the client and the target server. It intercepts, parses, and modifies the traffic. For HTTPS and TLS traffic, it uses a self-signed certificate to generate encrypted connections and decrypts/encrypts traffic using symmetric keys. For TCP and WebSocket protocols, it directly forwards data while optionally modifying it.

## Notes

1. **For Legal Security Testing Only**: Please use this tool only for authorized security testing. Never use it for malicious purposes.
2. **MITM Limitations**: Some websites or services may use certificate pinning, HSTS, and other mechanisms to prevent MITM attacks.
3. **Performance**: The proxy may introduce some latency due to the processing of traffic. The performance depends on the amount of traffic and the complexity of processing logic.

## Example Projects

If you want to see real-world examples of how to use this package, check out the following projects:

- [MITM Proxy Example Tool](https://github.com/vpxuser/proxy/examples/tool)

## Contributing

We welcome contributions via issues and pull requests! Any contributions to improve this tool are highly appreciated.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

------

You can adjust the package details and examples as needed. If you need more specific changes or further information, feel free to ask!