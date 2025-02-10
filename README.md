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
bash


复制编辑
go get github.com/vpxuser/proxy
```

## Usage

Here’s a simple example of how to use the package:

### Start HTTP/HTTPS MITM Proxy

```go
//todo
```

### Configure TLS Handshake and Certificates

`proxy` supports using custom certificates for TLS MITM. You can provide your own CA certificate and private key:

```go
//todo
```

### WebSocket MITM Proxy

The package also supports WebSocket protocol MITM. You can easily intercept and modify WebSocket messages.

```go
go复制编辑proxy := proxy.NewProxy()

//todo
```

### TCP MITM Proxy

For TCP traffic, the `proxy` package allows you to intercept and modify any TCP-based protocol traffic.

```go
//todo
```

## Configuration Options

- **Listen Address and Port**: Use `proxy.Listen(address)` to configure the address and port to listen on.
- **Certificates and Keys**: Use `proxy.SetCertificate(certPath, keyPath)` to configure the certificates for TLS MITM.
- **Protocol Support**: Enable specific protocols (HTTP, HTTPS, WebSocket, TCP) using methods like `proxy.EnableProtocol(protocol)`.
- **Logging**: The package logs events to the console by default, but custom logging can be configured.

## Man-in-the-Middle Attack Workflow

This tool works by acting as a middleman between the client and the target server. It intercepts, parses, and modifies the traffic. For HTTPS and TLS traffic, it uses a self-signed certificate to generate encrypted connections and decrypts/encrypts traffic using symmetric keys. For TCP and WebSocket protocols, it directly forwards data while optionally modifying it.

## Notes

1. **For Legal Security Testing Only**: Please use this tool only for authorized security testing. Never use it for malicious purposes.
2. **MITM Limitations**: Some websites or services may use certificate pinning, HSTS, and other mechanisms to prevent MITM attacks.
3. **Performance**: The proxy may introduce some latency due to the processing of traffic. The performance depends on the amount of traffic and the complexity of processing logic.

## Example Projects

If you want to see real-world examples of how to use this package, check out the following projects:

- [MITM Proxy Example](https://github.com/vpxuser/proxy/example)

## Contributing

We welcome contributions via issues and pull requests! Any contributions to improve this tool are highly appreciated.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

------

You can adjust the package details and examples as needed. If you need more specific changes or further information, feel free to ask!