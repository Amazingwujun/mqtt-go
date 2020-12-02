package message

type PubMsg struct {
	// topic name
	Topic string

	// qos 级别
	Qos byte

	// 会话是否存在
	SessionPresent bool

	// 消息体
	Payload []byte
}
