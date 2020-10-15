package oud

import (
	"os"
)

// EventLoop 事件循环
// 负责循环 调用poll 和执行时间
type EventLoop struct {
	looping        bool
	quit           bool
	pid            int // 进程ID
	poller         *Poller
	activeChannels []*Channel
}

// NewEventLoop 创建一个事件循环
func NewEventLoop() *EventLoop {
	el := &EventLoop{
		looping:        false,
		quit:           false,
		pid:            os.Getpid(),
		activeChannels: []*Channel{},
	}
	el.poller = NewPoller(el)

	return el
}

// Loop 事件循环开始
func (e *EventLoop) Loop() {
	e.looping = true
	
	for !e.quit {
		e.activeChannels = e.activeChannels[:0] // 清空
		e.poller.Poll(-1, &e.activeChannels)    // 系统调用的封装

		for i := 0; i < len(e.activeChannels); i++ {
			e.activeChannels[i].handelEvents()
		}

	}

	e.looping = false
}

func (e *EventLoop) updateChannel(channel *Channel) {
	e.poller.updateChannel(channel)
	return
}

func (e *EventLoop) removeChannel(channel *Channel) {
	e.poller.removeChannel(channel)
	return
}
