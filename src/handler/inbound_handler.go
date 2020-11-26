package handler

import (
	"mqtt-go/src/channel"
	"mqtt-go/src/message"
	"sync"
)

var ChannelGroup = sync.Map{}

type InboundHandler interface {

	// tcp 连接建立成功
	ChannelActive(channel *channel.Channel)

	// tcp 连接建立失败
	ChannelInactive(channel *channel.Channel)

	// 收到消息
	ChannelRead(channel *channel.Channel, msg *message.MqttMessage)
}

// 默认实现
type DefaultInboundHandler struct {
}

// 单例对象
var inboundHandler = DefaultInboundHandler{}

func FixedInboundHandler() *DefaultInboundHandler {
	return &inboundHandler
}

// tcp 连接建立
func (this *DefaultInboundHandler) ChannelActive(channel *channel.Channel) {
	ChannelGroup.Store(channel.Id, channel)
}

func (this *DefaultInboundHandler) ChannelInactive(channel *channel.Channel) {
	// 移除 channel
	ChannelGroup.Delete(channel.Id)
}

func (this *DefaultInboundHandler) ChannelRead(channel *channel.Channel, msg *message.MqttMessage) {
	messageType := msg.FixedHeader.MessageType
	switch messageType {
	case message.CONNECT:
		HandleConn(channel, msg)
	case message.PUBLISH:
		HandlePub(channel, msg)
	case message.PUBACK:
		HandlePubAck(channel, msg)
	case message.PUBREC:
		HandlePubRec(channel, msg)
	case message.PUBREL:
		HandlePubRel(channel, msg)
	case message.PUBCOMP:
		HandlePubCom(channel, msg)
	case message.SUBSCRIBE:
		HandleSub(channel, msg)
	case message.UNSUBSCRIBE:
		HandleUnSub(channel, msg)
	case message.PINGREQ:
		HandlePingReq(channel, msg)
	case message.DISCONNECT:
		HandleDisconnect(channel, msg)
	}

}
