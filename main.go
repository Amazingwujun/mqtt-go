package main

import (
	"log"
	"mqtt-go/mqtt"
	"net"
	"time"
)

func main() {
	log.Printf("启动服务...")

	l, err := net.Listen("tcp", ":1884")
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
	wrapConn.InboundHandler.ChannelActive(wrapConn)
	go startWriter(wrapConn) // 启动写入

	// 释放资源并广播连接断开事件
	defer func() {
		err := wrapConn.Close()
		if err != nil {
			log.Printf("连接关闭异常：%v", err)
		}
		wrapConn.InboundHandler.ChannelInactive(wrapConn)
	}()

	// 开始解码
	var cumulation []byte
	for {
		// 读 buf
		buf := wrapConn.Get()

		n, err := wrapConn.Read(buf)
		if err != nil {
			log.Printf("连接断开: %s\n", err.Error())
			return
		}
		if n == 0 {
			// 回收 buf
			wrapConn.Put(buf)
			continue
		}

		// 可用 slice
		cumulation = append(cumulation, buf[:n]...)

		// 回收 buf
		wrapConn.Put(buf)

		// 开始解码
		for {
			if mqttMessage, left, err := mqtt.Decode(cumulation); err != nil {
				log.Printf("解码错误: %s\n", err.Error())
				return
			} else {
				if mqttMessage != nil {
					wrapConn.InboundHandler.ChannelRead(wrapConn, mqttMessage)
					cumulation = left
					continue
				}

				if left != nil && len(left) > 0 {
					cumulation = left
				} else {
					cumulation = make([]byte, 0)
				}
				break
			}
		}
	}
}

func startWriter(channel *mqtt.Channel) {
	for {
		select {
		case buf := <-channel.Out:
			if _, err := channel.Write0(buf); err != nil {
				log.Printf("写入失败: %s\n", err)
			}
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
