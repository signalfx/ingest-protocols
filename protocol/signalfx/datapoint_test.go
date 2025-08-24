package signalfx

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	sfxmodel "github.com/signalfx/com_signalfx_metrics_protobuf/model"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/ingest-protocols/logkey"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testMetricName = "test.metric"

// MockDSink implements dpsink.DSink for testing
type MockDSink struct {
	mock.Mock
}

func (m *MockDSink) AddDatapoints(ctx context.Context, dps []*datapoint.Datapoint) error {
	args := m.Called(ctx, dps)
	return args.Error(0)
}

// MockSink implements dpsink.Sink for testing
type MockSink struct {
	mock.Mock
}

func (m *MockSink) AddDatapoints(ctx context.Context, dps []*datapoint.Datapoint) error {
	args := m.Called(ctx, dps)
	return args.Error(0)
}

func (m *MockSink) AddEvents(ctx context.Context, events []*event.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

// MockMetricTypeGetter implements MericTypeGetter for testing
type MockMetricTypeGetter struct {
	mock.Mock
}

func (m *MockMetricTypeGetter) GetMetricTypeFromMap(metricName string) sfxmodel.MetricType {
	args := m.Called(metricName)
	return args.Get(0).(sfxmodel.MetricType)
}

// MockLogger implements log.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(keyValues ...interface{}) {
	m.Called(keyValues...)
}

