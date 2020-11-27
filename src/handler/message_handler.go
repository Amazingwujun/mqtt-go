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

	connack := message.BuildConnAck(false, 0)
	channel.Write(connack)
}

// 处理 conn 报文
func HandlePub(channel0 *channel.Channel, msg *message.MqttMessage) {
	variableHeader := msg.VariableHeader.(*message.MqttPublishVaribleHeader)
	payload := msg.Payload.([]byte)

	log.Printf("消息id:%d topic: %s 内容:%s\n", variableHeader.MessageId, variableHeader.TopicName, msg.Payload)

	switch msg.FixedHeader.Qos {
	case 0:
		clients := store.Store.Search(variableHeader.TopicName)

		for _, clientSub := range clients {

			// qos 处理
			qos := clientSub.Qos
			if qos > msg.FixedHeader.Qos {
				qos = msg.FixedHeader.Qos
			}

			// 构建 qos
			publish := message.BuildPublish(false, false, clientSub.Qos, variableHeader.TopicName, channel0.NextPackageId(), payload)

			// 发送消息
			if channelId, ok := ClientChannelMap.Load(clientSub.ClientId); ok {
				if c, ok := ChannelGroup.Load(channelId); ok {
					c.(*channel.Channel).Write(publish)
				}
			}
		}
	case 1:
		ack := message.BuildPubAck(variableHeader.MessageId)
		channel0.Write(ack)
	case 2:
		ack := message.BuildPubRec(variableHeader.MessageId)
		channel0.Write(ack)
	default:
		panic(fmt.Sprintf("非法的 Qos:%d\n", msg.FixedHeader.Qos))
	}
}

// 处理 conn 报文
func HandlePubAck(channel *channel.Channel, msg *message.MqttMessage) {

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
