package mqtt

var ChannelGroup = make(map[string]*Channel, 10000)

type InboundHandler interface {

	// tcp 连接建立成功
	ChannelActive(channel *Channel)

	// tcp 连接建立失败
	ChannelInactive(channel *Channel)

	// 收到消息
	ChannelRead(channel *Channel, msg *MqttMessage)
}

// 默认实现
type DefaultInboundHandler struct {
}

var inboundHandler = DefaultInboundHandler{}

func NewInboundHandler() *DefaultInboundHandler {
	return &inboundHandler
}

// tcp 连接建立
func (this *DefaultInboundHandler) ChannelActive(channel *Channel) {
	ChannelGroup[channel.Id] = channel
}

func (this *DefaultInboundHandler) ChannelInactive(channel *Channel) {
	// 移除 channel
	delete(ChannelGroup, channel.Id)
}

func (this *DefaultInboundHandler) ChannelRead(channel *Channel, msg *MqttMessage) {
	messageType := msg.FixedHeader.MessageType
	switch messageType {
	case CONNECT:
		connack := BuildConnack(false, 0)
		channel.Write(connack)
	case PINGREQ:
		ack := BuildPingAck()
		channel.Write(ack)
	}

}
