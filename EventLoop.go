package oud

import (
	"os"
	"sync"
)

// EventLoop 事件循环
// 负责循环 调用poll 和执行时间
type EventLoop struct {
	looping        bool
	quit           bool
	pid            int // 进程ID
	poller         *Poller
	activeChannels []*Channel
	connectNum     int // 当前持有的连接或文件描述符数
	loopGor        *LoopGoroutine
	taskQueue      []func()
	mutex          sync.Mutex
}

// NewEventLoop 创建一个事件循环
func NewEventLoop() *EventLoop {
	el := &EventLoop{
		looping:        false,
		quit:           false,
		pid:            os.Getpid(),
		activeChannels: []*Channel{},
	}

	// 每个loop 有自己单独的poller 单独管理描述符事件的监听
	el.poller = NewPoller(el)

	return el
}

// Loop 事件循环开始
func (e *EventLoop) Loop() {
	e.looping = true

	for !e.quit {
		// 清空操作
		// 值得注意的是 不置nil,只是切片截断
		// 会导致底层的数组持有指针 导致内存无法释放
		for i := 0; i < len(e.activeChannels); i++ {
			e.activeChannels = nil
		}
		e.activeChannels = e.activeChannels[:0]

		// 系统调用的封装
		e.poller.Poll(1, &e.activeChannels)

		for i := 0; i < len(e.activeChannels); i++ {
			e.activeChannels[i].handelEvents()
		}

		// todo 定时器 和 定时任务

		// 执行其它线程传过来的任务
		e.doTask()
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

// 选择loopGoroutine 中最小的负载
// 没有例程的话 就说明是单独的事件循环 返回自身
func (e *EventLoop) getNextLoop() *EventLoop {
	if e.loopGor == nil {
		return e
	}
	return e.loopGor.getNextLoop()
}

// 多线程之间 添加任务队列
// 线程安全
func (e *EventLoop) addTask(task func()) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.taskQueue = append(e.taskQueue, task)

	return
}

func (e *EventLoop) doTask() {
	// 将任务队列移出执行
	// 减小临界区域的大小 避免阻塞其它线程
	e.mutex.Lock()
	tmpQueue := e.taskQueue
	e.taskQueue = nil
	e.mutex.Unlock()

	for _, task := range tmpQueue {
		task()
	}

	return
}

// 在指定的loop线程中执行task
func (e *EventLoop) runInLoop(loop *EventLoop, task func()) {
	if e == loop {
		task()
	} else {
		loop.addTask(task)
	}
}
