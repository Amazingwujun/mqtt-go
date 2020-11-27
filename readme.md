# Project MQTT-GO
`Mqtt-GO` 基于 [MQTT v3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html) 协议，提供一个***完全基于内存*** 的 mqtt broker。



特点：完整实现 [MQTT v3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html) 协议，不支持消息持久化。

> **应用重启会导致 qos1, qos2 消息丢失**



## 快速开始

**windows** 环境下构建：

- linux: `GOOS=linux GOARCH=amd64 go build -o mqtt-go`
- windows: `go build -o mqtt-go`

构建完成后，直接运行二进制包即可(Linux 系统需要赋与 `mqtt-go` 可执行权限，`chmod 744 ./mqtt-go`)



