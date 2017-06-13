// Provides leaky buffer, based on the example in Effective Go.
package shadowsocks

import "bytes"

type LeakyBuf struct {
	bufSize     int // size of each buffer
	btsFreeList chan []byte
	bufFreeList chan *bytes.Buffer
}

const leakyBufSize = 4108 // data.len(2) + hmacsha1(10) + data(4096)
const maxNBuf = 2048

var leakyBuf = NewLeakyBuf(maxNBuf, leakyBufSize)

// NewLeakyBuf creates a leaky buffer which can hold at most n buffer, each
// with bufSize bytes.
func NewLeakyBuf(n, bufSize int) *LeakyBuf {
	return &LeakyBuf{
		bufSize:     bufSize,
		btsFreeList: make(chan []byte, n),
		bufFreeList: make(chan *bytes.Buffer, n),
	}
}

// Get returns a buffer from the leaky buffer or create a new buffer.
func (lb *LeakyBuf) GetBts() (b []byte) {
	select {
	case b = <-lb.btsFreeList:
	default:
		b = make([]byte, lb.bufSize)
	}
	return
}

// Put add the buffer into the free buffer pool for reuse. Panic if the buffer
// size is not the same with the leaky buffer's. This is intended to expose
// error usage of leaky buffer.
func (lb *LeakyBuf) PutBts(b []byte) {
	if len(b) != lb.bufSize {
		panic("invalid buffer size that's put into leaky buffer")
	}
	select {
	case lb.btsFreeList <- b:
	default:
	}
	return
}

func (lb *LeakyBuf) GetBuf() (b *bytes.Buffer) {
	select {
	case b = <-lb.bufFreeList:
	default:
		b = new(bytes.Buffer)
	}
	return
}

func (lb *LeakyBuf) PutBuf(b *bytes.Buffer) {
	//丢弃占用内存较大的
	if b.Cap() > lb.bufSize {
		return
	}
	b.Reset()
	select {
	case lb.bufFreeList <- b:
	default:
	}
	return
}

func PutBuf(b *bytes.Buffer) {
	leakyBuf.PutBuf(b)
}

func GetBuf() *bytes.Buffer {
	return leakyBuf.GetBuf()
}
