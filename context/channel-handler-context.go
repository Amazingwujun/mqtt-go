package context

import (
	"mqtt-go/channel"
	"mqtt-go/handler"
	"time"
)

type ChannelHandlerContext struct {

	/** channel  */
	channel *channel.Channel

	inboundHandlers []handler.InboundHandler
}

func (ChannelHandlerContext) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (ChannelHandlerContext) Done() <-chan struct{} {
	panic("implement me")
}

func (ChannelHandlerContext) Err() error {
	panic("implement me")
}

func (ChannelHandlerContext) Value(key interface{}) interface{} {
	panic("implement me")
}
