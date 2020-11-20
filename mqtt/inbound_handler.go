package mqtt

import "log"

var ChannelGroup = make(map[string]*Channel, 10000)

type InboundHandler interface {

	// tcp 连接建立成功
	ChannelActive(channel *Channel)

	// tcp 连接建立失败
	ChannelInactive(channel *Channel)

	// 收到消息
	ChannelRead(msg *MqttMessage)
}

// 默认实现
type DefaultInboundHandler struct {
}

// tcp 连接建立
func (this *DefaultInboundHandler) ChannelActive(channel *Channel) {
	// 阻塞至此
	ChannelGroup[channel.Id] = channel
}

func (this *DefaultInboundHandler) ChannelInactive(channel *Channel) {
	// 移除 channel
	delete(ChannelGroup, channel.Id)
}

func (this *DefaultInboundHandler) ChannelRead(msg *MqttMessage) {
	log.Printf("收到 mqtt 报文:%v\n", msg)
}
