package main

import (
	"fmt"
	"net"
)

type ConnectionReader struct {
	conn net.Conn;
	rb *RingBuffer;
	bufSize int;
	connErr error;
}

func NewConnectionReader(conn net.Conn, bufSize int) *ConnectionReader {
	fmt.Println("Creating connection reader")
	return &ConnectionReader{
		conn: conn,
		rb: NewRingBuffer(bufSize),
		bufSize: bufSize,
		connErr: nil,
	}
}

func (cr *ConnectionReader) Read(buf []byte) (int, error) {
	toRead := len(buf)
	pos := 0

	netData := make([]byte, cr.bufSize)
	bufData := make([]byte, toRead)

	for true {
		r := cr.rb.Read(bufData[:toRead - pos])
		copy(buf[pos:], bufData[:r])
		pos += r

		if pos == toRead {
			return toRead, nil
		}

		if cr.connErr != nil {
			return pos, cr.connErr
		}

		n, err := cr.conn.Read(netData)

		if n > 0 {
			cr.rb.Write(netData[:n])
		}

		if err != nil {
			cr.connErr = err
		}
	}

	return pos, nil
}
