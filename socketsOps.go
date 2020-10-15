// SocketsPos
// 关于Socket的各种常用操作的封装

package oud

import (
	"log"
	"syscall"
)

// CreateNonblockingOrDie 创建一个非阻塞的 TCP Socket文件描述符
func CreateNonblockingOrDie() int {
	socket, err := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC,
		syscall.IPPROTO_TCP,
	)
	if err != nil {
		log.Fatal(err)
	}
	return socket
}


// BindOrDie bind系统调用和错误检查
func BindOrDie(sockfd int, addr syscall.Sockaddr) {
	err := syscall.Bind(sockfd, addr)
	if err != nil {
		log.Fatal("Bind Err:", err)                                                     
	}
	return
}

// CloseSocket 关闭套接字
func CloseSocket(sockfd int) {
	err :=syscall.Close(sockfd)
	if err!=nil {
		log.Println("close socket Err:",err)
	}else {
		log.Printf("sockt fd %d closed\n",sockfd)
	}

}