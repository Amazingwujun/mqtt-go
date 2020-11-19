package handler

import (
	"encoding/hex"
	"fmt"
	"log"
	"mqtt-go/channel"
	"net"
)

type InboundHandler interface {

	// tcp 连接建立成功
	channelActive(channel <-chan *channel.Channel)

	// tcp 连接建立失败
	channelInactive(channel <-chan *channel.Channel)

	// 收到消息
	channelRead(channel <-chan interface{})
}

type DefaultInboundHandler struct {
}

func ChannelActive(channel <-chan net.Conn) {
	// 阻塞至此
	conn := <-channel

	// todo
	fmt.Printf("连接建立：[%s]\n", conn.RemoteAddr())
}

func ChannelInactive(channel <-chan bool) {
	panic("implement me")
}

func ChannelRead(channel <-chan interface{}) {
	for {
		msg := <-channel

		switch t := msg.(type) {
		case []byte:
			log.Printf("收到报文: %s\n", hex.EncodeToString(t))
		default:
			log.Printf("未知类型: %t\n", t)
		}
	}
}
