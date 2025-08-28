package zipper

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/signalfx/golib/v3/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type foo struct{}

func (f *foo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(req.Body); err == nil {
		if buf.String() == "OK" {
			rw.WriteHeader(http.StatusOK)
		}
	}
	rw.WriteHeader(http.StatusBadRequest)
}

func TestZipper(t *testing.T) {
	zippers := NewZipper()
	badZippers := newZipper(func(io.Reader) (*gzip.Reader, error) {
		return new(gzip.Reader), errors.New("nope")
	})
	f := new(foo)
	zipped := new(bytes.Buffer)
	w := gzip.NewWriter(zipped)
	_, err := w.Write([]byte("OK"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	tests := []struct {
		zipper  *ReadZipper
		name    string
		data    []byte
		status  int
		headers map[string]string
	}{
		{zippers, "test non gzipped", []byte("OK"), http.StatusOK, map[string]string{}},
		{zippers, "test gzipped", zipped.Bytes(), http.StatusOK, map[string]string{"Content-Encoding": "gzip"}},
		{zippers, "test gzipped but bad", zipped.Bytes()[:5], http.StatusBadRequest, map[string]string{"Content-Encoding": "gzip"}},
		{badZippers, "test gzipped failure", zipped.Bytes(), http.StatusBadRequest, map[string]string{"Content-Encoding": "gzip"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), "GET", "/health-check", bytes.NewBuffer(test.data))
			for k, v := range test.headers {
				req.Header.Set(k, v)
			}
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			g := test.zipper.GzipHandler(f)
			g.ServeHTTP(rr, req)
			assert.Equal(t, test.status, rr.Code)
		})
	}
	t.Run("check datapoints", func(t *testing.T) {
		assert.Equal(t, 4, len(zippers.Datapoints()))
	})
}
