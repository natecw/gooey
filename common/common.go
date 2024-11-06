package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func ToBytes(i int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	return buf.Bytes(), err
}

func FromBytes(b []byte) (int32, error) {
	buf := bytes.NewReader(b)
	var result int32
	err := binary.Read(buf, binary.BigEndian, &result)
	return result, err
}

func Write(c net.Conn, msg string) error {
	bytes, err := ToBytes(int32(len([]byte(msg))))
	if err != nil {
		return err
	}

	_, err = c.Write(bytes)
	if err != nil {
		return err
	}

	_, err = c.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func Read(c net.Conn) (string, error) {
	b := make([]byte, 4)
	_, err := c.Read(b)
	if err != nil {
		return "", err
	}

	len, err := FromBytes(b)
	if err != err {
		return "", err
	}

	read := 0
	buf := make([]byte, len)
	for read < int(len) {
		n, err := c.Read(buf[read:])
		read += n
		if err == io.EOF {
			return "", fmt.Errorf("received EOF too soon")
		}
		if err != nil {
			return "", fmt.Errorf("error reading: %s", err.Error())
		}
	}
	return string(buf), nil
}
