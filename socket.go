// socket的syscall封装
// 暂时没有使用封装 

package oud

import (
	"syscall"
)

type socket int

func (s socket) read(p []byte) (int, error) {
	n, err := syscall.Read(int(s), p)
	return n, err
}

func (s socket) write(p []byte) (int, error) {
	n, err := syscall.Write(int(s), p)
	return n, err
}

func (s socket) close() error {
	err := syscall.Close(int(s))
	return err
}

func (s socket) setNonblocking(b bool) error {
	err := syscall.SetNonblock(int(s), b)
	return err
}
