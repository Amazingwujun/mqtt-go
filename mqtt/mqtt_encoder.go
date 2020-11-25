package mqtt

func Encode(msg *MqttMessage) []byte {
	fixedHeader := msg.FixedHeader
	switch fixedHeader.MessageType {
	case CONNACK:
		buf := make([]byte, 2, 4)

		// 单字节固定头
		buf[0] = CONNACK << 4

		// 长度固定
		buf[1] = 2

		connackVariableHeader := msg.VariableHeader.(*MqttConnackVariableHeader)
		bytes := connackVariableHeader.toBytes()
		return append(buf, bytes...)
	case PUBLISH:
		panic("待完成")
	case PUBACK:
		fallthrough
	case PUBREC:
		fallthrough
	case PUBREL:
		fallthrough
	case UNSUBACK:
		fallthrough
	case PUBCOMP:
		variableHeader := msg.VariableHeader.(*MqttMessageIdVariableHeader)
		return encodeMessageIdButNoPayload(fixedHeader.MessageType, variableHeader.MessageId)
	case SUBACK:
		variableHeader := msg.VariableHeader.(*MqttMessageIdVariableHeader)
		resp := msg.Payload.([]byte)

		// 计算出 remaining length
		lenBuF := encodeMqttBytes(2 + len(resp))

		// 推测 lenBuf 长度不超过 1
		buf := make([]byte, 1, 1+len(lenBuF)+2+len(resp))
		buf[0] = SUBACK << 4

		// remaining length
		buf = append(buf, lenBuF...)
		buf = append(buf, byte(variableHeader.MessageId>>8), byte(variableHeader.MessageId))
		buf = append(buf, resp...)

		return buf
	case PINGRESP:
		buf := make([]byte, 2)
		buf[0] = PINGRESP << 4
		return buf
	}

	return nil
}

func encodeMessageIdButNoPayload(messageType byte, messageId uint16) []byte {
	buf := make([]byte, 4, 4)
	buf[0] = messageType << 4
	buf[1] = 2
	buf[2] = byte(messageId >> 8)
	buf[3] = byte(messageId)
	return buf
}

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
