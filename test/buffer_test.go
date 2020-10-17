package test

import (
	"fmt"
	"github.com/haunanz/oud"
)

func a() {
	// BUffer test
	buf := oud.NewBuffer()
	buf.Append([]byte("123"))
	fmt.Println(buf, buf.ReadableSize())
	buf.Prepend([]byte{12, 13})
	fmt.Println(buf, buf.ReadableSize())

	_, err := buf.Prepend([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf, buf.ReadableSize())

	buf.Prepend([]byte{3})
	n := buf.Peek(1)
	buf.Retrieve(1)
	fmt.Println("Peek:", n)
	tmp := make([]byte, n[0])
	buf.Read(tmp)
	fmt.Println(tmp)

	buf.Retrieve(int(n[0]))
	fmt.Println(buf, buf.ReadableSize())
}
