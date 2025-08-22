package grpc

import (
	"context"
	"testing"

	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCAuth(t *testing.T) {
	a := &SignalFxTokenAuth{
		Token:                    "test",
		DisableTransportSecurity: false,
	}

	assert.True(t, a.RequireTransportSecurity())

	md, err := a.GetRequestMetadata(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "test", md[sfxclient.TokenHeaderName])

	a.DisableTransportSecurity = true
	assert.False(t, a.RequireTransportSecurity())
}
