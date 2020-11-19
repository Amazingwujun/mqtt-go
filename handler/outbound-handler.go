package handler

type OutboundHandler interface {
	write(buf interface{})

	writeAndFlush(buf interface{})

	// 是否支持处理 buf
	support(buf interface{}) bool
}
