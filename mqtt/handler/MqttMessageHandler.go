package handler

import "mqtt-go/mqtt"

// 消息处理器接口
type MqttMessageHandler interface {
	Process(ctx *mqtt.Channel, msg mqtt.MqttMessage)

	Support(messageType byte) bool
}
