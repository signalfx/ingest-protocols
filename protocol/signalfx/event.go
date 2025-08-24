package signalfx

import (
	"bytes"
	"context"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	sfxmodel "github.com/signalfx/com_signalfx_metrics_protobuf/model"
	"github.com/signalfx/golib/v3/datapoint/dpsink"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/log"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/ingest-protocols/protocol"
	signalfxformat "github.com/signalfx/ingest-protocols/protocol/signalfx/format"
)

// ProtobufEventDecoderV2 decodes protocol buffers in signalfx's v2 format and sends them to Sink
type ProtobufEventDecoderV2 struct {
	Sink   dpsink.ESink
	Logger log.Logger
}

func (decoder *ProtobufEventDecoderV2) Read(ctx context.Context, req *http.Request) (err error) {
	jeff := buffs.Get().(*bytes.Buffer)
	defer buffs.Put(jeff)
	jeff.Reset()
	if err = protocol.ReadFromRequest(jeff, req, decoder.Logger); err != nil {
		return err
	}
	var msg sfxmodel.EventUploadMessage
	if err = proto.Unmarshal(jeff.Bytes(), &msg); err != nil {
		return err
	}
	evts := make([]*event.Event, 0, len(msg.GetEvents()))
	for _, protoDP := range msg.GetEvents() {
		if e, err1 := NewProtobufEvent(protoDP); err1 == nil {
			evts = append(evts, e)
		}
	}
	if len(evts) > 0 {
		err = decoder.Sink.AddEvents(ctx, evts)
	}
	return err
}

// JSONEventDecoderV2 decodes v2 json data for signalfx events and sends it to Sink
type JSONEventDecoderV2 struct {
	Sink   dpsink.ESink
	Logger log.Logger
}

func (decoder *JSONEventDecoderV2) Read(ctx context.Context, req *http.Request) error {
	var e signalfxformat.JSONEventV2
	if err := easyjson.UnmarshalFromReader(req.Body, &e); err != nil {
		return err
	}
	evts := make([]*event.Event, 0, len(e))
	for _, jsonEvent := range e {
		if jsonEvent.Category == nil {
			jsonEvent.Category = pointer.String("USER_DEFINED")
		}
		if jsonEvent.Timestamp == nil {
			jsonEvent.Timestamp = pointer.Int64(0)
		}
		cat := event.USERDEFINED
		if pbcat, ok := sfxmodel.EventCategory_value[*jsonEvent.Category]; ok {
			cat = event.Category(pbcat)
		}
		evt := event.NewWithProperties(jsonEvent.EventType, cat, jsonEvent.Dimensions, jsonEvent.Properties, fromTS(*jsonEvent.Timestamp))
		evts = append(evts, evt)
	}
	return decoder.Sink.AddEvents(ctx, evts)
}

// SetupJSONV2EventPaths tells the router which paths the given handler (which should handle v2 protobufs)
func SetupJSONV2EventPaths(r *mux.Router, handler http.Handler) {
	SetupJSONByPaths(r, handler, "/v2/event")
}

// SetupProtobufV2EventPaths tells the router which paths the given handler (which should handle v2 protobufs)
func SetupProtobufV2EventPaths(r *mux.Router, handler http.Handler) {
	SetupProtobufV2ByPaths(r, handler, "/v2/event")
}
