package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var msCount int64 = 0

func main() {

	// addr:port goroutine-num conn-num-per-goroutine
	if len(os.Args) != 4 {
		panic("please input: addr:port goroutine-num conn-num-per-goroutine ")
	}

	addr := os.Args[1]
	gNum, _ := strconv.Atoi(os.Args[2])
	cNum, _ := strconv.Atoi(os.Args[3])

	wg := sync.WaitGroup{}

	for i := 0; i < gNum; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {

			conns := make([]net.Conn, cNum)
			for i := 0; i < cNum; i++ {
				var err error
				conns[i], err = net.Dial("tcp", addr)
				if err != nil {
					log.Fatal(err)
				}
			}

			count := 10000
			for i := 0; i < cNum; i++ {
				str := fmt.Sprintf("%d", time.Now().UnixNano())
				conns[i].Write([]byte(str))

				var buf [128]byte
				n, err := conns[i].Read(buf[:])
				if err != nil {
					log.Fatal(err)
				}
				pre, _ := strconv.Atoi(string(buf[:n]))
				now := time.Now().UnixNano() 

				atomic.AddInt64(&msCount, diffMs(now, int64(pre)))

				count--
				if count == 0 {
					break
				}
			}
			wg.Done()
		}(&wg)
	}

	wg.Wait()
	fmt.Println(msCount)
	fmt.Println(int(msCount) / gNum / cNum / 10000)

}

func diffMs(now, last int64) int64 {
	diff := now - last
	if diff > 0 {
		return diff
	}
	return 0
}
