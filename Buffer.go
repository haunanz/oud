package oud

import (
	"errors"
	"syscall"
)

// Buffer struct
// 仿照muduo的buffer设计
type Buffer struct {
	buf        []byte
	readIndex  int
	writeIndex int
}

// NewBuffer func
// 返回一个 Buffer对象
// 预留8字节的缓冲区
func NewBuffer() *Buffer {
	return &Buffer{
		buf:        make([]byte, bufferInitialSize),
		writeIndex: 8,
		readIndex:  8,
	}
}

// Append 向缓冲区尾部中添加数据
// Append
func (b *Buffer) Append(data []byte) (int, error) {
	writeAble := len(b.buf) - b.writeIndex
	readAble := b.writeIndex - b.readIndex
	// 缓冲区足够大到可写
	if len(data) < writeAble {
		copy(b.buf[b.writeIndex:b.writeIndex+len(data)], data)
		b.writeIndex += len(data)
		return len(data), nil
	}

	// 缓冲区不足
	// 1、如果再加上前端缓冲区足够 腾挪
	if len(data) < b.readIndex-8+writeAble {
		for i := 8; i < readAble+8; i++ {
			b.buf[i] = b.buf[i+b.readIndex]
		}
		copy(b.buf[8+readAble:], data)
		b.writeIndex = 8 + readAble + len(data)
		b.readIndex = 8
		return len(data), nil
	}

	// 2、否则 扩容
	size := 8 + readAble + len(data)
	if size < 1024*10 {
		size *= 2
	} else {
		size = int(float64(size) * 1.2)
	}
	temp := make([]byte, size)
	copy(temp[8:8+readAble], b.buf[b.readIndex:b.writeIndex])
	copy(temp[8+readAble:8+readAble+len(data)], data)
	b.buf = temp
	b.readIndex = 8
	b.writeIndex = 8 + readAble + len(data)
	return len(data), nil
}



// AddIndex 配合WriteSlice使用 
func (b *Buffer) AddIndex(n int) {
	if b.writeIndex+n > len(b.buf) {
		b.writeIndex = len(b.buf) - 1
		return
	}
	b.writeIndex += n
}

// ReadFD 传入fd 读取数据
// TODO 改进读取 使其不不用再有额外的复制
// 具体做法是传入可写的区域 如果慢了就扩容
func (b *Buffer) ReadFD(fd int) (int, error) {
	var extrabuf [65536]byte
	n, err := syscall.Read(fd, extrabuf[:])
	if err != nil {
		return n, err
	}
	b.Append(extrabuf[:n])
	return n, nil
}



// BufSize 获得整个缓冲区的大小 非必须
func (b *Buffer) BufSize() int {
	return len(b.buf)
}

// ReadableSize 可读的缓冲区大小
func (b *Buffer) ReadableSize() int {
	return b.writeIndex - b.readIndex
}

// writeableSize 可写的缓冲区大小
func (b *Buffer) writeableSize() int {
	return len(b.buf) - b.writeIndex
}

// preWriteableSize 前端可写的缓冲区大小
func (b *Buffer) preWriteableSize() int {
	return b.readIndex
}

// Prepend 往前面添加数据
// 返回 写入字节数 如果超过头部剩余字节会返回错误
// TODO 头部不足就增长或挪移
func (b *Buffer) Prepend(data []byte) (n int, err error) {
	if len(data) > b.preWriteableSize() {
		return 0, errors.New("preWriteableSize too small")
	}
	copy(b.buf[b.readIndex-len(data):b.readIndex], data)
	b.readIndex -= len(data)
	return len(data), nil
}

// Retrieve 从缓冲区消去多少字节的数据
// 返回消去了多少数据
func (b *Buffer) Retrieve(n int) int {
	rLen := b.ReadableSize()

	// 缓冲区数据大于要消去的数据
	if n < rLen {
		b.readIndex += n
		return n
	}
	// 缓冲区数据小于等于要消去的数据
	retrieveLen := b.ReadableSize()
	b.readIndex = 8
	b.writeIndex = 8
	return retrieveLen

}

// Peek 从缓冲区顶部读多少字节的数据
// 如果可读的缓冲区不够长 那么后面的数据是零值
func (b *Buffer) Peek(n int) []byte {
	result := make([]byte, n)
	// 不会越界
	if b.readIndex+n < b.writeIndex {
		copy(result[:], b.buf[b.readIndex:b.readIndex+n])
		return result
	}
	// 会越界
	copy(result[:], b.buf[b.readIndex:b.writeIndex])
	return result
}

// TODO
//func (b *Buffer) PeekInt8
//func (b *Buffer) PeekInt16
//func (b *Buffer) PeekInt32
//func (b *Buffer) PeekIn64

// Read 从缓冲区读取数据 数据拷贝到传入的切片参数上
// 传出的数据为 min(len(d),b.ReadableSize())
// 读取的数据不会被消去
// 消去数据请用 Retrieve(int)int
func (b *Buffer) Read(d []byte) (int, error) {
	n := len(d)
	ra := b.ReadableSize()
	if n < ra { // 可读区域足够
		copy(d, b.buf[b.readIndex:b.readIndex+n])
		return n, nil
	}
	// 可读区域不足
	copy(d, b.buf[b.readIndex:b.writeIndex])
	return (b.writeIndex - b.readIndex), nil
}

// ReadSlice 返回和缓冲区公用底层数组的切片
// !注意对返回的切片不要做改变大小的操作 以免出现错误
// 如果我想要和缓冲区共享数据 直接把缓冲区的切片返回
// 这时如果是单线程 用了直接释放 不会又安全的风险
// 并且减少了一次内存的分配 和 复制 提高了效率
func (b *Buffer) ReadSlice() []byte {
	// if b.ReadableSize() < n {
	// 	return b.buf[b.readIndex:b.writeIndex]
	// }
	return b.buf[b.readIndex : b.writeIndex]
}

// WriteSlice 返回可写的切片
func (b *Buffer) WriteSlice() []byte {
	return b.buf[b.writeIndex:]
}