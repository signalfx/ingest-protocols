package protocol

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/datapoint/dptest"
	"github.com/signalfx/golib/v3/log"
	"github.com/signalfx/golib/v3/nettest"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listenerServer struct {
	CloseableHealthCheck
	server   http.Server
	listener net.Listener
}

func (c *listenerServer) Close() error {
	err := c.listener.Close()
	return err
}

func (c *listenerServer) DebugDatapoints() []*datapoint.Datapoint {
	return c.HealthDatapoints()
}

func (c *listenerServer) DefaultDatapoints() []*datapoint.Datapoint {
	return []*datapoint.Datapoint{}
}

func (c *listenerServer) Datapoints() []*datapoint.Datapoint {
	return append(c.DebugDatapoints(), c.DefaultDatapoints()...)
}

var _ Listener = &listenerServer{}

func grabHealthCheck(t *testing.T, baseURI string, expectedStatus int) {
	t.Helper()
	client := http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), "GET", baseURI+"/healthz", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, expectedStatus, resp.StatusCode)
}

func TestHealthCheck(t *testing.T) {
	listenAddr := "127.0.0.1:0"
	sendTo := dptest.NewBasicSink()
	sendTo.Resize(1)
	listener, err := net.Listen("tcp", listenAddr)
	require.NoError(t, err)
	r := mux.NewRouter()
	fullHandler := web.NewHandler(context.Background(), web.FromHTTP(r))
	listenServer := &listenerServer{
		listener: listener,
		server: http.Server{
			Handler: fullHandler,
			Addr:    listener.Addr().String(),
		},
	}
	baseURI := fmt.Sprintf("http://127.0.0.1:%d", nettest.TCPPort(listener))
	require.NotNil(t, listener.Addr())
	listenServer.SetupHealthCheck(pointer.String("/healthz"), r, log.Discard)
	go func() {
		err := listenServer.server.Serve(listener)
		log.IfErr(log.Discard, err)
	}()

	// Should expose health check
	grabHealthCheck(t, baseURI, http.StatusOK)
	dp := listenServer.HealthDatapoints()
	assert.Equal(t, 1, len(dp))
	assert.EqualValues(t, 1, dp[0].Value)

	// Close the health check
	listenServer.CloseHealthCheck()
	grabHealthCheck(t, baseURI, http.StatusNotFound)
	dp = listenServer.HealthDatapoints()
	assert.Equal(t, 1, len(dp))
	assert.EqualValues(t, 2, dp[0].Value)
}
