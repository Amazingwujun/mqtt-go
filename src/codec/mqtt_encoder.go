package codec

import (
	"fmt"
	"mqtt-go/src/message"
)

func Encode(msg *message.MqttMessage) []byte {
	fixedHeader := msg.FixedHeader
	switch fixedHeader.MessageType {
	case message.CONNACK:
		buf := make([]byte, 2, 4)

		// 单字节固定头
		buf[0] = fixedHeader.MessageType << 4

		// 长度固定
		buf[1] = 2

		connackVariableHeader := msg.VariableHeader.(*message.MqttConnAckVariableHeader)
		bytes := connackVariableHeader.ToBytes()
		return append(buf, bytes...)
	case message.PUBLISH:
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
		lenBuf := encodeMqttBytes(payloadLen + mpvhLen)

		// 最终结果
		buf := make([]byte, 1, 1+len(lenBuf)+mpvhLen+payloadLen)

		// fixedHeader
		buf[0] = fixedHeader.MessageType<<4 + fixedHeader.Qos<<1
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
		buf = append(buf, byte(mpvh.MessageId>>8), byte(mpvh.MessageId))

		// payload
		return append(buf, payload...)
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
		variableHeader := msg.VariableHeader.(*message.MqttMessageIdVariableHeader)
		resp := msg.Payload.([]byte)

		// 计算出 remaining length
		lenBuF := encodeMqttBytes(2 + len(resp))

		// 推测 lenBuf 长度不超过 1
		buf := make([]byte, 1, 1+len(lenBuF)+2+len(resp))
		buf[0] = fixedHeader.MessageType << 4

		// remaining length
		buf = append(buf, lenBuF...)
		buf = append(buf, byte(variableHeader.MessageId>>8), byte(variableHeader.MessageId))
		buf = append(buf, resp...)

		return buf
	case message.PINGRESP:
		buf := make([]byte, 2)
		buf[0] = fixedHeader.MessageType << 4
		return buf
	default:
		panic(fmt.Sprintf("无法解码的消息类别: %d", fixedHeader.MessageType))
	}
}

func encodeMessageIdButNoPayload(messageType byte, messageId uint16) []byte {
	buf := make([]byte, 4, 4)
	buf[0] = messageType << 4
	buf[1] = 2
	buf[2] = byte(messageId >> 8)
	buf[3] = byte(messageId)
	return buf
}

// encode remaining length
func encodeMqttBytes(len int) []byte {
	buf := make([]byte, 0, 4)
	index := 0
	for {
		buf = append(buf, byte(len%128))
		len = len / 128
		if len > 0 {
			buf[index] = buf[index] | 128
		} else {
			return buf
		}
		index++
	}
}
