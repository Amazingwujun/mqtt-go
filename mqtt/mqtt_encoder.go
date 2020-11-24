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
		buf[2] = byte(variableHeader.PackageId >> 8)
		buf[3] = byte(variableHeader.PackageId)
		return buf
	case PUBREC:
	case PUBREL:
	case PUBCOMP:
	case SUBACK:
	case UNSUBACK:
	case PINGRESP:
		buf := make([]byte, 2)
		buf[0] = PINGRESP << 4
		return buf
	}

	return nil
}