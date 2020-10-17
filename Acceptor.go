package oud

import (
	"log"
	"syscall"
)

// NewConnectionCallback 新连接事件回调
type NewConnectionCallback func(sockfd int, peerAddr syscall.Sockaddr)

// Acceptor 接收器
type Acceptor struct {
	loop                  *EventLoop
	acceptSocket          int
	acceptChannel         Channel // 注意这里不是指针 如果是呢？
	newConnectionCallback func(sockfd int, peerAddr syscall.Sockaddr)
	listenning            bool
	listenAddr            syscall.Sockaddr
}

// NewAcceptor 创建一个Acceptor
// 需要设置 loop 和 要监听的地址
func NewAcceptor(loop *EventLoop, listenAddr syscall.Sockaddr) *Acceptor {
	acp := &Acceptor{
		loop:         loop,
		acceptSocket: CreateNonblockingOrDie(), // 非阻塞套接字 socket
		listenning:   false,
		listenAddr:   listenAddr,
	}
	acp.acceptChannel = *NewChannel(loop, acp.acceptSocket)
	// acp.acceptorSocket.SetReuseAddr(true)
	BindOrDie(acp.acceptSocket, listenAddr) // bind
	acp.acceptChannel.SetReadCallback(acp.HandleRead())
	return acp
}

// HandleRead 注册到Channel的EventCallback 中供调用
// 这里使用了闭包匹配了类型又可以访问当前的数据集
func (a *Acceptor) HandleRead() func() {
	// 考虑到文件描述符耗尽的情况如何处理
	// 非阻塞的poll一下 如果可写表示正常

	// 当前的连接接收策略：
	// 1、一次接收一个连接 适用于长连接
	// 其它的连接接收策略： 适用于短链接
	// 2、一次接收多个连接 直到没有新的连接
	// 3、一次接收固定个连接 一般是10个
	return func() {
		// 非阻塞IO
		nfd, sa, err := syscall.Accept4(a.acceptSocket, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC)
		if err != nil {
			log.Fatal("syscall.Accept Err:", err)
		}

		if a.newConnectionCallback != nil {
			a.newConnectionCallback(nfd, sa)
		} else {
			syscall.Close(nfd)
		}
	}
}

// Listen 将accrpto 的channel写入loop中
// 开始监听端口 接收连接
func (a *Acceptor) Listen() {
	a.listenning = true
	err := syscall.Listen(a.acceptSocket, syscall.SOMAXCONN)
	if err != nil {
		log.Fatal(err)
	}
	a.acceptChannel.EnableReading()
}

// SetNewConnectionCallback 如其名
func (a *Acceptor) SetNewConnectionCallback(cb NewConnectionCallback) {
	a.newConnectionCallback = cb
	//a.acceptChannel.SetReadCallback(a.HandleRead())
	return
}

// GetListenAddr 返回监听地址
func (a *Acceptor) GetListenAddr() syscall.Sockaddr {
	return a.listenAddr
}
