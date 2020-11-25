package mqtt

import (
	"fmt"
	"log"
	"sync"
)

var ChannelGroup = sync.Map{}

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
	ChannelGroup.Store(channel.Id, channel)
}

func (this *DefaultInboundHandler) ChannelInactive(channel *Channel) {
	// 移除 channel
	ChannelGroup.Delete(channel.Id)
}

func (this *DefaultInboundHandler) ChannelRead(channel *Channel, msg *MqttMessage) {
	messageType := msg.FixedHeader.MessageType
	switch messageType {
	case CONNECT:
		connack := BuildConnack(false, 0)
		channel.Write(connack)
	case PUBLISH:
		header := msg.VariableHeader.(*MqttPublishVaribleHeader)

		log.Printf("消息id:%d topic: %s 内容:%s\n", header.MessageId, header.TopicName, msg.Payload)

		switch msg.FixedHeader.Qos {
		case 0:
		case 1:
			ack := BuildPubAck(header.MessageId)
			channel.Write(ack)
		case 2:
			ack := BuildPubRec(header.MessageId)
			channel.Write(ack)
		default:
			panic(fmt.Sprintf("非法的 Qos:%d\n", msg.FixedHeader.Qos))
		}
	case PUBACK:
		// todo 待完成
	case PUBREL:
		header := msg.VariableHeader.(*MqttMessageIdVariableHeader)

		log.Printf("收到 PUBREL 消息, id:%d\n", header.MessageId)

		ack := BuildPubComp(header.MessageId)
		channel.Write(ack)
	case SUBSCRIBE:
		header := msg.VariableHeader.(*MqttMessageIdVariableHeader)

		payload := msg.Payload.(*MqttSubscribePayload)
		resp := make([]byte, 0, 1)
		for _, topic := range payload.Topics {
			resp = append(resp, topic.Qos)
		}

		ack := BuildSubAck(header.MessageId, resp)
		channel.Write(ack)
	case UNSUBSCRIBE:
		header := msg.VariableHeader.(*MqttMessageIdVariableHeader)
		payload := msg.Payload.([]string)

		log.Printf("移除订阅信息: %v\n", payload)

		ack := BuildUnsubAck(header.MessageId)
		channel.Write(ack)
	case PINGREQ:
		ack := BuildPingAck()
		channel.Write(ack)
	case DISCONNECT:
		if err := channel.Close(); err != nil {
			log.Printf("连接关闭异常: %v\n", err)
		}
	}

}
