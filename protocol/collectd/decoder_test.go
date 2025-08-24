package collectd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// MockLogger implements log.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(keyValues ...interface{}) {
	m.Called(keyValues...)
}

func TestJSONDecoder_ServeHTTPC_Success(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoder{
		SendTo: mockSink,
		Logger: mockLogger,
	}

	// Mock successful datapoint processing only (testDecodeCollectdBody contains only datapoints)
	mockSink.On("AddDatapoints", mock.Anything, mock.MatchedBy(func(dps []*datapoint.Datapoint) bool {
		return len(dps) > 0
	})).Return(nil)

	req := httptest.NewRequest("POST", "/collectd", strings.NewReader(testDecodeCollectdBody))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	decoder.ServeHTTPC(context.Background(), rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Contains(t, rw.Body.String(), `"OK"`)
	assert.Equal(t, int64(0), decoder.TotalErrors)
	mockSink.AssertExpectations(t)
}

func TestJSONDecoder_ServeHTTPC_InvalidJSON(t *testing.T) {
	mockSink := new(MockSink)
	mockLogger := new(MockLogger)

	decoder := &JSONDecoder{
		SendTo: mockSink,
		Logger: mockLogger,
	}

	// Mock logger call for error
	mockLogger.On("Log", mock.Anything).Return()

	req := httptest.NewRequest("POST", "/collectd", strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	decoder.ServeHTTPC(context.Background(), rw, req)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Contains(t, rw.Body.String(), "Unable to decode json")
	assert.Equal(t, int64(1), decoder.TotalErrors)
}

func TestJSONDecoder_Read_Success(t *testing.T) {
	mockSink := new(MockSink)

	decoder := &JSONDecoder{
		SendTo: mockSink,
	}

	// Mock successful datapoint processing only (testDecodeCollectdBody contains only datapoints)
	mockSink.On("AddDatapoints", mock.Anything, mock.MatchedBy(func(dps []*datapoint.Datapoint) bool {
		return len(dps) > 0
	})).Return(nil)

	req := httptest.NewRequest("POST", "/collectd", strings.NewReader(testDecodeCollectdBody))
	err := decoder.Read(context.Background(), req)

	assert.NoError(t, err)
	mockSink.AssertExpectations(t)
}

func TestJSONDecoder_Read_InvalidJSON(t *testing.T) {
	mockSink := new(MockSink)

	decoder := &JSONDecoder{
		SendTo: mockSink,
	}

	req := httptest.NewRequest("POST", "/collectd", strings.NewReader(`invalid json`))
	err := decoder.Read(context.Background(), req)

	assert.Error(t, err)
}

func TestJSONDecoder_Read_WithEvents(t *testing.T) {
	mockSink := new(MockSink)

	decoder := &JSONDecoder{
		SendTo: mockSink,
	}

	// Mock event processing only
	mockSink.On("AddEvents", mock.Anything, mock.MatchedBy(func(events []*event.Event) bool {
		return len(events) == 1
	})).Return(nil)

	req := httptest.NewRequest("POST", "/collectd", strings.NewReader(testDecodeCollectdEventBody))
	err := decoder.Read(context.Background(), req)

	assert.NoError(t, err)
	mockSink.AssertExpectations(t)
}

func TestJSONDecoder_Read_SinkError(t *testing.T) {
	mockSink := new(MockSink)

	decoder := &JSONDecoder{
		SendTo: mockSink,
	}

	expectedErr := fmt.Errorf("sink error")
	mockSink.On("AddDatapoints", mock.Anything, mock.Anything).Return(expectedErr)

	req := httptest.NewRequest("POST", "/collectd", strings.NewReader(testDecodeCollectdBody))
	err := decoder.Read(context.Background(), req)

	assert.Error(t, err)
	mockSink.AssertExpectations(t)
}

func TestJSONDecoder_defaultDims_WithQueryParams(t *testing.T) {
	decoder := &JSONDecoder{}

	// Create request with query parameters
	req := httptest.NewRequest("POST", "/collectd?sfxdim_env=prod&sfxdim_service=web&regular_param=ignored", nil)

	dims := decoder.defaultDims(req)

	expected := map[string]string{
		"env":     "prod",
		"service": "web",
	}
	assert.Equal(t, expected, dims)
	assert.Equal(t, int64(0), decoder.TotalBlankDims)
}

func TestJSONDecoder_defaultDims_WithBlankValues(t *testing.T) {
	decoder := &JSONDecoder{}

	// Create request with blank dimension value
	req := httptest.NewRequest("POST", "/collectd?sfxdim_env=&sfxdim_service=web", nil)

	dims := decoder.defaultDims(req)

	expected := map[string]string{
		"service": "web",
	}
	assert.Equal(t, expected, dims)
	assert.Equal(t, int64(1), decoder.TotalBlankDims)
}

func TestJSONDecoder_defaultDims_NoSfxParams(t *testing.T) {
	decoder := &JSONDecoder{}

	req := httptest.NewRequest("POST", "/collectd?regular_param=value&another=test", nil)

	dims := decoder.defaultDims(req)

	assert.Empty(t, dims)
	assert.Equal(t, int64(0), decoder.TotalBlankDims)
}

func TestJSONDecoder_Datapoints(t *testing.T) {
	decoder := &JSONDecoder{
		TotalErrors:    5,
		TotalBlankDims: 3,
	}

	dps := decoder.Datapoints()

	assert.Len(t, dps, 2)

	var blankDimsDP, errorsDP *datapoint.Datapoint
	for _, dp := range dps {
		if dp.Metric == "total_blank_dims" {
			blankDimsDP = dp
		} else if dp.Metric == "invalid_collectd_json" {
			errorsDP = dp
		}
	}

	assert.NotNil(t, blankDimsDP)
	assert.NotNil(t, errorsDP)
	assert.Equal(t, int64(3), blankDimsDP.Value.(datapoint.IntValue).Int())
	assert.Equal(t, int64(5), errorsDP.Value.(datapoint.IntValue).Int())
	assert.Equal(t, datapoint.Counter, blankDimsDP.MetricType)
	assert.Equal(t, datapoint.Counter, errorsDP.MetricType)
}

func TestNewDataPoints_Success(t *testing.T) {
	f := &JSONWriteFormat{
		Dsnames:        []*string{stringPtr("value1"), stringPtr("value2")},
		Dstypes:        []*string{stringPtr("gauge"), stringPtr("counter")},
		Values:         []*float64{floatPtr(1.0), floatPtr(2.0)},
		Time:           floatPtr(1415062577.4960001),
		TypeS:          stringPtr("test"),
		TypeInstance:   stringPtr("instance"),
		Host:           stringPtr("test-host"),
		Plugin:         stringPtr("test-plugin"),
		PluginInstance: stringPtr("plugin-instance"),
		Interval:       floatPtr(10.0),
	}
	defaultDims := map[string]string{"env": "test"}

	dps := newDataPoints(f, defaultDims)

	assert.Len(t, dps, 2)
	assert.NotNil(t, dps[0])
	assert.NotNil(t, dps[1])
}

func TestNewDataPoints_MismatchedArrays(t *testing.T) {
	f := &JSONWriteFormat{
		Dsnames:        []*string{stringPtr("value1"), stringPtr("value2")},
		Dstypes:        []*string{stringPtr("gauge")}, // Only one type
		Values:         []*float64{floatPtr(1.0), floatPtr(2.0)},
		Time:           floatPtr(1415062577.4960001),
		TypeS:          stringPtr("test"),
		TypeInstance:   stringPtr("instance"),
		Host:           stringPtr("test-host"),
		Plugin:         stringPtr("test-plugin"),
		PluginInstance: stringPtr("plugin-instance"),
		Interval:       floatPtr(10.0),
	}
	defaultDims := map[string]string{"env": "test"}

	dps := newDataPoints(f, defaultDims)

	// Should only create datapoint for the first element since arrays don't match
	assert.Len(t, dps, 1)
}

func TestNewDataPoints_NilValue(t *testing.T) {
	f := &JSONWriteFormat{
		Dsnames:        []*string{stringPtr("value1"), stringPtr("value2")},
		Dstypes:        []*string{stringPtr("gauge"), stringPtr("counter")},
		Values:         []*float64{floatPtr(1.0), nil}, // Second value is nil
		Time:           floatPtr(1415062577.4960001),
		TypeS:          stringPtr("test"),
		TypeInstance:   stringPtr("instance"),
		Host:           stringPtr("test-host"),
		Plugin:         stringPtr("test-plugin"),
		PluginInstance: stringPtr("plugin-instance"),
		Interval:       floatPtr(10.0),
	}
	defaultDims := map[string]string{"env": "test"}

	dps := newDataPoints(f, defaultDims)

	// Should only create datapoint for the first element since second value is nil
	assert.Len(t, dps, 1)
}

func TestNewEvent_Success(t *testing.T) {
	f := &JSONWriteFormat{
		Time:           floatPtr(1435104306.0),
		Severity:       stringPtr("OKAY"),
		Message:        stringPtr("test message"),
		TypeS:          stringPtr("test"),
		TypeInstance:   stringPtr("instance"),
		Host:           stringPtr("test-host"),
		Plugin:         stringPtr("test-plugin"),
		PluginInstance: stringPtr("plugin-instance"),
	}
	defaultDims := map[string]string{"env": "test"}

	e := newEvent(f, defaultDims)

	assert.NotNil(t, e)
	assert.Equal(t, "test message", e.Properties["message"])
	assert.Equal(t, "OKAY", e.Properties["severity"])
}

func TestNewEvent_MissingFields(t *testing.T) {
	f := &JSONWriteFormat{
		Time:     floatPtr(1435104306.0),
		Severity: stringPtr("OKAY"),
		// Missing Message
	}
	defaultDims := map[string]string{"env": "test"}

	e := newEvent(f, defaultDims)

	assert.Nil(t, e)
}

func TestSetupCollectdPaths(t *testing.T) {
	r := mux.NewRouter()
	endpoint := "/v1/collectd"

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handled"))
	})

	SetupCollectdPaths(r, testHandler, endpoint)

	tests := []struct {
		name         string
		method       string
		contentType  string
		expectedCode int
	}{
		{
			name:         "Valid JSON content type",
			method:       "POST",
			contentType:  "application/json",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid JSON with charset",
			method:       "POST",
			contentType:  "application/json; charset=UTF-8",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty content type",
			method:       "POST",
			contentType:  "",
			expectedCode: http.StatusNotFound, // Route won't match because header doesn't match exactly
		},
		{
			name:         "Invalid content type",
			method:       "POST",
			contentType:  "text/plain",
			expectedCode: http.StatusBadRequest, // web.InvalidContentType returns 400
		},
		{
			name:         "GET method",
			method:       "GET",
			contentType:  "application/json",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, endpoint, nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rw := httptest.NewRecorder()

			r.ServeHTTP(rw, req)

			assert.Equal(t, tt.expectedCode, rw.Code)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