func TestDatapointProtobufIsInvalidForV1(t *testing.T) {
	tests := []struct {
		name     string
		msg      *sfxmodel.DataPoint
		expected bool
	}{
		{
			name:     "Empty metric name should be invalid",
			msg:      &sfxmodel.DataPoint{Metric: ""},
			expected: true,
		},
		{
			name:     "Missing metric name should be invalid",
			msg:      &sfxmodel.DataPoint{},
			expected: true,
		},
		{
			name: "No values should be invalid",
			msg: &sfxmodel.DataPoint{
				Metric: testMetricName,
				Value:  sfxmodel.Datum{},
			},
			expected: true,
		},
		{
			name: "Valid double value should be valid",
			msg: &sfxmodel.DataPoint{
				Metric: testMetricName,
				Value:  sfxmodel.Datum{DoubleValue: pointer.Float64(1.0)},
			},
			expected: false,
		},
		{
			name: "Valid int value should be valid",
			msg: &sfxmodel.DataPoint{
				Metric: testMetricName,
				Value:  sfxmodel.Datum{IntValue: pointer.Int64(1)},
			},
			expected: false,
		},
		{
			name: "Valid string value should be valid",
			msg: &sfxmodel.DataPoint{
				Metric: testMetricName,
				Value:  sfxmodel.Datum{StrValue: pointer.String("test")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := datapointProtobufIsInvalidForV1(tt.msg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProtobufDecoderV1_Read_Success(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &ProtobufDecoderV1{
		Sink:       mockSink,
		TypeGetter: mockTypeGetter,
		Logger:     mockLogger,
	}

	// Create a valid protobuf message
	msg := &sfxmodel.DataPoint{
		Metric: testMetricName,
		Value:  sfxmodel.Datum{DoubleValue: pointer.Float64(1.5)},
	}

	data, err := proto.Marshal(msg)
	assert.NoError(t, err)

	// Create protobuf stream with length prefix
	buf := &bytes.Buffer{}
	varint := proto.EncodeVarint(uint64(len(data)))
	buf.Write(varint)
	buf.Write(data)

	req := httptest.NewRequest("POST", "/datapoint", buf)

	// Mock expectations
	mockTypeGetter.On("GetMetricTypeFromMap", testMetricName).Return(sfxmodel.MetricType_GAUGE)
	mockSink.On("AddDatapoints", mock.Anything, mock.MatchedBy(func(dps []*datapoint.Datapoint) bool {
		return len(dps) == 1 && dps[0].Metric == testMetricName
	})).Return(nil)

	err = decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	mockSink.AssertExpectations(t)
	mockTypeGetter.AssertExpectations(t)
}

func TestProtobufDecoderV1_Read_InvalidProtobuf(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &ProtobufDecoderV1{
		Sink:       mockSink,
		TypeGetter: mockTypeGetter,
		Logger:     mockLogger,
	}

	// Create invalid protobuf data (empty metric)
	msg := &sfxmodel.DataPoint{
		Metric: "", // Empty metric should trigger validation error
		Value:  sfxmodel.Datum{DoubleValue: pointer.Float64(1.5)},
	}

	data, err := proto.Marshal(msg)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	varint := proto.EncodeVarint(uint64(len(data)))
	buf.Write(varint)
	buf.Write(data)

	req := httptest.NewRequest("POST", "/datapoint", buf)

	err = decoder.Read(context.Background(), req)
	assert.Equal(t, errInvalidProtobuf, err)
}

func TestProtobufDecoderV1_Read_ProtobufTooLarge(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &ProtobufDecoderV1{
		Sink:       mockSink,
		TypeGetter: mockTypeGetter,
		Logger:     mockLogger,
	}

	// Create a buffer with a varint indicating message too large
	buf := &bytes.Buffer{}
	varint := proto.EncodeVarint(uint64(40000)) // > 32768 limit
	buf.Write(varint)
	// Add some data so it doesn't immediately EOF
	buf.Write([]byte("dummy data"))

	req := httptest.NewRequest("POST", "/datapoint", buf)

	err := decoder.Read(context.Background(), req)
	assert.Equal(t, errProtobufTooLarge, err)
}

func TestProtobufDecoderV1_Read_InvalidVarint(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &ProtobufDecoderV1{
		Sink:       mockSink,
		TypeGetter: mockTypeGetter,
		Logger:     mockLogger,
	}

	// Create invalid varint data
	buf := &bytes.Buffer{}
	buf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // Invalid varint

	req := httptest.NewRequest("POST", "/datapoint", buf)

	err := decoder.Read(context.Background(), req)
	assert.Equal(t, errInvalidProtobufVarint, err)
}

func TestJSONDecoderV1_Read_Success(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV1{
		TypeGetter: mockTypeGetter,
		Sink:       mockSink,
		Logger:     mockLogger,
	}

	jsonData := `{"metric":"test.metric","value":1.5,"source":"test-source"}
{"metric":"test.metric2","value":2.5,"source":"test-source2"}`

	req := httptest.NewRequest("POST", "/v1/datapoint", strings.NewReader(jsonData))

	// Mock expectations
	mockTypeGetter.On("GetMetricTypeFromMap", testMetricName).Return(sfxmodel.MetricType_GAUGE)
	mockTypeGetter.On("GetMetricTypeFromMap", "test.metric2").Return(sfxmodel.MetricType_COUNTER)
	mockSink.On("AddDatapoints", mock.Anything, mock.MatchedBy(func(dps []*datapoint.Datapoint) bool {
		return len(dps) == 1
	})).Return(nil).Twice()

	err := decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	mockSink.AssertExpectations(t)
	mockTypeGetter.AssertExpectations(t)
}

func TestJSONDecoderV1_Read_EmptyMetric(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV1{
		TypeGetter: mockTypeGetter,
		Sink:       mockSink,
		Logger:     mockLogger,
	}

	jsonData := `{"metric":"","value":1.5,"source":"test-source"}`

	req := httptest.NewRequest("POST", "/v1/datapoint", strings.NewReader(jsonData))

	err := decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	// Should not call sink since metric is empty
	mockSink.AssertNotCalled(t, "AddDatapoints")
}

func TestProtobufDecoderV2_Read_Success(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &ProtobufDecoderV2{
		Sink:   mockSink,
		Logger: mockLogger,
	}

	// Create test message
	msg := &sfxmodel.DataPointUploadMessage{
		Datapoints: []*sfxmodel.DataPoint{
			{
				Metric: testMetricName,
				Value:  sfxmodel.Datum{DoubleValue: pointer.Float64(1.5)},
			},
		},
	}

	data, err := proto.Marshal(msg)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/v2/datapoint", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/x-protobuf")

	mockSink.On("AddDatapoints", mock.Anything, mock.MatchedBy(func(dps []*datapoint.Datapoint) bool {
		return len(dps) == 1 && dps[0].Metric == testMetricName
	})).Return(nil)

	err = decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	mockSink.AssertExpectations(t)
}

func TestJSONDecoderV2_Read_Success(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV2{
		Sink:   mockSink,
		Logger: mockLogger,
	}

	jsonData := `{
		"gauge": [
			{
				"metric": "test.metric",
				"value": 1.5,
				"timestamp": 1000000000000,
				"dimensions": {
					"host": "test-host"
				}
			}
		]
	}`

	req := httptest.NewRequest("POST", "/v2/datapoint", strings.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	mockSink.On("AddDatapoints", mock.Anything, mock.MatchedBy(func(dps []*datapoint.Datapoint) bool {
		return len(dps) == 1 && dps[0].Metric == testMetricName
	})).Return(nil)

	err := decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	mockSink.AssertExpectations(t)
}

func TestJSONDecoderV2_Read_UnknownMetricType(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV2{
		Sink:   mockSink,
		Logger: mockLogger,
	}

	jsonData := `{
		"unknown_type": [
			{
				"metric": "test.metric",
				"value": 1.5
			}
		]
	}`

	req := httptest.NewRequest("POST", "/v2/datapoint", strings.NewReader(jsonData))

	mockLogger.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	err := decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), atomic.LoadInt64(&decoder.unknownMetricType))
}

func TestJSONDecoderV2_Read_InvalidValue(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV2{
		Sink:   mockSink,
		Logger: mockLogger,
	}

	// JSON with nested object as value - this will cause ValueToValue to fail
	jsonData := `{
		"gauge": [
			{
				"metric": "test.metric",
				"value": {"invalid": "object"}
			}
		]
	}`

	req := httptest.NewRequest("POST", "/v2/datapoint", strings.NewReader(jsonData))

	mockLogger.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	err := decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), atomic.LoadInt64(&decoder.invalidValue))
}

func TestJSONDecoderV2_Datapoints(t *testing.T) {
	decoder := &JSONDecoderV2{
		unknownMetricType: 5,
		invalidValue:      3,
	}

	dps := decoder.Datapoints()
	assert.Len(t, dps, 2)

	var unknownTypeDP, invalidValueDP *datapoint.Datapoint
	for _, dp := range dps {
		if dp.Dimensions["reason"] == "unknown_metric_type" {
			unknownTypeDP = dp
		} else if dp.Dimensions["reason"] == "invalid_value" {
			invalidValueDP = dp
		}
	}

	assert.NotNil(t, unknownTypeDP)
	assert.NotNil(t, invalidValueDP)
	assert.Equal(t, int64(5), unknownTypeDP.Value.(datapoint.IntValue).Int())
	assert.Equal(t, int64(3), invalidValueDP.Value.(datapoint.IntValue).Int())
}

func TestSetupProtobufV2DatapointPaths(t *testing.T) {
	r := mux.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handled"))
	})

	SetupProtobufV2DatapointPaths(r, handler)

	req := httptest.NewRequest("POST", "/v2/datapoint", nil)
	req.Header.Set("Content-Type", "application/x-protobuf")
	rw := httptest.NewRecorder()

	r.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "handled", rw.Body.String())
}

