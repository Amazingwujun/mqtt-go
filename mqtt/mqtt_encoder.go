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
	case PUBACK:
		variableHeader := msg.VariableHeader.(*MqttMessageIdVariableHeader)

		buf := make([]byte, 4, 4)
		buf[0] = PUBACK << 4
		buf[1] = 2
		buf[2] = byte(variableHeader.MessageId >> 8)
		buf[3] = byte(variableHeader.MessageId)
		return buf
	case PUBREC:
		variableHeader := msg.VariableHeader.(*MqttMessageIdVariableHeader)

		buf := make([]byte, 4, 4)
		buf[0] = PUBREC << 4
		buf[1] = 2
		buf[2] = byte(variableHeader.MessageId >> 8)
		buf[3] = byte(variableHeader.MessageId)
		return buf
	case PUBREL:
		variableHeader := msg.VariableHeader.(*MqttMessageIdVariableHeader)

		buf := make([]byte, 4, 4)
		buf[0] = PUBREL << 4
		buf[1] = 2
		buf[2] = byte(variableHeader.MessageId >> 8)
		buf[3] = byte(variableHeader.MessageId)
		return buf
	case PUBCOMP:
		variableHeader := msg.VariableHeader.(*MqttMessageIdVariableHeader)

		buf := make([]byte, 4, 4)
		buf[0] = PUBCOMP << 4
		buf[1] = 2
		buf[2] = byte(variableHeader.MessageId >> 8)
		buf[3] = byte(variableHeader.MessageId)
		return buf
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
	case UNSUBACK:
	case PINGRESP:
		buf := make([]byte, 2)
		buf[0] = PINGRESP << 4
		return buf
	}

	return nil
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
