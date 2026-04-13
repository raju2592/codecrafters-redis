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

	fmt.Printf("Read request for bytes: %d\n", toRead)

	pos := 0

	netData := make([]byte, cr.bufSize)
	bufData := make([]byte, toRead)

	for true {
		r := cr.rb.Read(bufData[:toRead - pos])
		copy(buf[pos:], bufData[:r])
		pos += r

		if pos == toRead {
			fmt.Printf("returing dataa, %s", string(buf))
			cr.rb.Print()
			return toRead, nil
		}

		fmt.Printf("partial data: %s\n", string(buf))
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

func (cr *ConnectionReader) ReadByte() (byte, error) {
	netData := make([]byte, cr.bufSize)

	for cr.rb.Size() == 0 {
		if cr.connErr != nil {
			return 0, cr.connErr
		}

		n, err := cr.conn.Read(netData)

		if n > 0 {
			cr.rb.Write(netData[:n])
		}

		if err != nil {
			cr.connErr = err
		}
	}

	b, _ := cr.rb.ReadByte()
	return b, nil
}