func TestSetupProtobufV2ByPaths(t *testing.T) {
	r := mux.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("custom"))
	})

	SetupProtobufV2ByPaths(r, handler, "/custom/path")

	req := httptest.NewRequest("POST", "/custom/path", nil)
	req.Header.Set("Content-Type", "application/x-protobuf")
	rw := httptest.NewRecorder()

	r.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "custom", rw.Body.String())
}

func TestSetupJSONV2DatapointPaths(t *testing.T) {
	r := mux.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("json_v2"))
	})

	SetupJSONV2DatapointPaths(r, handler)

	req := httptest.NewRequest("POST", "/v2/datapoint", nil)
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	r.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "json_v2", rw.Body.String())
}

func TestSetupProtobufV1Paths(t *testing.T) {
	r := mux.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("v1_protobuf"))
	})

	SetupProtobufV1Paths(r, handler)

	tests := []string{"/datapoint", "/v1/datapoint"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, nil)
			req.Header.Set("Content-Type", "application/x-protobuf")
			rw := httptest.NewRecorder()

			r.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, "v1_protobuf", rw.Body.String())
		})
	}
}

func TestSetupJSONV1Paths(t *testing.T) {
	r := mux.NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("v1_json"))
	})

	SetupJSONV1Paths(r, handler)

	tests := []string{"/datapoint", "/v1/datapoint"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, nil)
			req.Header.Set("Content-Type", "application/json")
			rw := httptest.NewRecorder()

			r.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, "v1_json", rw.Body.String())
		})
	}
}

