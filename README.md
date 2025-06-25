# Proxy Go Package
# Proxy Go 包说明

`proxy` is a Go package designed to create a man-in-the-middle (MITM) proxy tool for intercepting, modifying, and forwarding traffic over multiple protocols including HTTP, HTTPS, TLS, WebSocket, and TCP. This package is ideal for security testing, traffic analysis, and protocol research.  
`proxy` 是一个用于构建中间人攻击（MITM）代理工具的 Go 包，支持拦截、修改和转发多种协议的流量，包括 HTTP、HTTPS、TLS、WebSocket 和 TCP。该包非常适合用于安全测试、流量分析和协议研究。

## Features
## 功能特性

This package supports the following features:  
本包支持以下特性：

- **HTTP/HTTPS MITM**: Intercepts and modifies HTTP and HTTPS requests/responses, supporting TLS decryption and encryption.  
  **HTTP/HTTPS 中间人攻击**：拦截并修改 HTTP 和 HTTPS 请求/响应，支持 TLS 解密与加密。

- **TLS MITM**: Handles TLS handshake, decrypts and re-encrypts traffic for protocol analysis and tampering.  
  **TLS 中间人攻击**：处理 TLS 握手，解密并重新加密流量，便于协议分析和篡改。

- **WebSocket MITM**: Supports WebSocket connections, intercepting and modifying WebSocket protocol upgrade requests and data transmission.  
  **WebSocket 中间人攻击**：支持 WebSocket 连接，拦截并修改协议升级请求和数据传输。

- **TCP MITM**: Intercepts and modifies raw TCP traffic, useful for any TCP-based protocol.  
  **TCP 中间人攻击**：拦截并修改原始 TCP 流量，适用于任意基于 TCP 的协议。

- **Protocol Support**: MITM attacks for HTTP/HTTPS, WebSocket, TLS, and TCP protocols.  
  **协议支持**：支持 HTTP/HTTPS、WebSocket、TLS 和 TCP 协议的中间人攻击。

## Installation
## 安装

Make sure your Go environment is correctly set up, then install the package via the following command:  
确保你的 Go 环境已经正确配置，然后使用以下命令安装该包：

```bash
go get github.com/vpxuser/proxy
```

## Usage
## 使用方法

Here’s a simple example of how to use the package:  
以下是一个使用本包的简单示例：

### Start HTTP/HTTPS MITM Proxy
### 启动 HTTP/HTTPS 中间人代理

```go
package main

import "github.com/vpxuser/proxy"

func main() {
	config := proxy.NewConfig()
	err := proxy.ListenAndServe("tcp","0.0.0.0:1080",config)
	if err != nil {
        proxy.Fatal(err)
	}
}
```

### Configure TLS Handshake and Certificates
### 配置 TLS 握手与证书

`proxy` supports using custom certificates for TLS MITM. You can provide your own CA certificate and private key:  
`proxy` 支持使用自定义证书进行 TLS 中间人攻击。你可以提供自己的 CA 证书和私钥：

```go
certPEM, err := os.ReadFile("config/ca.crt")
if err != nil {
    proxy.Fatal(err)
}

block, _ := pem.Decode(certPEM)

cert, err := x509.ParseCertificate(block.Bytes)
if err != nil {
    proxy.Fatal(err)
}

privateKeyPEM, err := os.ReadFile("config/ca.key")
if err != nil {
    proxy.Fatal(err)
}

block, _ = pem.Decode(privateKeyPEM)

privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
if err != nil {
    proxy.Fatal(err)
}

cfg := proxy.NewConfig()
cfg.WithOptions(
    proxy.WithTLSConfigFn(proxy.FromCA(cert, privateKey)),
)
```

### HTTP MITM Proxy
### HTTP 中间人代理

You can modify HTTP Request and Response  
你可以拦截并修改 HTTP 请求与响应内容：

```go
cfg.WithReqMatcher().Handle(
    func(req *http.Request, ctx *proxy.Context) (*http.Request, *http.Response) {
        dump, err := httputil.DumpRequest(req, true)
        if err != nil {
            ctx.Error(err)
            return req, nil
        }
        ctx.Info(string(dump))
        return req, nil
    })

cfg.WithRespMatcher().Handle(
    func(resp *http.Response, ctx *proxy.Context) *http.Response {
        dump, err := httputil.DumpResponse(resp, true)
        if err != nil {
            ctx.Error(err)
            return resp
        }
        ctx.Info(string(dump))
        return resp
    })
```

### WebSocket MITM Proxy
### WebSocket 中间人代理

The package also supports WebSocket protocol MITM. You can easily intercept and modify WebSocket messages.  
本包也支持 WebSocket 协议的中间人攻击。你可以轻松地拦截和修改 WebSocket 消息：

```go
cfg.WithWsMatcher().Handle(
    func(frame ws.Frame, ctx *proxy.Context) ws.Frame {
        ctx.Info(string(frame.Payload))
        return frame
    })
```

### TCP MITM Proxy
### TCP 中间人代理

For TCP traffic, the `proxy` package allows you to intercept and modify any TCP-based protocol traffic.  
对于 TCP 流量，`proxy` 包允许你拦截并修改任意 TCP 协议的原始数据流：

```go
cfg.WithRawMatcher().Handle(
    func(raw []byte, ctx *proxy.Context) []byte {
        ctx.Info(string(raw))
        return raw
    })
```

## Man-in-the-Middle Attack Workflow
## 中间人攻击工作流程

This tool works by acting as a middleman between the client and the target server. It intercepts, parses, and modifies the traffic. For HTTPS and TLS traffic, it uses a self-signed certificate to generate encrypted connections and decrypts/encrypts traffic using symmetric keys. For TCP and WebSocket protocols, it directly forwards data while optionally modifying it.  
该工具通过充当客户端和目标服务器之间的中间人来工作。它拦截、解析并修改通信流量。对于 HTTPS 和 TLS 流量，它使用自签名证书建立加密连接，并通过对称密钥实现解密与加密；对于 TCP 和 WebSocket 协议，它可以直接转发数据，并支持修改。

## Notes
## 注意事项

1. **For Legal Security Testing Only**: Please use this tool only for authorized security testing. Never use it for malicious purposes.  
   **仅限合法的安全测试用途**：请仅在获得授权的安全测试场景中使用此工具，严禁用于任何恶意目的。

2. **MITM Limitations**: Some websites or services may use certificate pinning, HSTS, and other mechanisms to prevent MITM attacks.  
   **中间人攻击限制**：部分网站或服务可能启用了证书钉扎、HSTS 等机制以防止中间人攻击。

3. **Performance**: The proxy may introduce some latency due to the processing of traffic. The performance depends on the amount of traffic and the complexity of processing logic.  
   **性能考虑**：代理可能因流量处理引入延迟，性能表现与流量量级及处理逻辑复杂度有关。

## Example Projects
## 示例项目

If you want to see real-world examples of how to use this package, check out the following projects:  
如果你希望查看实际项目中如何使用该包，请参考以下示例：

- [MITM Proxy Example Tool](https://github.com/vpxuser/proxy/tree/main/examples/printer)

## Contributing
## 贡献指南

We welcome contributions via issues and pull requests! Any contributions to improve this tool are highly appreciated.  
欢迎通过 issue 和 pull request 提交改进建议！任何能帮助改进本工具的贡献都将不胜感激。

## License
## 开源许可证

This project is licensed under the MIT License. See the LICENSE file for details.  
本项目采用 MIT 开源许可证。详情请查看项目根目录下的 LICENSE 文件。
