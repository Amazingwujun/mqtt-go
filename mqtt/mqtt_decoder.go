package mqtt

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	CONNECT byte = iota + 1
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

// 解码状态
const (
	READ_FIXED_HEADER byte = iota
	READ_VARIABLE_HEADER
	READ_PAYLOAD
)

type MqttDecoder struct {

	// 状态值
	State byte

	FixedHeader *MqttFixedHeader

	VariableHeader interface{}
}

func Decode(buf []byte) (*MqttMessage, []byte, error) {
	// 最少最少也有两个字节的数据
	bufLen := len(buf)
	if bufLen < 2 {
		return nil, buf, nil
	}

	// 检查 buf 是否最少含有一个完整的消息
	remainingLen, digits, err := ParseRemainLength(buf[1:])
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
	fixedHeader := &MqttFixedHeader{
		MessageType:  buf[0] >> 4,
		Qos:          (buf[0] & 0b0110) >> 1,
		Dup:          ((buf[0] & 0b1000) >> 1) == 1,
		Retain:       (buf[0] & 0b1) == 1,
		RemainLength: remainingLen,
	}

	// 2 variable header
	msg := &MqttMessage{
		FixedHeader: fixedHeader,
	}
	switch fixedHeader.MessageType {
	case CONNECT:
		if connVariableHeader, err := ReadFrom(buf[1+digits:]); err != nil {
			return nil, nil, err
		} else {
			msg.VariableHeader = connVariableHeader
			// conn 类型的报文固定头为 10 个字节
			payload, err := decodePayload(fixedHeader.MessageType, connVariableHeader, buf[1+digits+10:])
			if err != nil {
				return nil, nil, err
			}
			msg.Payload = payload
			return msg, buf[mqttMsgLen:], nil
		}
	case PUBLISH:
		m := new(MqttPublishVaribleHeader)
		index, err := m.ParseFrom(buf, fixedHeader.Qos, 1+digits)
		if err != nil {
			return nil, nil, err
		}
		msg.VariableHeader = m

		// 获取 payload
		msg.Payload = buf[index:mqttMsgLen]

		return msg, buf[mqttMsgLen:], nil
	case PUBACK:
		fallthrough
	case PUBREC:
		fallthrough
	case PUBREL:
		fallthrough
	case PUBCOMP:
		m := new(MqttMessageIdVariableHeader)
		_, err := m.ParseFrom(buf, 1+digits)
		if err != nil {
			return nil, nil, err
		}
		msg.VariableHeader = m

		return msg, buf[mqttMsgLen:], nil
	case SUBSCRIBE:
		// Bits 3,2,1 and 0 of the fixed header of the SUBSCRIBE Control Packet
		// are reserved and MUST be set to 0,0,1 and 0 respectively. The Server MUST
		// treat any other value as malformed and close the Network Connection
		// [MQTT-3.8.1-1].
		if fixedHeader.Dup || fixedHeader.Qos != 1 || fixedHeader.Retain {
			return nil, nil, errors.New("SUBSCRIBE 报文格式非法")
		}

		// 可变头
		m := new(MqttMessageIdVariableHeader)
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
		payload := &MqttSubscribePayload{}
		if _, err := payload.ParseFrom(buf, index, mqttMsgLen); err != nil {
			return nil, nil, err
		}
		msg.Payload = payload

		return msg, buf[mqttMsgLen:], nil
	case UNSUBSCRIBE:
		m := new(MqttMessageIdVariableHeader)
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
			topic, index = DecodeMqttString(buf, index)
			payload = append(payload, topic)
			if index == mqttMsgLen {
				break
			} else if index > mqttMsgLen {
				return nil, nil, errors.New("非法的 UNSUBSCRIBE 报文")
			}
		}
		msg.Payload = payload

		return msg, buf[mqttMsgLen:], nil
	case PINGREQ:
		fallthrough
	case DISCONNECT:
		return msg, buf[mqttMsgLen:], nil
	default:
		return nil, nil, errors.New(fmt.Sprintf("非法的MQTT报文类型: %d", fixedHeader.MessageType))
	}
}

// 解码 MQTT 中的字符串数据
// 格式见: MQTTV3.1.1 -> 1.5.3 UTF-8 encoded strings
func DecodeMqttString(buf []byte, start int) (string, int) {
	// 先读取长度
	u := int(binary.BigEndian.Uint16(buf[start:]))

	// 返回下一个未读字节的索引
	return string(buf[start+2 : start+2+u]), start + 2 + u
}

// 解码可变字节数组
func DecodeMqttBytes(buf []byte, start int) ([]byte, int) {
	// 先读取长度
	u := int(binary.BigEndian.Uint16(buf[start:]))

	// 返回下一个未读字节的索引
	return buf[start+2 : start+2+u], start + 2 + u
}

// 解码载荷
func decodePayload(messageType byte, variableHeader interface{}, buf []byte) (interface{}, error) {
	switch messageType {
	case CONNECT:
		payload := new(MqttConnPayload)

		connVariableHeader := variableHeader.(*MqttConnVariableHeader)
		index := 0

		// clientId
		payload.ClientId, index = DecodeMqttString(buf, index)

		// 遗嘱消息
		if connVariableHeader.WillFlag {
			payload.WillTopic, index = DecodeMqttString(buf, index)
			payload.WillMessage, index = DecodeMqttBytes(buf, index)
		}

		// 用户名/密码
		if connVariableHeader.UsernameFlag {
			payload.Username, index = DecodeMqttString(buf, index)
		}
		if connVariableHeader.PasswordFlag {
			payload.Password, index = DecodeMqttString(buf, index)
		}

		return payload, nil
	case PUBLISH:
	case SUBSCRIBE:
	case UNSUBSCRIBE:
	}

	return nil, nil
}

// 算法参考 MQTTV3.1.1 协议
//       multiplier = 1
//       value = 0
//       do
//            encodedByte = 'next byte from stream'
//            value += (encodedByte AND 127) * multiplier
//            multiplier *= 128
//            if (multiplier > 128*128*128)
//               throw Error(Malformed Remaining Length)
//       while ((encodedByte AND 128) != 0)
func ParseRemainLength(buf []byte) (int, int, error) {
	bufLen := len(buf)

	multiplier, loops, value := 1, 0, 0
	var encodedByte int
	for {
		if bufLen-1 < loops {
			return 0, 0, nil
		}

		encodedByte = int(buf[loops])
		value += (encodedByte & 127) * multiplier
		multiplier *= 128
		loops++
		if encodedByte&128 != 0 && loops < 4 {
			continue
		} else {
			break
		}
	}

	// MQTT protocol limits Remaining Length to 4 bytes
	if loops == 4 && encodedByte == 128 {
		return 0, 0, errors.New("remain length 超过了规定的四个字节")
	}

	// 返回循环次数，表明 remain length 字段长度
	return value, loops, nil
}
