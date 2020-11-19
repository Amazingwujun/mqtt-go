package pipline

import (
	"mqtt-go/channel"
	"mqtt-go/handler"
)

type ChannelPipeline interface {
	addInboundHandlerAtLast(in handler.InboundHandler)

	addOutboundHandlerAtLast(out handler.OutboundHandler)

	fireWrite(msg interface{})

	fireRead(msg interface{})
}

type DefaultChannelPipeline struct {
	channel channel.Channel

	inboundHandlers []handler.InboundHandler

	outboundHandlers []handler.OutboundHandler
}

func (this *DefaultChannelPipeline) fireWrite(msg interface{}) {
	panic("implement me")
}

func (this *DefaultChannelPipeline) fireRead(msg interface{}) {
	panic("implement me")
}

func (this *DefaultChannelPipeline) addInboundHandlerAtLast(in handler.InboundHandler) {
	panic("implement me")
}

func (this *DefaultChannelPipeline) addOutboundHandlerAtLast(out handler.OutboundHandler) {
	panic("implement me")
}
