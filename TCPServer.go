package oud

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"syscall"
)

// TCPServer 用于管理TCPConection 的连接和销毁
type TCPServer struct {
	loop                  *EventLoop
	name                  string
	acceptor              *Acceptor
	connectionCallback    func(*TCPConnection)
	messageCallback       func(*TCPConnection, *Buffer)
	writeCompleteCallback func(*TCPConnection)

	started       bool
	nextConnID    int
	connectionMap map[string]*TCPConnection
	mutex         sync.Mutex
}

// NewTCPServer 创建一个TCPServer
func NewTCPServer(loop *EventLoop, listenAddr syscall.Sockaddr, name string) *TCPServer {
	server := &TCPServer{
		loop:          loop,
		name:          name,
		acceptor:      NewAcceptor(loop, listenAddr),
		started:       false,
		nextConnID:    1,
		connectionMap: map[string]*TCPConnection{},
	}
	server.acceptor.SetNewConnectionCallback(server.newConnection())
	return server
}

// newConnection 返回一个闭包
// 建立新的连接时 会调用这个函数
// loop->poller->channel->acceptor->tcpserver(this)
// type NewConnectionCallback func(sockfd int, peerAddr syscall.Sockaddr)
func (s *TCPServer) newConnection() NewConnectionCallback {
	return func(sockfd int, peerAddr syscall.Sockaddr) {
		connName := s.name + strconv.Itoa(s.nextConnID)
		s.nextConnID++

		inet4 := peerAddr.(*syscall.SockaddrInet4)
		addrStr := fmt.Sprintf("%v:%v", inet4.Addr, inet4.Port)

		log.Printf("TCPServer.NewConnection [%s]-new conn [%s] from %s\n", s.name, connName, addrStr)
		localAddr := s.acceptor.GetListenAddr()

		// 选择负载最少的Loop传入
		ioLoop := s.loop.getNextLoop()
		conn := NewTCPConnection(
			ioLoop,
			connName,
			sockfd,
			localAddr,
			peerAddr,
		)

		s.mutex.Lock()
		s.connectionMap[connName] = conn
		s.mutex.Unlock()

		conn.SetConnectionCallback(s.connectionCallback)
		conn.SetMessageCallback(s.messageCallback)
		conn.SetCloseCallback(func(conn *TCPConnection) {
			s.RemoveConnection(conn)
		})
		conn.SetWriteCompleteCallback(s.writeCompleteCallback)

		// 指定loop中建立连接
		s.loop.runInLoop(ioLoop, func() {
			conn.ConnectEstablished()
			ioLoop.connectNum++
		})
	}
}

// Start ->acceptor.Listen()
func (s *TCPServer) Start() {
	s.started = true
	s.acceptor.Listen()
	return
}

// SetConnectionCallback 在建立新连接时会被注入conn
// 并使用estblish 调用
func (s *TCPServer) SetConnectionCallback(cb func(*TCPConnection)) {
	s.connectionCallback = cb
	return
}

// SetMessageCallback 在建立新连接时会被注入conn
// 在处理读事件时调用
func (s *TCPServer) SetMessageCallback(cb func(*TCPConnection, *Buffer)) {
	s.messageCallback = cb
	return
}

// SetWriteCompleteCallback 写缓冲区清空回调
func (s *TCPServer) SetWriteCompleteCallback(cb func(*TCPConnection)) {
	s.writeCompleteCallback = cb
	return
}

// RemoveConnection 删除connection
func (s *TCPServer) RemoveConnection(conn *TCPConnection) {
	s.mutex.Lock()
	delete(s.connectionMap, conn.Name())
	s.mutex.Unlock()

	// 保证在loop->poller 中的删除是线程安全的
	s.loop.runInLoop(conn.loop, func() {
		conn.loop.removeChannel(conn.channel)
		conn.loop.connectNum--
	})
}
