package handler

import (
	"fmt"
	"log"
	"mqtt-go/src/channel"
	"mqtt-go/src/message"
	"mqtt-go/src/store"
	"sync"
	"time"
)

// clientId <==> channelId 映射
var ClientChannelMap sync.Map

// 处理 conn 报文
func HandleConn(channel *channel.Channel, msg *message.MqttMessage) {
	variableHeader := msg.VariableHeader.(*message.MqttConnVariableHeader)
	payload := msg.Payload.(*message.MqttConnPayload)

	// todo 认证

	// client 关联 channel
	channel.SaveClientId(payload.ClientId)

	// 保存 client 与 channelId 的映射
	ClientChannelMap.Store(payload.ClientId, channel.Id)

	// keepalive
	if v := float64(variableHeader.KeepAlive) * 1.5; v > 0 {
		channel.Heartbeat = time.Duration(v)

		// 提醒心跳变更
		channel.InputNotify <- channel.Heartbeat
	}

	connAck := message.BuildConnAck(false, 0)
	channel.Write(connAck)

	// 由于 mqtt-go 不支持消息持久化，所以不存在补发消息的逻辑
}

// 处理 conn 报文
func HandlePub(channel0 *channel.Channel, msg *message.MqttMessage) {
	variableHeader := msg.VariableHeader.(*message.MqttPublishVaribleHeader)
	payload := msg.Payload.([]byte)

	log.Printf("消息id:%d topic: %s 内容:%s\n", variableHeader.MessageId, variableHeader.TopicName, msg.Payload)
	clients := store.Store.Search(variableHeader.TopicName)

	switch msg.FixedHeader.Qos {
	case 0:
		for _, clientSub := range clients {

			// qos 处理
			qos := clientSub.Qos
			if qos > msg.FixedHeader.Qos {
				qos = msg.FixedHeader.Qos
			}

			pubMsg := message.PubMsg{
				Topic:          variableHeader.TopicName,
				Qos:            qos,
				SessionPresent: false,
				Payload:        payload,
			}

			publish0(channel0, &pubMsg)
		}
	case 1:
		for _, clientSub := range clients {

			// qos 处理
			qos := clientSub.Qos
			if qos > msg.FixedHeader.Qos {
				qos = msg.FixedHeader.Qos
			}

			pubMsg := message.PubMsg{
				Topic:          variableHeader.TopicName,
				Qos:            qos,
				SessionPresent: false,
				Payload:        payload,
			}

			// 发布消息
			publish0(channel0, &pubMsg)
		}

		ack := message.BuildPubAck(variableHeader.MessageId)
		channel0.Write(ack)
	case 2:
		for _, clientSub := range clients {

			// qos 处理
			qos := clientSub.Qos
			if qos > msg.FixedHeader.Qos {
				qos = msg.FixedHeader.Qos
			}

			pubMsg := message.PubMsg{
				Topic:          variableHeader.TopicName,
				Qos:            qos,
				SessionPresent: false,
				Payload:        payload,
			}

			// 发布消息
			publish0(channel0, &pubMsg)
		}

		ack := message.BuildPubRec(variableHeader.MessageId)
		channel0.Write(ack)
	default:
		panic(fmt.Sprintf("非法的 Qos:%d\n", msg.FixedHeader.Qos))
	}
}

// 发布消息
func publish0(channel0 *channel.Channel, msg *message.PubMsg) {
	var pubMsg *message.MqttMessage
	if msg.Qos == 0 {
		pubMsg = message.BuildPublish(false, false, 0, msg.Topic, 0, msg.Payload)
	} else {
		messageId := channel0.NextMessageId()

		// 构建 pubMsg
		pubMsg = message.BuildPublish(false, false, msg.Qos, msg.Topic, channel0.NextMessageId(), msg.Payload)

		// 保存 qos1/qos2 消息
		channel0.SavePubMsg(messageId, msg)
	}

	// 发送消息
	if channelId, ok := ClientChannelMap.Load(channel0.ClientId()); ok {
		if c, ok := ChannelGroup.Load(channelId); ok {
			c.(*channel.Channel).Write(pubMsg)
		}
	}
}

// 处理 conn 报文
func HandlePubAck(channel *channel.Channel, msg *message.MqttMessage) {
	variableHeader := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)

	// 移除 pubMsg
	channel.RemovePubMsg(variableHeader.MessageId)
}

// 处理 conn 报文
func HandlePubRec(channel *channel.Channel, msg *message.MqttMessage) {

}

// 处理 conn 报文
func HandlePubRel(channel *channel.Channel, msg *message.MqttMessage) {
	header := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)

	log.Printf("收到 PUBREL 消息, id:%d\n", header.MessageId)

	ack := message.BuildPubComp(header.MessageId)
	channel.Write(ack)
}

//
func HandlePubCom(channel *channel.Channel, msg *message.MqttMessage) {

}

// 订阅
func HandleSub(channel *channel.Channel, msg *message.MqttMessage) {
	header := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)
	payload := msg.Payload.(*message.MqttSubscribePayload)

	// 订阅
	store.Store.Subscribe(channel.ClientId(), payload.Topics...)

	// 响应
	resp := make([]byte, 0, 1)
	for _, topic := range payload.Topics {
		resp = append(resp, topic.Qos)
	}
	ack := message.BuildSubAck(header.MessageId, resp)
	channel.Write(ack)
}

// 解除订阅
func HandleUnSub(channel *channel.Channel, msg *message.MqttMessage) {
	header := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)
	payload := msg.Payload.([]string)

	// 移除订阅
	store.Store.RemoveSub(channel.ClientId(), payload...)

	// 响应
	ack := message.BuildUnsubAck(header.MessageId)
	channel.Write(ack)
}

// 心跳报文
func HandlePingReq(channel *channel.Channel, msg *message.MqttMessage) {
	ack := message.BuildPingAck()
	channel.Write(ack)
}

// 连接断开
func HandleDisconnect(channel *channel.Channel, msg *message.MqttMessage) {
	if err := channel.Close(); err != nil {
		log.Printf("连接关闭异常: %v\n", err)
	}
}