func TestGetTokenLogFormat_WithToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set(sfxclient.TokenHeaderName, "test-token-123")

	result := getTokenLogFormat(req)

	assert.Len(t, result, 4)
	// Should contain SHA1 hash and caller (first half of token)
	assert.Contains(t, result, logkey.SHA1)
	assert.Contains(t, result, logkey.Caller)
}

func TestGetTokenLogFormat_EmptyToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set(sfxclient.TokenHeaderName, "")

	result := getTokenLogFormat(req)

	assert.Empty(t, result)
}

func TestGetTokenLogFormat_NoToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)

	result := getTokenLogFormat(req)

	assert.Empty(t, result)
}

// Test error cases for various decoders
func TestProtobufDecoderV1_Read_SinkError(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &ProtobufDecoderV1{
		Sink:       mockSink,
		TypeGetter: mockTypeGetter,
		Logger:     mockLogger,
	}

	msg := &sfxmodel.DataPoint{
		Metric: testMetricName,
		Value:  sfxmodel.Datum{DoubleValue: pointer.Float64(1.5)},
	}

	data, err := proto.Marshal(msg)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	varint := proto.EncodeVarint(uint64(len(data)))
	buf.Write(varint)
	buf.Write(data)

	req := httptest.NewRequest("POST", "/datapoint", buf)

	expectedErr := fmt.Errorf("sink error")
	mockTypeGetter.On("GetMetricTypeFromMap", testMetricName).Return(sfxmodel.MetricType_GAUGE)
	mockSink.On("AddDatapoints", mock.Anything, mock.Anything).Return(expectedErr)
	mockLogger.On("Log", mock.Anything, mock.Anything).Return()

	err = decoder.Read(context.Background(), req)
	assert.NoError(t, err) // Read continues despite sink errors
	mockSink.AssertExpectations(t)
	mockTypeGetter.AssertExpectations(t)
}

func TestJSONDecoderV1_Read_InvalidJSON(t *testing.T) {
	mockSink := new(MockDSink)
	mockTypeGetter := new(MockMetricTypeGetter)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV1{
		TypeGetter: mockTypeGetter,
		Sink:       mockSink,
		Logger:     mockLogger,
	}

	invalidJSON := `{"invalid": json}`

	req := httptest.NewRequest("POST", "/v1/datapoint", strings.NewReader(invalidJSON))

	err := decoder.Read(context.Background(), req)
	assert.Error(t, err)
}

func TestJSONDecoderV2_Read_InvalidJSON(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV2{
		Sink:   mockSink,
		Logger: mockLogger,
	}

	invalidJSON := `invalid json`

	req := httptest.NewRequest("POST", "/v2/datapoint", strings.NewReader(invalidJSON))

	err := decoder.Read(context.Background(), req)
	assert.Equal(t, errInvalidJSONFormat, err)
}

func TestJSONDecoderV2_Read_EmptyDatapoints(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoderV2{
		Sink:   mockSink,
		Logger: mockLogger,
	}

	jsonData := `{}`

	req := httptest.NewRequest("POST", "/v2/datapoint", strings.NewReader(jsonData))

	err := decoder.Read(context.Background(), req)
	assert.NoError(t, err)
	// Should not call sink since no datapoints
	mockSink.AssertNotCalled(t, "AddDatapoints")
}

func TestGetLogTokenFormat(t *testing.T) {
	Convey("test log token format", t, func() {
		req := httptest.NewRequest("get", "/index.html", nil)
		req.Header.Set(sfxclient.TokenHeaderName, "FIRST_HALF")
		ret := getTokenLogFormat(req)
		So(ret, ShouldResemble, []interface{}{logkey.SHA1, "dg7tp+xlWq6sb2Aj6lyRvaIYaXY=", logkey.Caller, "FIRST"})
		req.Header.Set(sfxclient.TokenHeaderName, "")
		ret = getTokenLogFormat(req)
		So(len(ret), ShouldResemble, 0)
	})
}
