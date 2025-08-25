package protocol

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/signalfx/golib/v3/log"
	"github.com/stretchr/testify/assert"
)

var errRead = errors.New("could not read")

type testReader struct {
	content []byte
	err     error
}

func (tr *testReader) Read(b []byte) (int, error) {
	if tr.err != nil {
		return 0, tr.err
	}
	n := copy(b, tr.content)
	return n, io.EOF
}

func TestReadFromRequest(t *testing.T) {
	reader := &testReader{}
	var req http.Request
	out := bytes.NewBuffer([]byte{})

	t.Run("good read", func(t *testing.T) {
		reader.content = []byte{0x05}
		req.Body = io.NopCloser(reader)
		req.ContentLength = 1
		err := ReadFromRequest(out, &req, log.Discard)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x05}, out.Bytes())
	})

	t.Run("bad read", func(t *testing.T) {
		reader.err = errRead
		req.Body = io.NopCloser(reader)
		err := ReadFromRequest(out, &req, log.Discard)
		assert.Equal(t, errRead, err)
	})
}
