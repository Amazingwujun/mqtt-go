// conn 类型报文处理

package handler

import (
	"log"
	"mqtt-go/mqtt"
)

type ConnHandler struct {
}

func (this *ConnHandler) Process(ctx *mqtt.Channel, msg mqtt.MqttMessage) {
	log.Printf("收到 mqtt 报文:%v\n", msg)
}

func (this *ConnHandler) Support(messageType byte) bool {
	panic("implement me")
}
