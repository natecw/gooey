package tcp

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

const MAX_LENGTH = 10_000

var ErrPacketSize = errors.New("max packet size exceeded")

type Connection struct {
	Reader   FrameReader
	Writer   FrameWriter
	Id       int
	Username string
	conn     net.Conn

	previous []byte
	scratch  [1024]byte
}

type FrameReader struct {
	Reader   io.Reader
	previous []byte
	scratch  []byte
}

type FrameWriter struct {
	Writer io.Writer
}

func NewConnection(conn net.Conn, id int) Connection {
	return Connection{
		Reader: NewReader(conn),
		Writer: NewWriter(conn),
		Id:     id,
		conn:   conn,
	}
}

func NewReader(reader io.Reader) FrameReader {
	return FrameReader{
		Reader:   reader,
		previous: []byte{},
		scratch:  make([]byte, 1024),
	}
}

func NewWriter(writer io.Writer) FrameWriter {
	return FrameWriter{
		Writer: writer,
	}
}

func (c *Connection) Close() {
	c.conn.Close()
}

func (c *Connection) Next() (*Command, error) {
	cmdBytes, err := c.Reader.Read()
	if err != nil {
		return nil, err
	}

	var cmd Command
	err = cmd.UnmarshalBinary(cmdBytes)
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

func (r *FrameReader) Read() ([]byte, error) {
	for {
		n := r.packetLen(r.previous)
		if n > MAX_LENGTH {
			return nil, fmt.Errorf("FrameReader.Read %d %w", n, ErrPacketSize)
		}

		if r.canParse(r.previous) {
			out := r.previous[:n]
			rem := len(r.previous) - n
			new := make([]byte, rem)
			copy(new, r.previous[n:])
			r.previous = new
			return out, nil
		}

		n, err := r.Reader.Read(r.scratch)
		if err != nil {
			return nil, err
		}
		r.previous = append(r.previous, r.scratch[:n]...)
	}
}

func (w *FrameWriter) Write(enc encoding.BinaryMarshaler) error {
	data, err := enc.MarshalBinary()
	if err != nil {
		return err
	}
	read := len(data)
	for read > 0 {
		n, err := w.Writer.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
		read -= n
	}
	return nil
}

func (r *FrameReader) packetLen(p []byte) int {
	if len(p) < HEADER_SIZE {
		return -1
	}
	return int(binary.BigEndian.Uint16(p[2:])) + HEADER_SIZE
}

func (r *FrameReader) canParse(p []byte) bool {
	if len(p) < HEADER_SIZE {
		return false
	}
	length := int(binary.BigEndian.Uint16(p[2:]))
	return len(p) >= length+HEADER_SIZE
}
