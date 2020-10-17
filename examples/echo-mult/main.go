package main

import (
	"fmt"
	"syscall"

	"github.com/haunanz/oud"
)

func main() {
	loops := oud.NewLoops(10)

	listenAddr := syscall.SockaddrInet4{Port: 8080}
	server := oud.NewTCPServer(loops.BaseLoop(), &listenAddr, "echoServer-mult")
	server.SetConnectionCallback(onConnection)
	server.SetMessageCallback(onMessage)
	
	server.Start()
	loops.Loop()

}

func onConnection(conn *oud.TCPConnection) {
	if conn.Connected() {
		fmt.Println(conn.Name(), conn.PeerAddr(), "is connected")
	} else {
		fmt.Println(conn.Name(), conn.PeerAddr(), "is down")
	}
}

func onMessage(conn *oud.TCPConnection, buf *oud.Buffer) {
	fmt.Printf("%d/%d from %s\n", buf.ReadableSize(), buf.BufSize(), conn.Name())
	data := buf.ReadSlice()
	fmt.Println(string(data))
	conn.Send(data)
	buf.Retrieve(len(data))

}
