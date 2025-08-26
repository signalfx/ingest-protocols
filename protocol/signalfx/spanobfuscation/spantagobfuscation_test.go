package spanobfuscation

import (
	"context"
	"strconv"
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type obfend struct{}

func (e *obfend) AddSpans(context.Context, []*trace.Span) error {
	return nil
}

func (e *obfend) AddDatapoints(context.Context, []*datapoint.Datapoint) error {
	return nil
}

func (e *obfend) AddEvents(context.Context, []*event.Event) error {
	return nil
}

func TestObfuscate(t *testing.T) {
	config := []*TagMatchRuleConfig{
		{
			Service: pointer.String("test-service"),
			Tags:    []string{"obfuscate-me"},
		},
		{
			Service:   pointer.String("some*service"),
			Operation: pointer.String("sensitive*"),
			Tags:      []string{"PII", "SSN"},
		},
	}
	obf, err := NewObf(config, &obfend{})
	require.NoError(t, err)
	t.Run("should obfuscate tag from exact-match service", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("test-service", "shouldn't matter", map[string]string{"obfuscate-me": "val"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"obfuscate-me": OBFUSCATED}, spans[0].Tags)
	})
	t.Run("should not obfuscate tag from exact-match service as prefix", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("false-test-service", "shouldn't matter", map[string]string{"obfuscate-me": "val"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"obfuscate-me": "val"}, spans[0].Tags)
	})
	t.Run("should not obfuscate tag from exact-match service as suffix", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("test-service-extra", "shouldn't matter", map[string]string{"obfuscate-me": "val"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"obfuscate-me": "val"}, spans[0].Tags)
	})
	t.Run("should obfuscate tag from matching wildcard service and operation", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "sensitive-data-leak", map[string]string{"PII": "val"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"PII": OBFUSCATED}, spans[0].Tags)
	})
	t.Run("should not obfuscate tag with mismatched tag name", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "sensitive-data-leak", map[string]string{"obfuscate-me": "val"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"obfuscate-me": "val"}, spans[0].Tags)
	})
	t.Run("should not obfuscate tag with matching service but unmatched operation", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "secure-op", map[string]string{"PII": "val"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"PII": "val"}, spans[0].Tags)
	})
	t.Run("should obfuscate all tags defined in the removal rule", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "sensitive-data-leak", map[string]string{"PII": "val", "SSN": "111-22-3333"})}
		obf.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"PII": OBFUSCATED, "SSN": OBFUSCATED}, spans[0].Tags)
	})
	t.Run("should handle an empty span", func(t *testing.T) {
		spans := []*trace.Span{{}}
		err := obf.AddSpans(context.Background(), spans)
		assert.NoError(t, err)
	})
	t.Run("should handle a span with an empty service", func(t *testing.T) {
		spans := []*trace.Span{{LocalEndpoint: &trace.Endpoint{}}}
		err := obf.AddSpans(context.Background(), spans)
		assert.NoError(t, err)
	})
}

func TestNewObfMissingTags(t *testing.T) {
	_, err := NewObf([]*TagMatchRuleConfig{{}}, &obfend{})
	assert.Error(t, err)
}

func TestNewObfEmptyTagName(t *testing.T) {
	_, err := NewObf([]*TagMatchRuleConfig{{Tags: []string{""}}}, &obfend{})
	assert.Error(t, err)
}

func TestNewObfEmptyTagsArray(t *testing.T) {
	_, err := NewObf([]*TagMatchRuleConfig{{Tags: []string{}}}, &obfend{})
	assert.Error(t, err)
}

func TestObfPassthrough(t *testing.T) {
	obf := &SpanTagObfuscation{next: &obfend{}}
	assert.NoError(t, obf.AddDatapoints(context.Background(), []*datapoint.Datapoint{}))
	assert.NoError(t, obf.AddEvents(context.Background(), []*event.Event{}))
}

func BenchmarkObfOne(b *testing.B) {
	spans := make([]*trace.Span, 0, b.N)
	for i := 0; i < b.N; i++ {
		spans = append(spans, makeSpan("some-test-service", "test-op", map[string]string{"PII": "name", "otherTag": "ok"}))
	}
	config := []*TagMatchRuleConfig{
		{
			Service:   pointer.String("some*test*service"),
			Operation: pointer.String("test*"),
			Tags:      []string{"PII"},
		},
	}
	obf, _ := NewObf(config, &rmend{})
	b.ResetTimer()
	b.ReportAllocs()
	obf.AddSpans(context.Background(), spans)
}

func BenchmarkObfTen(b *testing.B) {
	spans := make([]*trace.Span, 0, b.N)
	for i := 0; i < b.N; i++ {
		spans = append(spans, makeSpan("some-test-service", "test-op", map[string]string{"PII": "name", "otherTag": "ok"}))
	}
	config := make([]*TagMatchRuleConfig, 0, 10)
	for i := 0; i < 9; i++ {
		rule := &TagMatchRuleConfig{
			Service:   pointer.String("some*test*service" + strconv.Itoa(i)),
			Operation: pointer.String("test*"),
			Tags:      []string{"PII"},
		}
		config = append(config, rule)
	}
	config = append(config, &TagMatchRuleConfig{
		Service:   pointer.String("some*test*service"),
		Operation: pointer.String("test*"),
		Tags:      []string{"PII"},
	})
	obf, _ := NewObf(config, &rmend{})
	b.ResetTimer()
	b.ReportAllocs()
	obf.AddSpans(context.Background(), spans)
}
