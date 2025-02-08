package main

import (
	"example/handler"
	"example/setting"
	"github.com/vpxuser/proxy"
	yaklog "github.com/yaklang/yaklang/common/log"
)

func init() {
	//fmt.Println(PROGRAM_NAME)
	yaklog.SetLevel(setting.Config.Log.Level)
}

func main() {
	httpProxy := proxy.NewHttpProxy()
	httpProxy.Host = setting.Config.Proxy.Host
	httpProxy.Threads = setting.Config.Proxy.Threads
	httpProxy.Cert = setting.Cert
	httpProxy.Key = setting.Key
	httpProxy.DefaultSNI = setting.Config.TLS.DefaultSNI

	if setting.Config.Proxy.Upstream != "" {
		dialer, err := proxy.FromURL(httpProxy.HTTPClient, setting.Config.Proxy.Upstream)
		if err != nil {
			yaklog.Fatalf("set upstream proxy failed - %v", err)
		}
		httpProxy.Dialer = dialer
	}

	if setting.Config.Switch.HTTP {
		// 请求处理
		httpProxy.OnRequest().Do(handler.DumpRequest)

		// 响应处理
		httpProxy.OnResponse().Do(handler.DumpResponse)
	}

	if setting.Config.Switch.WebSocket {
		httpProxy.OnWebSocket().Do(handler.DumpWebSocket)
	}

	if setting.Config.Switch.TCP {
		httpProxy.OnRaw().Do(handler.DumpRaw)
	}

	manual := httpProxy.Copy(proxy.MODE_ALL)
	manual.Port = setting.Config.Proxy.ManualPort

	go manual.Serve(&proxy.Manual{})

	transparent := httpProxy.Copy(proxy.MODE_ALL)
	transparent.Port = setting.Config.Proxy.TransparentPort

	go transparent.Serve(&proxy.Transparent{})

	select {}
}
