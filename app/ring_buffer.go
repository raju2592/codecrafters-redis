package main

import (
	"errors"
	"fmt"
)

type RingBuffer struct {
	buf []byte;
	cap int;
	start int;
	end int;
}

func NewRingBuffer(cap int) *RingBuffer {
	return &RingBuffer{
		buf: make([]byte, cap),
		cap: cap,
		start: 0,
		end: 0,
	}
}

func (rb *RingBuffer) Size() int {
	return (rb.cap + rb.end - rb.start) % rb.cap
}

func (rb *RingBuffer) Write(data []byte) error {
	if rb.Size() + len(data) > rb.cap {
		return errors.New("Buffer size exceeded")
	}

	n := len(data)

	for i := 0; i < n; i++ {
		rb.buf[rb.end] = data[i]
		rb.end = (rb.end + 1) % rb.cap
	}

	return nil
}

func (rb *RingBuffer) Read(buf []byte) int {
	n := len(buf)
	toRead := n
	if toRead > rb.Size() {
		toRead = rb.Size()
	}

	for i := 0; i < toRead; i++ {
		buf[i] = rb.buf[rb.start]
		rb.start = (rb.start + 1) % rb.cap
	}

	return toRead
}

func (rb *RingBuffer) Print() {
	fmt.Printf("%d %d %d\n", rb.start, rb.end, rb.cap)
}