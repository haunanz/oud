# oud
根据陈硕的《Linux多线程服务端编程 使用muduo C++网络库》中的muduo模型编写
muduo github地址:https://github.com/chenshuo/muduo

目的是为了简单的了解Reactor模模型在网络开发中的工作方式。

注意！注意！注意！  
只适用于linux系统。

相比于muduo时用go语言来描写，这个go语言版的muduo有更少的代码量。  
有几个muduo的功能还未实现，将在以后的版本实现。  
- 定时器
- 任务队列
- 多go例程的eventloop处理
- 引入viper来管理各种配置

准备写个简单的文档来描述这个网络库的处理逻辑。  
准备写示例程序演示功能。  
准备单元测试和功能测试。 

## 示例
获取
`go get github.com/haunanz/oud`

简单的回声服务器。
```go 
package main

import (
	"fmt"
	"syscall"

	"github.com/haunanz/oud"
)

func main() {
	loop := oud.NewEventLoop()
	addr := syscall.SockaddrInet4{Port: 12345}
	server := oud.NewTCPServer(loop, &addr, "server name")
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)

	server.Start()
	loop.Loop()

}

// 处理连接的建立
func onConnection(conn *oud.TCPConnection) {
	if conn.Connected() {
		fmt.Printf("%s is connected\n", conn.Name())
	} 
}

// 处理消息
func onMessage(conn *oud.TCPConnection, buf *oud.Buffer) {
	data := buf.ReadSlice()
	fmt.Printf("%s recive %d bytes in:%d\n", conn.Name(), len(data), conn.FD())
	conn.Send(data)
	buf.Retrieve(len(data))
}
```