package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"mqtt-go/src/utils"
	"time"
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

type MqttMessage struct {

	// 固定头
	FixedHeader *MqttFixedHeader

	// 可变头
	VariableHeader interface{}

	// 载荷
	Payload interface{}
}

func (this *MqttMessage) String() string {
	return fmt.Sprintf("fixedHeader: %v variableHeader: %v payload: %v", *this.FixedHeader, this.VariableHeader, this.Payload)
}

func BuildConnAck(sessionPresent bool, code byte) *MqttMessage {
	msg := &MqttMessage{
		FixedHeader: &MqttFixedHeader{
			MessageType:  CONNACK,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 0,
		},
		VariableHeader: &MqttConnAckVariableHeader{
			SessionPresent: sessionPresent,
			Code:           code,
		},
		Payload: nil,
	}

	return msg
}

// 构建 publish 报文
func BuildPublish(isDub bool, isRetain bool, qos byte, topic string, messageId uint16, payload []byte) *MqttMessage {
	return &MqttMessage{
		FixedHeader: &MqttFixedHeader{
			MessageType:  PUBLISH,
			Qos:          qos,
			Dup:          isDub,
			Retain:       isRetain,
			RemainLength: 0,
		},
		VariableHeader: &MqttPublishVaribleHeader{
			TopicName: topic,
			MessageId: messageId,
		},
		Payload: payload,
	}
}

func BuildPubAck(messageId uint16) *MqttMessage {
	return buildMsgWithMessageId(messageId, PUBACK)
}

func BuildPubRec(messageId uint16) *MqttMessage {
	return buildMsgWithMessageId(messageId, PUBREC)
}

func BuildPubRel(messageId uint16) *MqttMessage {
	return buildMsgWithMessageId(messageId, PUBREL)
}

func BuildPubComp(messageId uint16) *MqttMessage {
	return buildMsgWithMessageId(messageId, PUBCOMP)
}

func BuildSubAck(messageId uint16, resp []byte) *MqttMessage {
	msg := &MqttMessage{
		FixedHeader: &MqttFixedHeader{
			MessageType:  SUBACK,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 0,
		},
		VariableHeader: &MqttMessageIdVariableHeader{MessageId: messageId},
		Payload:        resp,
	}

	return msg
}

func BuildUnsubAck(messageId uint16) *MqttMessage {
	return buildMsgWithMessageId(messageId, UNSUBACK)
}

// 构建无 payload，variableHeader 仅含 messageId 的响应报文
// 支持 PUBACK, PUBREC, PUBREL, PUBCOMP, SUBACK, UNSUBACK
func buildMsgWithMessageId(messageId uint16, messageType byte) *MqttMessage {
	msg := &MqttMessage{
		VariableHeader: &MqttMessageIdVariableHeader{MessageId: messageId},
		Payload:        nil,
	}

	switch messageType {
	case PUBACK:
		msg.FixedHeader = &MqttFixedHeader{
			MessageType:  PUBACK,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 2,
		}
	case PUBREC:
		msg.FixedHeader = &MqttFixedHeader{
			MessageType:  PUBREC,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 2,
		}
	case PUBREL:
		msg.FixedHeader = &MqttFixedHeader{
			MessageType:  PUBREL,
			Qos:          1,
			Dup:          false,
			Retain:       false,
			RemainLength: 2,
		}
	case PUBCOMP:
		msg.FixedHeader = &MqttFixedHeader{
			MessageType:  PUBCOMP,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 2,
		}
	case UNSUBACK:
		msg.FixedHeader = &MqttFixedHeader{
			MessageType:  UNSUBACK,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 2,
		}
	}

	return msg
}

func BuildPingAck() *MqttMessage {
	msg := &MqttMessage{
		FixedHeader: &MqttFixedHeader{
			MessageType:  PINGRESP,
			Qos:          0,
			Dup:          false,
			Retain:       false,
			RemainLength: 0,
		},
	}

	return msg
}

/*             fixed header                  */

// src 消息固定头
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

type MqttConnAckVariableHeader struct {
	SessionPresent bool

	Code byte
}

func (this *MqttConnAckVariableHeader) ToBytes() []byte {
	buf := make([]byte, 2)
	if this.SessionPresent {
		buf[0] = 1
	}
	buf[1] = this.Code

	return buf
}

func ReadFrom(buf []byte) (result *MqttConnVariableHeader, _ error) {

	// 校验协议名称
	if name, _ := utils.DecodeMqttString(buf, 0); name != "MQTT" {
		return nil, errors.New(fmt.Sprintf("非法的协议名:%s\n", name))
	}

	// src v3.1.1 版本值为 4
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

// 可变头包含主题
type MqttPublishVaribleHeader struct {
	TopicName string
	MessageId uint16
}

func (this *MqttPublishVaribleHeader) ParseFrom(buf []byte, qos byte, start int) (int, error) {
	this.TopicName, start = utils.DecodeMqttString(buf, start)
	if qos == 0 {
		return start, nil
	}
	this.MessageId = binary.BigEndian.Uint16(buf[start:])
	if this.MessageId == 0 {
		return 0, errors.New("非法的 packageId:0")
	}
	return start + 2, nil
}

// 可变头仅包含 MessageId
type MqttMessageIdVariableHeader struct {
	MessageId uint16
}

func (this *MqttMessageIdVariableHeader) ParseFrom(buf []byte, start int) (int, error) {
	msgId := binary.BigEndian.Uint16(buf[start:])
	if msgId == 0 {
		return 0, errors.New("非法的 packageId:0")
	}

	this.MessageId = msgId
	return start + 2, nil
}

/*             payload                  */
type MqttConnPayload struct {
	ClientId string

	Username string

	Password string

	WillTopic string

	WillMessage []byte
}

type MqttSubscribePayload struct {
	Topics []*Topic
}

func (this *MqttSubscribePayload) ParseFrom(buf []byte, start int, messageLen int) (int, error) {
	topics := make([]*Topic, 0, 1)
	topic, index := "", start
	for {
		topic, index = utils.DecodeMqttString(buf, index)

		// The Server MUST treat a SUBSCRIBE packet as malformed and close the Network Connection
		// if any of Reserved bits in the payload are non-zero, or QoS is not 0,1 or 2 [MQTT-3-8.3-4].
		if buf[index]&0b11111100 != 0 || buf[index]&0b11 == 3 {
			return 0, errors.New("非法的 SUBSCRIBE 报文")
		}
		topics = append(topics, &Topic{
			Name: topic,
			Qos:  buf[index] & 0b11,
		})
		index++

		// 需要防止非法数据导致的无限循环
		if messageLen == index {
			this.Topics = topics
			return index, nil
		} else if messageLen < index {
			return 0, errors.New("非法的 SUBSCRIBE 报文")
		}
	}
}

type Topic struct {
	Name string

	Qos byte
}
