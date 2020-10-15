package oud

import (
	"log"
	"strconv"
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
	
	started               bool
	nextConnID            int
	connectionMap         map[string]*TCPConnection
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
	server.acceptor.SetNewConnectionCallback(server.NewConnection())
	return server
}

// NewConnection 返回一个闭包
//
// type NewConnectionCallback func(sockfd int, peerAddr syscall.Sockaddr)
func (s *TCPServer) NewConnection() NewConnectionCallback {
	return func(sockfd int, peerAddr syscall.Sockaddr) {
		connName := s.name + strconv.Itoa(s.nextConnID)
		s.nextConnID++
		log.Printf("TCPServer.NewConnection [%s]- new conn [%s] from %v\n", s.name, connName, peerAddr)
		localAddr := s.acceptor.GetListenAddr()

		conn := NewTCPConnection(s.loop,
			connName,
			sockfd,
			localAddr,
			peerAddr,
		)

		s.connectionMap[connName] = conn

		conn.SetConnectionCallback(s.connectionCallback)
		conn.SetMessageCallback(s.messageCallback)
		conn.SetCloseCallback(func(conn *TCPConnection) {
			s.RemoveConnection(conn)
		})
		conn.SetWriteCompleteCallback(s.writeCompleteCallback)

		conn.ConnectEstablished()

	}
}

// Start ->acceptor.Listen()
func (s *TCPServer) Start() {
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
	delete(s.connectionMap, conn.Name())
	s.loop.removeChannel(conn.channel)
}
