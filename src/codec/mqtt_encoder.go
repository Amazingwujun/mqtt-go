package codec

import (
	"fmt"
	"mqtt-go/src/message"
	"mqtt-go/src/utils"
)

func Encode(msg *message.MqttMessage) []byte {
	fixedHeader := msg.FixedHeader
	switch fixedHeader.MessageType {
	case message.CONNACK:
		return encodeConnAck(msg)
	case message.PUBLISH:
		return encodePublish(msg)
	case message.PUBACK:
		fallthrough
	case message.PUBREC:
		fallthrough
	case message.PUBREL:
		fallthrough
	case message.UNSUBACK:
		fallthrough
	case message.PUBCOMP:
		variableHeader := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)
		return encodeMessageIdButNoPayload(fixedHeader.MessageType, variableHeader.MessageId)
	case message.SUBACK:
		return encodeSubAck(msg)
	case message.PINGRESP:
		buf := make([]byte, 2)
		buf[0] = fixedHeader.MessageType << 4
		return buf
	default:
		panic(fmt.Sprintf("无法解码的消息类别: %d", fixedHeader.MessageType))
	}
}

func encodeConnAck(msg *message.MqttMessage) []byte {
	buf := make([]byte, 2, 4)

	// 单字节固定头
	buf[0] = message.CONNACK << 4

	// 长度固定
	buf[1] = 2

	connAckVariableHeader := msg.VariableHeader.(*message.MqttConnAckVariableHeader)
	bytes := connAckVariableHeader.ToBytes()
	return append(buf, bytes...)
}

func encodePublish(msg *message.MqttMessage) []byte {
	fixedHeader := msg.FixedHeader
	payload := msg.Payload.([]byte)
	payloadLen := len(payload)

	// 可变头
	mpvh := msg.VariableHeader.(*message.MqttPublishVaribleHeader)
	topicLen := len(mpvh.TopicName)
	mpvhLen := 2 + topicLen + 2
	if fixedHeader.Qos == 0 {
		// qos0 没有 messageId
		mpvhLen = 2 + topicLen
	}

	// remaining length
	lenBuf := utils.EncodeRemainLength(payloadLen + mpvhLen)

	// 最终结果
	buf := make([]byte, 1, 1+len(lenBuf)+mpvhLen+payloadLen)

	// fixedHeader
	buf[0] = message.PUBLISH<<4 + fixedHeader.Qos<<1
	if fixedHeader.Dup {
		buf[0] += 0b1000
	}
	if fixedHeader.Retain {
		buf[0] += 1
	}
	buf = append(buf, lenBuf...)

	// variableHeader
	buf = append(buf, byte(topicLen>>8), byte(topicLen))
	buf = append(buf, []byte(mpvh.TopicName)...)
	if fixedHeader.Qos != 0 {
		buf = append(buf, byte(mpvh.MessageId>>8), byte(mpvh.MessageId))
	}

	// payload
	return append(buf, payload...)
}

func encodeSubAck(msg *message.MqttMessage) []byte {
	variableHeader := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)
	resp := msg.Payload.([]byte)

	// 计算出 remaining length
	lenBuF := utils.EncodeRemainLength(2 + len(resp))

	// 推测 lenBuf 长度不超过 1
	buf := make([]byte, 1, 1+len(lenBuF)+2+len(resp))
	buf[0] = message.SUBACK << 4

	// remaining length
	buf = append(buf, lenBuF...)
	buf = append(buf, byte(variableHeader.MessageId>>8), byte(variableHeader.MessageId))
	buf = append(buf, resp...)

	return buf
}

func encodeMessageIdButNoPayload(messageType byte, messageId uint16) []byte {
	buf := make([]byte, 4, 4)
	buf[0] = messageType << 4
	buf[1] = 2
	buf[2] = byte(messageId >> 8)
	buf[3] = byte(messageId)
	return buf
}
