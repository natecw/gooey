package tcp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnection(t *testing.T) {
	cmd := &Command{
		Command: byte('m'),
		Data:    []byte("hello world"),
	}

	b := []byte{}
	bin, err := cmd.MarshalBinary()
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		require.NoError(t, err)
		b = append(b, bin...)
	}

	r := bytes.NewReader(b)
	w := bytes.NewBuffer(nil)

	conn := Connection{
		Reader: NewReader(r),
		Writer: NewWriter(w),
		Id:     0,
	}

	for i := 0; i < 100; i++ {
		response, err := conn.Next()
		require.NoError(t, err)
		require.Equal(t, cmd, response)
	}
}
