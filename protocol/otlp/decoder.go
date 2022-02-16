package otlp

import (
	"bytes"
	"context"
	"net/http"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint/dpsink"
	"github.com/signalfx/golib/v3/log"
	"github.com/signalfx/ingest-protocols/protocol"
	"github.com/signalfx/ingest-protocols/protocol/signalfx"
	metricsservicev1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/protobuf/proto"
)

type httpMetricDecoder struct {
	sink   dpsink.Sink
	logger log.Logger
	buffs  sync.Pool
	logDPs bool
}

func NewHTTPMetricDecoder(sink dpsink.Sink, logger log.Logger, logDPs bool) signalfx.ErrorReader {
	return &httpMetricDecoder{
		sink:   sink,
		logger: logger,
		buffs: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		logDPs: logDPs,
	}
}

func (d *httpMetricDecoder) Read(ctx context.Context, req *http.Request) (err error) {
	jeff := d.buffs.Get().(*bytes.Buffer)
	defer d.buffs.Put(jeff)
	jeff.Reset()
	if err = protocol.ReadFromRequest(jeff, req, d.logger); err != nil {
		return err
	}
	var msg metricsservicev1.ExportMetricsServiceRequest
	if err = proto.Unmarshal(jeff.Bytes(), &msg); err != nil {
		return err
	}
	dps := FromOTLPMetricRequest(&msg)
	if d.logDPs {
		d.logger.Log("datapoints", spew.Sdump(dps), "original", msg, "Decoded OTLP datapoints")
	}
	if len(dps) > 0 {
		err = d.sink.AddDatapoints(ctx, dps)
	}
	return nil
}
