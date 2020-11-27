package codec

import (
	"errors"
	"fmt"
	"mqtt-go/src/message"
	"mqtt-go/src/utils"
)

// 拆包
func Decode(buf []byte) (*message.MqttMessage, []byte, error) {
	// 最少最少也有两个字节的数据
	bufLen := len(buf)
	if bufLen < 2 {
		return nil, buf, nil
	}

	// 检查 buf 是否最少含有一个完整的消息
	remainingLen, digits, err := utils.DecodeRemainLength(buf[1:])
	if err != nil {
		return nil, nil, err
	}
	if digits == 0 {
		return nil, buf, nil
	}
	mqttMsgLen := 1 + digits + remainingLen
	if bufLen < mqttMsgLen {
		return nil, buf, nil
	}
	// 至此，可以确定 buf 最少含有一个完整数据包

	// 解析固定头
	fixedHeader := &message.MqttFixedHeader{
		MessageType:  buf[0] >> 4,
		Qos:          (buf[0] & 0b0110) >> 1,
		Dup:          ((buf[0] & 0b1000) >> 1) == 1,
		Retain:       (buf[0] & 0b1) == 1,
		RemainLength: remainingLen,
	}

	// 2 variable header
	msg := &message.MqttMessage{
		FixedHeader: fixedHeader,
	}
	switch fixedHeader.MessageType {
	case message.CONNECT:
		if connVariableHeader, err := message.ReadFrom(buf[1+digits:]); err != nil {
			return nil, nil, err
		} else {
			msg.VariableHeader = connVariableHeader
			// conn 类型的报文固定头为 10 个字节
			payload := decodeConnPayload(connVariableHeader, buf[1+digits+10:])

			msg.Payload = payload
			return msg, buf[mqttMsgLen:], nil
		}
	case message.PUBLISH:
		m := new(message.MqttPublishVaribleHeader)
		index, err := m.ParseFrom(buf, fixedHeader.Qos, 1+digits)
		if err != nil {
			return nil, nil, err
		}
		msg.VariableHeader = m

		// 获取 payload
		msg.Payload = buf[index:mqttMsgLen]

		return msg, buf[mqttMsgLen:], nil
	case message.PUBACK:
		fallthrough
	case message.PUBREC:
		fallthrough
	case message.PUBREL:
		fallthrough
	case message.PUBCOMP:
		m := new(message.MqttMessageIdVariableHeader)
		_, err := m.ParseFrom(buf, 1+digits)
		if err != nil {
			return nil, nil, err
		}
		msg.VariableHeader = m

		return msg, buf[mqttMsgLen:], nil
	case message.SUBSCRIBE:
		// Bits 3,2,1 and 0 of the fixed header of the SUBSCRIBE Control Packet
		// are reserved and MUST be set to 0,0,1 and 0 respectively. The Server MUST
		// treat any other value as malformed and close the Network Connection
		// [MQTT-3.8.1-1].
		if fixedHeader.Dup || fixedHeader.Qos != 1 || fixedHeader.Retain {
			return nil, nil, errors.New("SUBSCRIBE 报文格式非法")
		}

		// 可变头
		m := new(message.MqttMessageIdVariableHeader)
		index, err := m.ParseFrom(buf, 1+digits)
		if err != nil {
			return nil, nil, err
		}
		msg.VariableHeader = m

		// The payload of a SUBSCRIBE packet MUST contain at least one Topic Filter / QoS pair.
		// A SUBSCRIBE packet with no payload is a protocol violation [MQTT-3.8.3-3].
		// payload 检查, 至少四个字节
		if mqttMsgLen-index < 4 {
			return nil, nil, errors.New("非法的 SUBSCRIBE 报文")
		}
		payload := &message.MqttSubscribePayload{}
		if _, err := payload.ParseFrom(buf, index, mqttMsgLen); err != nil {
			return nil, nil, err
		}
		msg.Payload = payload

		return msg, buf[mqttMsgLen:], nil
	case message.UNSUBSCRIBE:
		m := new(message.MqttMessageIdVariableHeader)
		index, err := m.ParseFrom(buf, 1+digits)
		if err != nil {
			return nil, nil, err
		}
		msg.VariableHeader = m

		// The Payload of an UNSUBSCRIBE packet MUST contain at least one Topic Filter.
		// An UNSUBSCRIBE packet with no payload is a protocol violation [MQTT-3.10.3-2].
		// payload 检查，至少三个字节
		if mqttMsgLen-index < 3 {
			return nil, nil, errors.New("非法的 UNSUBSCRIBE 报文")
		}

		topic, payload := "", make([]string, 0, 1)
		for {
			topic, index = utils.DecodeMqttString(buf, index)
			payload = append(payload, topic)
			if index == mqttMsgLen {
				break
			} else if index > mqttMsgLen {
				return nil, nil, errors.New("非法的 UNSUBSCRIBE 报文")
			}
		}
		msg.Payload = payload

		return msg, buf[mqttMsgLen:], nil
	case message.PINGREQ:
		fallthrough
	case message.DISCONNECT:
		return msg, buf[mqttMsgLen:], nil
	default:
		return nil, nil, errors.New(fmt.Sprintf("非法的MQTT报文类型: %d", fixedHeader.MessageType))
	}
}

// 解码载荷
func decodeConnPayload(variableHeader interface{}, buf []byte) *message.MqttConnPayload {
	payload := new(message.MqttConnPayload)

	connVariableHeader := variableHeader.(*message.MqttConnVariableHeader)
	index := 0

	// clientId
	payload.ClientId, index = utils.DecodeMqttString(buf, index)

	// 遗嘱消息
	if connVariableHeader.WillFlag {
		payload.WillTopic, index = utils.DecodeMqttString(buf, index)
		payload.WillMessage, index = utils.DecodeMqttBytes(buf, index)
	}

	// 用户名/密码
	if connVariableHeader.UsernameFlag {
		payload.Username, index = utils.DecodeMqttString(buf, index)
	}
	if connVariableHeader.PasswordFlag {
		payload.Password, index = utils.DecodeMqttString(buf, index)
	}

	return payload

}
