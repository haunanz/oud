// TCPConection
// 一个实际的连接对象封装

package oud

import (
	"errors"
	"log"
	"syscall"
)

// TCPConnection 一个连接的必要数据
type TCPConnection struct {
	loop                  *EventLoop
	name                  string
	state                 int // todo chang type name
	sockfd                int
	channel               *Channel
	localAddr             syscall.Sockaddr
	peerAddr              syscall.Sockaddr
	connectionCallback    func(*TCPConnection)
	messageCallback       func(*TCPConnection, *Buffer)
	closeCallback         func(*TCPConnection)
	writeCompleteCallback func(*TCPConnection)
	inputBuf              *Buffer
	outputBuf             *Buffer
}

// NewTCPConnection 创建一个新的Conn
func NewTCPConnection(loop *EventLoop, name string, sockfd int,
	localAddr syscall.Sockaddr, peerAddr syscall.Sockaddr) *TCPConnection {
	c := &TCPConnection{
		loop: loop,
		name: name,
		// state              :
		sockfd:    sockfd,
		channel:   NewChannel(loop, sockfd),
		localAddr: localAddr,
		peerAddr: peerAddr,
		inputBuf:  NewBuffer(),
		outputBuf: NewBuffer(),
	}

	c.channel.SetReadCallback(func() {
		c.handleRead()
	})
	c.channel.SetWriteCallback(func() {
		c.handelWrite()
	})
	c.channel.SetCloseCallback(func() {
		c.handelWrite()
	})
	c.channel.SetErrorCallback(func() {
		_, err := syscall.GetsockoptInt(c.sockfd, syscall.SOL_SOCKET, syscall.SO_ERROR)
		if err != nil {
			c.handleError(err)
		} else {
			newErr := errors.New("syscall.GetsockoptInt val")
			c.handleError(newErr)
		}
	})

	return c
}

// Connected 是否已连接
func (c *TCPConnection) Connected() bool {
	// todo
	return true
}

// Name 返回connection name
func (c *TCPConnection) Name() string {
	return c.name
}

// PeerAddr 返回peerAddr
func (c *TCPConnection) PeerAddr() syscall.Sockaddr {
	return c.peerAddr
}

// FD 返回文件描述符
func (c *TCPConnection) FD() int {
	return c.sockfd
}

// SetState 设置状态
func (c *TCPConnection) SetState(s int) {
	c.state = s
	return
}

// SetConnectionCallback 设置连接回调函数
// 在连接时会调用
func (c *TCPConnection) SetConnectionCallback(cb func(*TCPConnection)) {
	c.connectionCallback = cb
	return
}

// SetMessageCallback 设置消息处理的回调函数
func (c *TCPConnection) SetMessageCallback(cb func(*TCPConnection, *Buffer)) {
	c.messageCallback = cb
	return
}

// SetCloseCallback 设置关闭操作的回调
func (c *TCPConnection) SetCloseCallback(cb CloseCallback) {
	c.closeCallback = cb
	return
}

// SetWriteCompleteCallback 写缓冲区清空回调
func (c *TCPConnection) SetWriteCompleteCallback(cb func(*TCPConnection)) {
	c.writeCompleteCallback = cb
	return
}

// ConnectEstablished 建立连接
// 一些处理过程 和 调用建立连接的函数
// called when TcpServer accepts a new connection
func (c *TCPConnection) ConnectEstablished() {
	c.SetState(tcpConnConnected)
	c.channel.EnableReading()
	if c.connectionCallback != nil {
		c.connectionCallback(c)
	}
}

// ConnectDestroyed 断开连接 释放资源
// called when TcpServer has removed me from its map
func (c *TCPConnection) ConnectDestroyed() {

	c.SetState(tcpConnDisConnecting)
	c.channel.DisableAll()

	// 如果是nil直接panic
	// if c.connectionCallback != nil {
	c.connectionCallback(c)
	c.loop.removeChannel(c.channel)
}

func (c *TCPConnection) handleRead() {
	n, err := c.inputBuf.ReadFD(c.sockfd)
	if err != nil {
		c.handleError(err)
	}

	if n > 0 {
		if c.messageCallback != nil {
			c.messageCallback(c, c.inputBuf)
		}
	} else if n == 0 { // 读0 对方已关闭连接
		c.handleClose() // 我方选择关闭连接
	}

}

func (c *TCPConnection) handelWrite() {
	if c.channel.IsWriting() {
		n, err := syscall.Write(c.sockfd, c.outputBuf.ReadSlice())
		if err != nil { // 处理错误
			log.Println("TCPConnection handlWrite write Err:", err)

		} else { // 没有错误
			c.outputBuf.Retrieve(n)
			if c.outputBuf.ReadableSize() == 0 { // 写完
				c.channel.DisableWriting()

				if c.writeCompleteCallback != nil { // 写缓冲区清空回调
					c.writeCompleteCallback(c)
				}

				if c.state == tcpConnDisConnecting {
					c.Shutdown()
				}
			} else { // 没写完
				log.Println("i am going to write more data")
			}
		}
	} else {
		log.Println("Connection is down, no more writing")
	}
	return
}

func (c *TCPConnection) handleClose() {
	log.Printf("TCPConnection [%s] state: %v\n", c.name, c.state)
	c.channel.DisableAll()
	if c.closeCallback != nil {
		c.closeCallback(c)
	}
}

// muduo 源码
// int sockets::getSocketError(int sockfd)
// {
//   int optval;
//   socklen_t optlen = static_cast<socklen_t>(sizeof optval);
//   if (::getsockopt(sockfd, SOL_SOCKET, SO_ERROR, &optval, &optlen) < 0)
//   {
//     return errno;
//   }
//   else
//   {
//     return optval;
//   }
// }

func (c *TCPConnection) handleError(err error) {
	if err != nil {
		// todo add more conn information
		log.Fatalln("TCPConnection handlerr Err:", err)
	}
}

// Send 发送数据
func (c *TCPConnection) Send(p []byte) {
	nwrote := 0
	if c.state != tcpConnConnected {
		log.Println("disconnected, give up writing")
		return
	}

	if !c.channel.IsWriting() &&
		c.outputBuf.ReadableSize() == 0 {
		var err error
		nwrote, err = syscall.Write(c.sockfd, p) // 直接写
		if err != nil {                          // 出现错误
			log.Println("TCPConnection Send  Write Err:", err)
			nwrote = 0
		}
	}

	if nwrote < len(p) { // 没有写完
		c.outputBuf.Append(p[nwrote:]) // 写入缓冲区

		// 高水位回调
		if c.outputBuf.ReadableSize() >= highWaterMark &&
			c.writeCompleteCallback != nil {
			c.writeCompleteCallback(c)
		}

		if !c.channel.IsWriting() {
			c.channel.EnableWriting() // 注册写
		}

	} else if c.writeCompleteCallback != nil { // 写完 writeComplet 回调
		c.writeCompleteCallback(c)
		c.channel.DisableWriting() // 取消写期望
	}

	return
}

// Shutdown 关闭连接
func (c *TCPConnection) Shutdown() {
	if c.state == tcpConnConnected {
		c.SetState(tcpConnDisConnecting)
	}
}
