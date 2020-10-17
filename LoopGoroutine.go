// one loop per thread -> one loop per goroutine
// 通过多路复用单线程也可以有很高的性能 许多著名的项目也是单线程
// 但现在通常通过多核来提升运算力 这需要多线程并发的技术
// 并发正是go的拿手好戏 对于计算密集型应用场景 高并发带来了性能上的显著提升

package oud

import "sort"

// LoopGoroutine 多个单线程的 TCPServer 的集合
// 充分的利用 多多核的能力
type LoopGoroutine struct {
	baseLoop *EventLoop   // 只用于建立连接
	loops    []*EventLoop // 处理事件循环
	started  bool
}

// NewLoops 建立并发的事件循环
// 处理连接建立的事件循环为1
// n 指定处理其它事件的eventLoop 个数
func NewLoops(n int) *LoopGoroutine {
	l := &LoopGoroutine{
		baseLoop: NewEventLoop(),
		loops:    make([]*EventLoop, n),
	}
	l.baseLoop.loopGor = l
	for i := 0; i < n; i++ {
		l.loops[i] = NewEventLoop()
		l.loops[i].loopGor = l
	}
	return l
}

// Loop 开始循环
func (l *LoopGoroutine) Loop() {
	l.started = true
	for _, loop := range l.loops {
		go loop.Loop()
	}
	l.baseLoop.Loop()
}

// BaseLoop 返回baseLoop
func (l *LoopGoroutine) BaseLoop() *EventLoop {
	return l.baseLoop
}

// 得到负载最小的loop 并返回
func (l *LoopGoroutine) getNextLoop() *EventLoop {
	n := len(l.loops)
	load := make([][2]int, n)
	for i := 0; i < n; i++ {
		load[i] = [2]int{i, l.loops[i].connectNum}
	}
	sort.Slice(load, func(i, j int) bool {
		if load[i][1] < load[j][1] {
			return true
		}
		return false
	})
	return l.loops[load[0][0]]
}
