package oud

import (
	"syscall"
)

var (
	// Buffer
	bufferInitialSize int = 1024 // 缓冲区初始小
	headFreeSapce     int = 8    // 初始头部空闲空间
	bufferMaxSize     int = 1024 * 10

	// epoll Event的长度
	epollEventsNum int = 100

	// 高水位回调
	// 接收缓冲区达到水位时触发
	highWaterMark int = bufferMaxSize
)

// Channel 的状态
const (
	channelNoneEvent  uint32 = 0
	channelReadEvent  uint32 = syscall.EPOLLIN | syscall.EPOLLPRI
	channelWriteEvent uint32 = syscall.EPOLLOUT
)

// TCPConnection 状态枚举
// kDisconnected, kConnecting, kConnected, kDisconnecting
const (
	tcpConnConnected int = iota
	tcpConnConnecting
	tcpConnDisConnected
	tcpConnDisConnecting
)
