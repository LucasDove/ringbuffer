package ringbuffer

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestRingBuffer_Used(t *testing.T) {
	type fields struct {
		head     int
		tail     int
		capacity int
		data     []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "smoke",
			fields: fields{
				head:     0,
				tail:     10,
				capacity: 10,
				data:     nil,
			},
			want: 10,
		},
		{
			name: "smoke2",
			fields: fields{
				head:     0,
				tail:     9,
				capacity: 10,
				data:     nil,
			},
			want: 9,
		},
		{
			name: "smoke3",
			fields: fields{
				head:     2,
				tail:     1,
				capacity: 10,
				data:     nil,
			},
			want: 10,
		},
		{
			name: "smoke4",
			fields: fields{
				head:     10,
				tail:     9,
				capacity: 10,
				data:     nil,
			},
			want: 10,
		},
		{
			name: "smoke5",
			fields: fields{
				head:     5,
				tail:     2,
				capacity: 10,
				data:     nil,
			},
			want: 8,
		},
		{
			name: "smoke6",
			fields: fields{
				head:     2,
				tail:     8,
				capacity: 10,
				data:     nil,
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RingBuffer{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				capacity: tt.fields.capacity,
				data:     tt.fields.data,
			}
			if got := r.Used(); got != tt.want {
				t.Errorf("Used() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRingBuffer_Write(t *testing.T) {
	type fields struct {
		head       int
		tail       int
		capacity   int
		rightLimit int
		data       []byte
	}
	type args struct {
		in []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr bool
	}{
		{
			name: "write not enough bytes",
			fields: fields{
				head:       0,
				tail:       0,
				capacity:   10,
				rightLimit: 11,
				data:       make([]byte, 11),
			},
			args:    args{in: []byte("hello")},
			wantN:   5,
			wantErr: false,
		},
		{
			name: "write full bytes",
			fields: fields{
				head:       0,
				tail:       0,
				capacity:   10,
				rightLimit: 11,
				data:       make([]byte, 11),
			},
			args:    args{in: []byte("helloworld")},
			wantN:   10,
			wantErr: false,
		},
		{
			name: "write overflow bytes",
			fields: fields{
				head:       0,
				tail:       0,
				capacity:   10,
				rightLimit: 11,
				data:       make([]byte, 11),
			},
			args:    args{in: []byte("hello_world aha")},
			wantN:   10,
			wantErr: true,
		},
		{
			name: "write wrap aroud bytes",
			fields: fields{
				head:       2,
				tail:       8,
				capacity:   10,
				rightLimit: 11,
				data:       make([]byte, 11),
			},
			args:    args{in: []byte("hello_world aha")},
			wantN:   4,
			wantErr: true,
		},
		{
			name: "write wrap aroud bytes",
			fields: fields{
				head:       0,
				tail:       8,
				capacity:   10,
				rightLimit: 11,
				data:       make([]byte, 11),
			},
			args:    args{in: []byte("hello_world aha")},
			wantN:   2,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RingBuffer{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				capacity: tt.fields.capacity,
				size:     tt.fields.rightLimit,
				data:     tt.fields.data,
			}
			gotN, err := r.Write(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("Write() gotN = %v, want %v", gotN, tt.wantN)
			}
			fmt.Printf("%v\n", r.data)
		})
	}
}

func TestSliceCopy(t *testing.T) {
	src := []byte("0123456789-=abcd")
	dst := make([]byte, 10)
	n := copy(dst[0:5], src)
	fmt.Printf("len: %d, %v\n", n, dst)
	n = copy(dst[0:10], src)
	fmt.Printf("len: %d, %v\n", n, dst)
}

func TestRingBuffer_Consume1(t *testing.T) {
	type fields struct {
		head       int
		tail       int
		capacity   int
		rightLimit int
		data       []byte
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "normal",
			fields: fields{
				head:       0,
				tail:       9,
				capacity:   10,
				rightLimit: 11,
				data:       []byte("123456789"),
			},
			args: args{n: 3},
			want: 3,
		},
		{
			name: "wrap around",
			fields: fields{
				head:       8,
				tail:       2,
				capacity:   10,
				rightLimit: 11,
				data:       []byte("1234567890"),
			},
			args: args{n: 3},
			want: 0,
		},
		{
			name: "wrap around 2",
			fields: fields{
				head:       8,
				tail:       4,
				capacity:   10,
				rightLimit: 11,
				data:       []byte("1234567890"),
			},
			args: args{n: 4},
			want: 1,
		},
		{
			name: "wrap around 3",
			fields: fields{
				head:       8,
				tail:       4,
				capacity:   10,
				rightLimit: 11,
				data:       []byte("1234567890a"),
			},
			args: args{n: 7},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RingBuffer{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				capacity: tt.fields.capacity,
				size:     tt.fields.rightLimit,
				data:     tt.fields.data,
			}
			if got := r.Consume(tt.args.n); got != tt.want {
				t.Errorf("Consume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRingBuffer_Bytes(t *testing.T) {
	rb := NewRingBuffer(20)
	assert.True(t, rb.IsEmpty())

	n, err := rb.Write([]byte("12345"))
	assert.Equal(t, 5, n)
	assert.NoError(t, err)
	assert.Equal(t, []byte("12345"), rb.Bytes())

	n, err = rb.Write([]byte("67890-=abc"))
	assert.Equal(t, 10, n)
	assert.NoError(t, err)
	assert.Equal(t, []byte("1234567890-=abc"), rb.Bytes())

	n, err = rb.Write([]byte("ASDFG"))
	assert.Equal(t, 5, n)
	assert.NoError(t, err)
	assert.Equal(t, []byte("1234567890-=abcASDFG"), rb.Bytes())
	assert.True(t, rb.IsFull())

	n, err = rb.Write([]byte("QWER"))
	assert.Equal(t, 0, n)
	assert.EqualErrorf(t, err, "RingBuffer: Buffer is full", "")
	assert.Equal(t, []byte("1234567890-=abcASDFG"), rb.Bytes())

	rb.Consume(10)
	assert.Equal(t, []byte("-=abcASDFG"), rb.Bytes())
	rb.Consume(4)
	assert.Equal(t, []byte("cASDFG"), rb.Bytes())

	n, err = rb.Write([]byte("zxcvbnm"))
	assert.Equal(t, 7, n)
	assert.NoError(t, err)
	assert.Equal(t, []byte("cASDFGzxcvbnm"), rb.Bytes())

	n, err = rb.Write([]byte("1234567"))
	assert.Equal(t, 7, n)
	assert.NoError(t, err)
	assert.True(t, rb.IsFull())
	assert.Equal(t, []byte("cASDFGzxcvbnm1234567"), rb.Bytes())

	rb.Consume(7)
	assert.Equal(t, []byte("xcvbnm1234567"), rb.Bytes())
	assert.False(t, rb.IsFull())
	assert.Equal(t, 13, rb.Used())

	rb.Consume(12)
	assert.Equal(t, []byte("7"), rb.Bytes())
	assert.False(t, rb.IsFull())
	assert.False(t, rb.IsEmpty())
	assert.Equal(t, 1, rb.Used())

	n, err = rb.Write([]byte("abcd"))
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.False(t, rb.IsFull())
	assert.Equal(t, []byte("7abcd"), rb.Bytes())

	rb.Consume(5)
	assert.Equal(t, []byte(""), rb.Bytes())
	assert.False(t, rb.IsFull())
	assert.True(t, rb.IsEmpty())
	assert.Equal(t, 0, rb.Used())
}

type packet struct {
}

func (p *packet) Encode() []byte {
	container := make(map[string]int)
	kth := rand.Intn(3)
	for k := 0; k < kth; k++ {
		key := rand.Intn(2000)
		container[fmt.Sprintf("%d", key)] = key
	}
	data, _ := json.Marshal(&container)
	length := len(data) + 6
	raw := make([]byte, length)
	copy(raw[0:], "(")
	binary.LittleEndian.PutUint32(raw[1:5], uint32(length))
	copy(raw[5:], data)
	copy(raw[5+len(data):], ")")
	return raw
}

func (p *packet) Decode([]byte) {

}

type testJsonReader struct {
	ctx              context.Context
	data             []byte
	index            int
	readOneTimeLimit int
}

func newTestJsonReader(ctx context.Context, readOneTimeLimit int) *testJsonReader {
	pkt := &packet{}
	var rawData []byte
	limitLen := 20000000
	for len(rawData) < limitLen {
		rawData = append(rawData, pkt.Encode()...)
	}
	return &testJsonReader{ctx: ctx, data: rawData, readOneTimeLimit: readOneTimeLimit}
}

func (tjr *testJsonReader) Read(p []byte) (n int, err error) {
	if tjr.index >= len(tjr.data) {
		return 0, io.EOF
	}

	sendOneTime := rand.Intn(tjr.readOneTimeLimit)
	limit := tjr.index + sendOneTime
	if limit > len(tjr.data) {
		limit = len(tjr.data)
	}
	n = copy(p, tjr.data[tjr.index:limit])
	tjr.index += n
	err = nil
	return
}

func TestRingBuffer_Consume4(t *testing.T) {
	randomTest(t, 3921, 2000)
}

func TestRingBuffer_Consume(t *testing.T) {
	for i := 0; i < 200; i++ {
		bufferCap := randomRange(200, 8192)
		readOneTime := randomRange(10, 2000)
		randomTest(t, bufferCap, readOneTime)
	}
}

func randomTest(t *testing.T, bufferCap int, readLenOneTime int) {
	fmt.Printf("bufferCap: %d, readLenOneTime: %d\n", bufferCap, readLenOneTime)
	rb := NewRingBuffer(bufferCap)
	ctx, _ := context.WithCancel(context.Background())

	reader := newTestJsonReader(ctx, readLenOneTime)

	for {

		tmpBuffer := make([]byte, rb.Unused())
		n, err := reader.Read(tmpBuffer)
		_, _ = rb.Write(tmpBuffer[:n])
		//fmt.Printf("n:%d, err:%+v, content:%v\n", n, err, rb.Bytes())

		parseJsonJoinString(rb)

		if err == io.EOF {
			fmt.Println("reader drained")
			if rb.Used() != 0 {
				t.Error("there are remained data!!!\n")
			}
			break
		}
	}
}

func randomRange(left, right int) int {
	r := 0
	for r < left {
		r = rand.Intn(right)
	}
	return r
}

func parseJsonJoinString(rb *RingBuffer) {
	for !rb.IsEmpty() {
		//fmt.Printf("remained len:%d content:%v\n", rb.Used(), rb.Bytes())
		if rb.Used() < 5 {
			fmt.Printf("no more than 5 bytes, content:%+v\n", rb.Bytes())
			break
		}

		header := rb.NextBytes(5)
		if header[0] != '(' {
			log.Printf("start flag wrong, %v\n", rb.Bytes())
			break
		}

		dataLen := int(binary.LittleEndian.Uint32(header[1:5]))
		if rb.Used() < dataLen {
			fmt.Printf("no more than datalen(%d) bytes, content:%+v\n", dataLen, rb.Bytes())
			break
		}

		packet := rb.NextBytes(dataLen)
		if packet[dataLen-1] != ')' {
			log.Printf("end flag wrong, '%c'\n", packet[dataLen-1])
			break
		}

		if !json.Valid(packet[5 : dataLen-1]) {
			log.Printf("parse json wrong, %v\n", packet[5:dataLen-1])
			break
		}

		rb.Consume(dataLen)
		fmt.Printf("consume datalen:%d, head:%d, tail:%d, remainLen:%d\n", dataLen, rb.head, rb.tail, rb.Used())
	}
}

func TestRingBuffer_ReadFromTcpConn(t *testing.T) {
	go func() {
		ln, err := net.Listen("tcp", ":10086")
		if err != nil {
			panic("listen on 10086 failed")
		}
		defer ln.Close()

		tcpConn, err := ln.Accept()
		if err == nil {
			go func(conn net.Conn) {
				defer conn.Close()
				rb := NewRingBuffer(1024)
				for {
					_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 200))
					n, err := rb.ReadFromTcpConn(conn)

					if n > 0 {
						fmt.Printf("n: %d, length: %d, rb content: %v\n", n, rb.Used(), rb.Bytes())
						parseJsonJoinString(rb)
					}

					if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
						//超时，忽略，继续读
						continue
					} else if err != nil {
						fmt.Printf("err: %+v, conn closed\n", err)
						return
					}
				}
			}(tcpConn)
		}
	}()

	time.Sleep(time.Second * 2)

	go func() {
		conn, err := net.Dial("tcp", ":10086")
		if err != nil {
			panic(fmt.Sprintf("dial err: %+v", err))
		}
		defer conn.Close()

		timer := time.NewTimer(time.Second * 30)
		sender := newTestJsonReader(context.Background(), 2000)
		sendBuffer := make([]byte, 2000)
	SendLoop:
		for {
			select {
			case <-timer.C:
				break SendLoop
			default:
				n, err := sender.Read(sendBuffer)
				fmt.Printf("client send, n:%d, err:%+v\n", n, err)
				if n > 0 {
					_, nerr := conn.Write(sendBuffer[:n])
					if nerr != nil {
						fmt.Printf("send err:%+v\n", nerr)
					}
				}
				if err != nil {
					break SendLoop
				}
				time.Sleep(time.Second)
			}
		}
		fmt.Printf("client closed\n")
	}()

	<-context.Background().Done()
}
