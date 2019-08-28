package ringbuffer

import (
	"bytes"
	"errors"
	"math"
)

var (
	ErrNegativeRead = errors.New("RingBuffer: reader returned negative count from Read")
	ErrBufferFull   = errors.New("RingBuffer: Buffer is full")
)

type RingBuffer struct {
	head       int
	tail       int
	capacity   int
	rightLimit int

	data []byte
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		capacity:   capacity,
		rightLimit: capacity + 1,
		data:       make([]byte, capacity+1),
	}
}

func (r *RingBuffer) Write(in []byte) (n int, err error) {
	if len(in) == 0 {
		return 0, nil
	}
	end := r.head - 1
	if end < 0 {
		end += r.rightLimit
	}
	if r.tail <= end {
		n += copy(r.data[r.tail:end], in)
	} else {
		n += copy(r.data[r.tail:r.rightLimit], in)
		if len(in) > n {
			n += copy(r.data[0:end], in[n:])
		}
	}
	r.tail = (r.tail + n) % r.rightLimit

	if n < len(in) {
		err = ErrBufferFull
	}

	return n, err
}

//不推荐大量使用该接口，会造成大量拷贝
//Deprecated
func (r *RingBuffer) Read(out []byte) (n int, err error) {
	return 0, nil
}

func (r *RingBuffer) NextBytes(n int) (result []byte) {
	if n >= r.capacity {
		return r.Bytes()
	}
	end := r.head + n
	rightLimit := int(math.Min(float64(end), float64(r.rightLimit)))
	result = r.data[r.head:rightLimit]
	if end > r.rightLimit {
		end -= r.rightLimit
		rightLimit = int(math.Min(float64(end), float64(r.tail)))
		result = append(result, r.data[0:rightLimit]...)
	}
	return
}

//Consume discard n bytes beginning from r.head, return head's pos in buffer
func (r *RingBuffer) Consume(n int) int {
	if n >= r.Used() {
		r.Reset()
	} else {
		if r.tail >= r.head {
			r.head = int(math.Min(float64(r.tail), float64(r.head+n)))
		} else {
			end := r.head + n
			if end >= r.rightLimit {
				end = end % r.rightLimit
				r.head = int(math.Min(float64(end), float64(r.head)))
			} else {
				r.head = int(math.Min(float64(end), float64(r.rightLimit)))
			}
		}
	}
	return r.head
}

func (r *RingBuffer) Bytes() []byte {
	if r.tail >= r.head {
		return r.data[r.head:r.tail]
	} else {
		var buf bytes.Buffer
		buf.Write(r.data[r.head:r.rightLimit])
		buf.Write(r.data[0:r.tail])
		return buf.Bytes()
	}
}

func (r *RingBuffer) Reset() {
	r.head = 0
	r.tail = 0
}

func (r *RingBuffer) Capacity() int {
	return r.capacity
}

func (r *RingBuffer) IsFull() bool {
	return r.Used() == r.capacity
}

func (r *RingBuffer) IsEmpty() bool {
	return r.Used() == 0
}

func (r *RingBuffer) Unused() int {
	return r.capacity - r.Used()
}

func (r *RingBuffer) Used() int {
	tail := r.tail
	if tail < r.head {
		tail += r.capacity + 1
	}
	return tail - r.head
}
