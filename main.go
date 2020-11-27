package main

import (
	"flag"
	"log"
	"mqtt-go/src/channel"
	"mqtt-go/src/codec"
	"mqtt-go/src/handler"
	"net"
	"time"
)

func main() {
	addr := flag.String("port", ":1884", "指定监听地址")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("监听: %s", l.Addr().String())

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("连接建立失败: %s", err.Error())
			continue
		}

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
	wrapConn := channel.NewChannel(conn)
	handler.ChannelActive(wrapConn)

	// 释放资源并广播连接断开事件
	defer func() {
		err := wrapConn.Close()
		if err != nil {
			log.Printf("连接关闭异常：%v", err)
		}
		log.Printf("客户端[%s]连接断开", wrapConn.Id)
		handler.ChannelInactive(wrapConn)
	}()

	// 启动写入 goroutine
	go startWriter(wrapConn)

	// 开始处理数据流
	startReader(wrapConn)
}

func startReader(channel *channel.Channel) {
	// 开始解码
	var cumulation []byte
	for {
		// 读 buf
		buf := channel.Get()

		n, err := channel.Read(buf)
		if err != nil {
			log.Printf("连接断开: %s\n", err.Error())
			return
		}
		if n == 0 {
			// 回收 buf
			channel.Put(buf)
			continue
		}

		// 可用 slice
		cumulation = append(cumulation, buf[:n]...)

		// 回收 buf
		channel.Put(buf)

		// 开始解码
		for {
			if mqttMessage, left, err := codec.Decode(cumulation); err != nil {
				log.Printf("解码错误: %s\n", err.Error())
				return
			} else {
				if mqttMessage != nil {
					handler.ChannelRead(channel, mqttMessage)
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

func startWriter(channel *channel.Channel) {
	for {
		select {
		case buf := <-channel.Out:
			if _, err := channel.Write0(buf); err != nil {
				log.Printf("写入失败: %s\n", err)
			}
		case <-channel.Stop:
			return
		}
	}
}

// 心跳处理
func idleHandler(c chan bool, channel *channel.Channel) {
	for {
		select {
		case <-time.After(time.Second * 15):
			channel.Close()
			return
		case <-c:
			log.Printf("收到消息通知")
		}
	}
}
