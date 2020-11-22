package mqtt

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type MqttMessage struct {

	// 固定头
	FixedHeader *MqttFixedHeader

	// 可变头
	VariableHeader interface{}

	// 载荷
	Payload interface{}
}

/*             fixed header                  */

func (this *MqttMessage) Write(channel *Channel) {
	if channel.Closed {
		return
	}

	switch this.FixedHeader.MessageType {
	case CONNACK:

	case PINGREQ:

	}

	//channel.Write()
}

// mqtt 消息固定头
type MqttFixedHeader struct {
	MessageType byte

	Qos byte

	Dup bool

	Retain bool

	RemainLength int
}

// 从完整的报文读取固定头
func (this *MqttFixedHeader) readFrom(buf []byte) ([]byte, error) {
	this.MessageType = buf[0] >> 4
	this.Qos = (buf[0] & 0b0110) >> 1
	this.Retain = (buf[0] & 0b1) == 1
	this.Dup = ((buf[0] & 0b1000) >> 1) == 1

	multiplier, loops, value := 1, 1, 0
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

	// MQTT protocol limits Remaining Length to 4 bytes
	if loops == 4 && encodedByte == 128 {
		return nil, errors.New("remain length 超过了规定的四个字节")
	}
	this.RemainLength = value

	return buf[loops:], nil
}

/*             variable header                  */

type MqttConnVariableHeader struct {
	CleanSession bool

	WillFlag bool

	WillQos byte

	WillRetain bool

	UsernameFlag bool

	PasswordFlag bool

	// 心跳
	KeepAlive time.Duration
}

func ReadFrom(buf []byte) (result *MqttConnVariableHeader, _ error) {

	// 校验协议名称
	if name, _ := DecodeMqttString(buf, 0); name != "MQTT" {
		return nil, errors.New(fmt.Sprintf("非法的协议名:%s\n", name))
	}

	// mqtt v3.1.1 版本值为 4
	if buf[6] != 4 {
		return nil, errors.New(fmt.Sprintf("不支持版本: %d", buf[6]))
	}

	// conn flags
	// 先检查保留字段
	connectFlags := buf[7]
	if connectFlags&0b1 != 0 {
		return nil, errors.New("conn flag 保留字段非法！")
	}
	result = new(MqttConnVariableHeader)
	result.CleanSession = (connectFlags&0b10)>>1 == 1
	result.WillFlag = (connectFlags&0b100)>>2 == 1
	result.UsernameFlag = (connectFlags&0x80)>>7 == 1
	result.PasswordFlag = (connectFlags&0x40)>>6 == 1

	// If the Will Flag is set to 0, then the Will Retain Flag MUST be set to 0 [MQTT-3.1.2-15].
	// If the Will Flag is set to 1:
	// If Will Retain is set to 0, the Server MUST publish the Will Message as a non-retained message [MQTT-3.1.2-16].
	// If Will Retain is set to 1, the Server MUST publish the Will Message as a retained message [MQTT-3.1.2-17].
	if result.WillFlag {
		result.WillQos = (connectFlags & 0b11000) >> 3
		result.WillRetain = (connectFlags&0b10_0000)>>5 == 1
	}

	// keep alive
	k := binary.BigEndian.Uint16(buf[8:10])
	result.KeepAlive = time.Duration(k) * time.Second

	return result, nil
}

/*             payload                  */
type MqttConnPayload struct {
	ClientId string

	Username string

	Password string

	WillTopic string

	WillMessage []byte
}
