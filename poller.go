package oud

import (
	"log"
	"syscall"
)

// EpollFdList  Epoll事件的列表
type EpollFdList []syscall.EpollEvent

// ChannelMap 根据文件描述符查找Channel
type ChannelMap map[int]*Channel

type poller interface {
	Poll(timeoutMs int, activeChannels *[]*Channel)
	UpdateChannel(channel *Channel)
	RemoveChannel(channel *Channel)
}

type poll struct {
	// ! syscall.?
	channelsMap map[int32]*Channel
}

type epoll struct {
	epfd        int                  // epoll 对象
	events      []syscall.EpollEvent // 接受的事件
	channelsMap map[int32]*Channel   // 文件描述符对channel的映射
}

func newEpoll() *epoll {
	epfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		log.Fatal("epoll_create1 Err:", err)
	}
	return &epoll{
		epfd:        epfd,
		events:      make([]syscall.EpollEvent, epollEventsNum),
		channelsMap: map[int32]*Channel{},
	}
}

func (p *epoll) fillActiveChannels(numEvents int, activeChannels *[]*Channel) {
	for i := 0; i < numEvents; i++ {
		// 将发生事件的的channel放入 activeChannels 中
		channel, ok := p.channelsMap[p.events[i].Fd]
		if ok {
			// 更新活动事件 在handlEvnets的时候会检查
			// 以判断触发了什么事件
			channel.SetRevents(p.events[i].Events)
			*activeChannels = append(*activeChannels, channel)
		} else {
			log.Fatalf("FD:%d no exist\n", p.events[i].Fd)
		}
	}
}

func (p *epoll) Poll(timeoutMs int, activeChannels *[]*Channel) {
	numEvents, err := syscall.EpollWait(p.epfd, p.events, timeoutMs)
	if numEvents == len(p.events) { // 满了-扩容
		tmp := make([]syscall.EpollEvent, 2*numEvents)
		copy(tmp, p.events)
		p.events = tmp
		log.Printf("Poller.events 扩容:%d -> %d \n", numEvents, 2*numEvents)
	}

	// 忽略由于接收调试信号而产生的"错误"返回
	if err != nil && err != syscall.EINTR {
		log.Fatal("EpollWait Err:", err)
	}

	if numEvents > 0 {
		// ? log.trace
		p.fillActiveChannels(numEvents, activeChannels)
	}

}

func (p *epoll) UpdateChannel(channel *Channel) {
	// ? log.trace
	if channel.Index() == 0 {
		// 一个新的channel
		//!
		channel.SetIndex(1)

		event := syscall.EpollEvent{
			Fd:     int32(channel.FD()),
			Events: channel.Events(),
		}
		err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_ADD, channel.FD(), &event)
		if err != nil {
			log.Fatal("syscall.EpollCtl ADD Err:", err)
		}

		// 不需要自己维护socket events 列表

		p.channelsMap[int32(channel.FD())] = channel
	} else {
		// 更新一个存在的channel
		// idx :=channel.Index()
		event := syscall.EpollEvent{

			Events: channel.Events(),
		}
		err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_MOD, channel.FD(), &event)
		if err != nil {
			log.Fatal("syscall.EpollCtl MOD Err:", err)
		}

	}
	return
}

// RemoveChannel 删除Channel
func (p *epoll) RemoveChannel(channel *Channel) {
	// ? log.trace fd channel->fd
	// 传入的channel的 fd 在表中存在
	// 得到的指针与传入的相同 相同
	// ! moduo的这里一大堆断言 下次再看
	v, ok := p.channelsMap[int32(channel.FD())]
	if ok && v == channel {
		delete(p.channelsMap, int32(channel.FD()))

		event := syscall.EpollEvent{
			Fd: int32(channel.FD()),
		}
		syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_DEL, channel.FD(), &event)

		CloseSocket(channel.FD()) // 在这里关闭了套接字的连接

	} else {
		log.Fatalln("Epoll want to del nonexistent event")
	}
}

// Poller epoll系统调用
// IO多路复用
// 使用了poller 接口只要实现了 Poll 函数就可以替换
// 这里可以使用 epoll 或 poll（暂时未实现）
type Poller struct {
	ownerLoop *EventLoop // 所属的事件循环
	poll      poller
}

// 必要接口
// new 创建接口
// poll 接口系统调用
// fillActive 接口?

// NewPoller 创建poller
func NewPoller(loop *EventLoop) *Poller {
	return &Poller{
		ownerLoop: loop,
		poll:      newEpoll(),
	}
}

// Poll Poller的封装
func (p *Poller) Poll(timeoutMs int, activeChannels *[]*Channel) {
	p.poll.Poll(timeoutMs, activeChannels)
	return
}

// 调用poller接口的updateChannel函数
// 这里其实可以使用匿名字段就免去了封装
func (p *Poller) updateChannel(channel *Channel) {
	p.poll.UpdateChannel(channel)
	return
}

// 对poll接口的封装同上
func (p *Poller) removeChannel(channel *Channel) {
	p.poll.RemoveChannel(channel)
	return
}
