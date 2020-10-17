package main

import (
	"fmt"
	"github.com/haunanz/oud"
	"syscall"
)

func main() {
	fmt.Println(syscall.Getpid(), "start")

	listenAddr := syscall.SockaddrInet4{Port: 8080}
	loop := oud.NewEventLoop()

	server := oud.NewTCPServer(loop, &listenAddr,"test8-server")
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)
	server.Start()

	loop.Loop()
}

func onConnection(conn *oud.TCPConnection) {
	if conn.Connected() {
		fmt.Println(conn.Name(), conn.PeerAddr(), "is connected")
	} else {
		fmt.Println(conn.Name(), conn.PeerAddr(), "is down")
	}
}

func onMessage(conn *oud.TCPConnection, buf *oud.Buffer) {
	fmt.Printf("on messag recive %d/%d bytes | from %s\n", buf.ReadableSize(), buf.BufSize(), conn.Name())
	data := buf.ReadSlice()
	fmt.Print(string(data))
	conn.Send(data)
	buf.Retrieve(len(data))

}
