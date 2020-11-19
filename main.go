package main

import (
	"encoding/binary"
	"log"
	"mqtt-go/codec"
	"mqtt-go/handler"
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
	// todo start hook

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("连接建立失败: %s", err.Error())
			continue
		}

		log.Printf("remote: [%s]", conn.RemoteAddr().String())

		tcpConn, _ := conn.(*net.TCPConn)
		handleNewConn(tcpConn)
	}
}

func handleNewConn(conn *net.TCPConn) {
	defer conn.Close()

	// 解码消息传递 channel
	c := make(chan interface{}, 10)
	//c1 := make(chan bool)
	go handler.ChannelRead(c)
	//go idleHandler(c1, conn)

	// 开始解码
	var cumulation []byte
	for {
		// 读 buf
		buf := make([]byte, 512)

		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("连接断开: %s\n", err.Error())
			return
		}
		// 收到消息
		//c1 <- true

		// 可用 slice
		buf = buf[:n]
		if len(cumulation) > 0 {
			buf = append(cumulation, buf...)
		}

		left, err := codec.Decode(buf, c)
		if err != nil {
			log.Printf("解码错误: %s\n", err.Error())
			return
		}
		if left != nil && len(left) > 0 {
			cumulation = left
		} else {
			cumulation = make([]byte, 0)
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

// tcp 流拆包
func decode(msg []byte, c chan<- interface{}) ([]byte, error) {
	// 数据缓存大小
	bufLen := uint32(len(msg))
	if bufLen < 4 {
		return msg, nil
	}

	// 消息体长度
	msgLen := binary.BigEndian.Uint32(msg)
	if (bufLen - 4) < msgLen {
		return msg, nil
	}

	// 开始解码
	message := msg[:msgLen+4]
	c <- message

	// 检查数据流状态
	cumulation := msg[msgLen+4:]

	return decode(cumulation, c)
}
