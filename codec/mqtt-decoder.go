package codec

import (
	"errors"
	"fmt"
)

const PROTOCOL_NAME = "MQTT"

const (
	CONNECT = iota + 1
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

func Decode(buf []byte, channel chan<- interface{}) ([]byte, error) {
	// 最少最少也有两个字节的数据
	bufLen := len(buf)
	if bufLen < 2 {
		return buf, nil
	}

	// 1 fixed header
	packageType := buf[0] >> 4
	//flags := buf[0] & 0x0f

	remainingLen, digits, err := parseVariableLen(buf)
	if err != nil {
		return nil, err
	}
	// 一个完整的 mqtt msg 报文长度
	mqttMsgLen := remainingLen + 1 + digits

	// 检查剩余报文长度
	if bufLen < mqttMsgLen {
		return buf, nil
	}

	// 2 variable header
	switch packageType {
	case CONNECT:
		// 协议名称处理
		protocolName := string(buf[1+digits+2 : 1+digits+2+4])
		if protocolName != PROTOCOL_NAME {
			return nil, errors.New(fmt.Sprintf("非法的协议名称: %s\n", protocolName))
		}

		// 协议版本，仅支持 v3.1.1 版本
		if ver := buf[1+digits+2+4]; ver != 4 {
			return nil, errors.New(fmt.Sprintf("不支持版本: %d", ver))
		}

		return nil, nil
	case PINGREQ:
		channel <- buf[:mqttMsgLen]
		if bufLen > mqttMsgLen {
			return Decode(buf[remainingLen+1+digits:], channel)
		} else {
			return nil, nil
		}
	default:
		return nil, errors.New("未知类型的报文")
	}
}

// 算法参考 MQTTV3.1.1 协议的
//       multiplier = 1
//       value = 0
//       do
//            encodedByte = 'next byte from stream'
//            value += (encodedByte AND 127) * multiplier
//            multiplier *= 128
//            if (multiplier > 128*128*128)
//               throw Error(Malformed Remaining Length)
//       while ((encodedByte AND 128) != 0)
func parseVariableLen(buf []byte) (int, int, error) {

	multiplier := 1
	loops := 1
	value := 0
	var encodedByte int
	for {
		encodedByte = int(buf[loops])
		value += (encodedByte & 127) * multiplier
		multiplier *= 128
		loops++
		if encodedByte&128 != 0 && loops < 5 {
			continue
		} else {
			break
		}
	}

	loops--
	// 返回循环次数，表明 remain length 字段长度
	return value, loops, nil
}
