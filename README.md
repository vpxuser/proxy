# proxy
一个致力于实现多协议中间人攻击的工具

- 已实现系统级HTTP代理和Proxifier
- 已实现TCP、HTTP、TLS中间人攻击功能

## 前言

### 安卓APP渗透抓不到包的常见原因

1. 没有正确地将CA证书安装到/system/etc/security/cacerts（即没有将中间人CA证书安装到操作系统证书受信任根目录）
2. 应用设置了SSL Pinning（即只信任应用包下的特定证书）
3. 应用设置了NO_PROXY（即应用不走系统代理）或自行设置了应用层级的代理

### 解决安卓APP渗透抓包困境的办法

#### 安装证书到收信人CA目录

##### 方法一：直接安装（适合UserDebug版本的系统）

- 先使用[remount_to_system.bat](https://github.com/vpxuser/Awesome-Script/blob/main/remount_to_system.bat)脚本重新挂载硬盘到系统盘
  - d：安卓设备ID（DeviceID），通过`adb devices`命令可以获取
  - f：重启参数，可选，有些设备需要禁用安卓固件验证才能挂载成功

```cmd
.\remount_to_system.bat -d [安卓设备ID] -f
```

- 再使用[upload_ca_cert.bat](https://github.com/vpxuser/Awesome-Script/blob/main/upload_ca_cert.bat)脚本上传CA证书到安卓设备
  - d：安卓设备ID（DeviceID），通过`adb devices`命令可以获取
  - c：CA证书文件所在的物理路径
  

```cmd
.\upload_ca_cert.bat -d [安卓设备ID] -c [证书文件路径]
```

##### 方法二：使用面具模块载入（适合真机）

- 使用[MoveCertificate](https://github.com/ys1231/MoveCertificate)模块将用户证书目录的证书移动到系统证书目录

![](./images/move_certificate.png)

##### 方法三：使用frida动态注入（适合没有内存动态防护的应用）

- 使用[inject_ca_certificate.js](https://github.com/vpxuser/Awesome-Script/blob/main/inject_ca_certificate.js)脚本将证书注入到安卓APP，脚本运行前，请自行将证书文件名修改为ca.crt或打开脚本修改证书路径

```cmd
frida -U -f [APK包名] -l [脚本文件路径]
```

#### 取消证书锁定

##### 方法一：使用frida动态注入（适合没有内存动态防护的应用）

- 使用[bypass_ssl_pinning.js](https://github.com/vpxuser/Awesome-Script/blob/main/bypass_ssl_pinning.js)脚本解除证书锁定

```cmd
frida -U -f [APK包名] -l [脚本文件路径]
```

##### 方法二：使用面具模块载入（适合真机）

- 使用[JustTrustMe](https://github.com/Fuzion24/JustTrustMe)模块解除证书锁定

![](./images/just_trust_me.png)

#### 使用强制代理

##### 方法一：使用具有透明代理功能的代理应用

- 使用[proxy](https://github.com/vpxuser/proxy)配合透明代理工具强抓TCP流量，如：[Proxifier](https://www.proxifier.com/download/#android-tab)

```cmd
.\proxy.exe
```

##### 方法二：使用frida动态注入（适合没有内存动态防护的应用）

- 使用[force_use_proxy.js](https://github.com/vpxuser/Awesome-Script/blob/main/force_use_proxy.js)脚本让安卓APP强制走代理，代理地址默认为127.0.0.1:8080，如有需求请打开脚本修改默认配置

```cmd
frida -U -f [APK包名] -l [脚本文件路径]
```

##### 方法三：系统命令设置iptables

- 使用[set_iptables_proxy.bat](https://github.com/vpxuser/Awesome-Script/blob/main/set_iptables_proxy.bat)脚本设置透明代理

```cmd
.\set_iptables_proxy.bat set -d [安卓设备ID] -h [代理服务器IP] -p [代理服务器端口]
```

## 编译

- 编译linux可执行文件

```cmd
set GOOS=linux
set GOARCH=amd64
go build -o proxy main.go
```

- 编译windows可执行文件

```cmd
set GOOS=windows
set GOARCH=amd64
go build -o proxy.exe main.go
```

- 编译macOS可执行文件

```cmd
set GOOS=darwin
set GOARCH=amd64
go build -o proxy main.go
```

## 配置

- 在可执行程序目录下创建一个config文件夹，在config文件夹下创建一个config.yml文件，config.yml文件配置参考

```yaml
log:
  # 日志级别 5为debug级
  level: 5
proxy:
  # 代理服务监督地址
  host: 0.0.0.0
  # http代理监听端口
  manualPort: 8080
  # 透明代理监听端口
  transparentPort: 8081
  # 线程控制，默认开启1000个线程
  threads: 1000
  # 上游代理设置，支持http、sock5、socks5h协议
  #upstream: http://127.0.0.1:8081
ca:
  # 代理服务中间人CA证书
  cert: config/ca/ca.crt
  # 代理服务中间人CA私钥
  key: config/ca/ca.key
switch:
  # http嗅探开关
  http: true
  # websocket嗅探开关
  websocket: true
  # tcp嗅探开关
  tcp: true
tls:
  # 默认sni，当客户端clienthello没有设置sni时，会使用这个配置下的默认sni进行设置
  defaultSNI: "*.baidu.com"
```

## 运行

- 打开命令行，并进入可执行程序所在目录，运行可执行程序

```powershell
.\proxy.exe
```

## 代理

- 使用HTTPS代理客户端配置代理，这里使用Proxifier做演示

![proxifier配置](./images/1.png)

- 安装抓包工具证书到移动设备或模拟器（注意：需要ROOT权限），这里使用BurpSuite
- 在config.yml文件配置下游代理为BurpSuite代理地址（这里使用BurpSuite默认地址http://127.0.0.1:8080）
- BurpSuite通过上游代理获取到HTTP报文，抓包成功

![burp抓包](./images/2.png)

## 其他

如有疑问，请在Issues提出
