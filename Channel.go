// Channel 负责事件的分发处理
// 包含期望事件

package oud

import (
	"syscall"
)

// Channel 负责消息分发
type Channel struct {
	loop          *EventLoop
	sockfd        int
	events        uint32 // 期望事件
	revents       uint32 // 当前活跃事件
	index         int    // Poller 中使用
	readCallback  EventCallback
	writeCallback EventCallback
	errCallback   EventCallback
	closeCallback EventCallback
}

// NewChannel 创建一个Channel
func NewChannel(loop *EventLoop, fd int) *Channel {
	return &Channel{
		loop:   loop,
		sockfd: fd,
	}
}

func (c *Channel) handelEvents() {

	// 关闭连接
	if c.revents&syscall.EPOLLHUP != 0 &&
		!(c.revents&syscall.EPOLLIN != 0) {
		if c.closeCallback != nil {
			c.closeCallback()
		}
	}

	// 错误处理
	// syscall.EPOLLNVAL
	if c.revents&(syscall.EPOLLERR) != 0 {
		if c.errCallback != nil {
			c.errCallback()
		}
	}

	// 读事件
	if c.revents&
		(syscall.EPOLLIN|syscall.EPOLLPRI|syscall.EPOLLRDHUP) != 0 {
		if c.readCallback != nil {
			c.readCallback()
		}
	}

	// 写事件
	if c.revents&syscall.EPOLLOUT != 0 {
		if c.writeCallback != nil {
			c.writeCallback()
		}
	}
}

// SetReadCallback 设置读事件触发的回调函数
func (c *Channel) SetReadCallback(cb EventCallback) {
	c.readCallback = cb
	return
}

// SetWriteCallback 设置写事件触发的回调函数
func (c *Channel) SetWriteCallback(cb EventCallback) {
	c.writeCallback = cb
	return
}

// SetErrorCallback 设置错误事件触发的回调函数
func (c *Channel) SetErrorCallback(cb EventCallback) {
	c.errCallback = cb
	return
}

// SetCloseCallback 设置关闭事件触发的回调函数
func (c *Channel) SetCloseCallback(cb EventCallback) {
	c.closeCallback = cb
	return
}

// FD 返回Channel管理的文件描述符
func (c *Channel) FD() int {
	return c.sockfd
}

// Events 获取期望事件
func (c *Channel) Events() uint32 {
	return c.events
}

// SetRevents 设置活动事件
func (c *Channel) SetRevents(revt uint32) {
	c.revents = revt
	return
}

// IsNoneEvent 判断当前无期望事件
func (c *Channel) IsNoneEvent() bool {
	return c.events == channelNoneEvent
}

// EnableReading 期望读事件
func (c *Channel) EnableReading() {
	c.events |= channelReadEvent
	c.update()
	return
}

// DisableReading 关闭读事件期望
func (c *Channel) DisableReading() {
	c.events &= ^channelReadEvent
	c.update()
	return
}

// EnableWriting 期望写入事件
func (c *Channel) EnableWriting() {
	c.events |= channelWriteEvent
	c.update()
	return
}

// DisableWriting 关闭写入事件期望
func (c *Channel) DisableWriting() {
	c.events &= ^channelWriteEvent
	c.update()
	return
}

// DisableAll 不期望任何事件
func (c *Channel) DisableAll() {
	c.events = channelNoneEvent
	c.update()
	return
}

// IsWriting 正在期望读事件
func (c *Channel) IsWriting() bool {
	return (c.events & channelWriteEvent) != 0
}

// IsReading 正在期望读事件
func (c *Channel) IsReading() bool {
	return (c.events & channelReadEvent) != 0
}

// Index Poller使用
func (c *Channel) Index() int {
	return c.index
}

// SetIndex Poller使用
func (c *Channel) SetIndex(idx int) {
	c.index = idx
	return
}

// OwnerLoop 返回OwnerLoop
func (c *Channel) OwnerLoop() *EventLoop {
	return c.loop
}

// channel->loop->poller->poll
func (c *Channel) update() {
	c.loop.updateChannel(c)
	return
}
