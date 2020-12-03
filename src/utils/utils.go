// 工具方法

package utils

import (
	"encoding/binary"
	errors "errors"
)

// 解码报文长度字节, 算法参考 MQTTV3.1.1 协议
//       multiplier = 1
//       value = 0
//       do
//            encodedByte = 'next byte from stream'
//            value += (encodedByte AND 127) * multiplier
//            multiplier *= 128
//            if (multiplier > 128*128*128)
//               throw Error(Malformed Remaining Length)
//       while ((encodedByte AND 128) != 0)
func DecodeRemainLength(buf []byte) (int, int, error) {
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

// 编码报文长度数据
//    do
//      encodedByte = X MOD 128
//      X = X DIV 128
//      // if there are more data to encode, set the top bit of this byte
//      if ( X > 0 )
//        encodedByte = encodedByte OR 128
//      endif
//         'output' encodedByte
//    while ( X > 0 )
func EncodeRemainLength(len int) []byte {
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

// 解码 MQTT 中的字符串数据
// 格式见: MQTTV3.1.1 -> 1.5.3 UTF-8 encoded strings
func DecodeMqttString(buf []byte, fromIndex int) (string, int) {
	// 先读取长度
	u := int(binary.BigEndian.Uint16(buf[fromIndex:]))

	// 返回下一个未读字节的索引
	return string(buf[fromIndex+2 : fromIndex+2+u]), fromIndex + 2 + u
}

// 解码可变字节数组
func DecodeMqttBytes(buf []byte, fromIndex int) ([]byte, int) {
	// 先读取长度
	u := int(binary.BigEndian.Uint16(buf[fromIndex:]))

	// 返回下一个未读字节的索引
	return buf[fromIndex+2 : fromIndex+2+u], fromIndex + 2 + u
}

// 用于判定客户订阅的主题是否匹配发布主题
//	pub: 发布主题
// 	sub: 定于主题 - 主题过滤器
// 不支持通配符
func Match(pub string, sub string) bool {
	if pub == sub {
		return true
	}

	return false
}
