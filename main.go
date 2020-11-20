package main

import (
	"log"
	"mqtt-go/mqtt"
	"net"
	"time"
)

func main() {
	log.Printf("启动服务...")

	l, err := net.Listen("tcp", ":32960")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("服务器启动完成")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("连接建立失败: %s", err.Error())
			continue
		}

		// 处理
		go handleNewConn(conn)
	}
}

func handleNewConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic info:%v\n", err)
		}
	}()

	log.Printf("remote: [%s]", conn.RemoteAddr().String())
	wrapConn := mqtt.NewChannel(conn)
	wrapConn.Handler.ChannelActive(wrapConn)

	// 释放资源并广播连接断开事件
	defer func() {
		err := wrapConn.Close()
		if err != nil {
			log.Printf("连接关闭异常：%v", err)
		}
		wrapConn.Handler.ChannelInactive(wrapConn)
	}()

	// 开始解码
	var cumulation []byte
	for {
		// 读 buf
		buf := make([]byte, 512)

		n, err := wrapConn.Read(buf)
		if err != nil {
			log.Printf("连接断开: %s\n", err.Error())
			return
		}

		// 可用 slice
		buf = buf[:n]
		if len(cumulation) > 0 {
			buf = append(cumulation, buf...)
		}

		mqttMessage, left, err := mqtt.Decode(buf)
		if err != nil {
			log.Printf("解码错误: %s\n", err.Error())
			return
		}
		if left != nil && len(left) > 0 {
			cumulation = left
		} else {
			cumulation = make([]byte, 0)
		}
		if mqttMessage != nil {
			wrapConn.Handler.ChannelRead(mqttMessage)
		}
	}
}

// 心跳处理
func idleHandler(c chan bool, conn net.Conn) {
	for {
		select {
		case <-time.After(time.Second * 15):
			log.Printf("%s 连接超时\n", conn.RemoteAddr())
			conn.Close()
			return
		case <-c:
			log.Printf("收到消息通知")
		}
	}
}
