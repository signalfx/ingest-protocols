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

type rmend struct{}

func (e *rmend) AddSpans(context.Context, []*trace.Span) error {
	return nil
}

func (e *rmend) AddDatapoints(context.Context, []*datapoint.Datapoint) error {
	return nil
}

func (e *rmend) AddEvents(context.Context, []*event.Event) error {
	return nil
}

func TestTagDelete(t *testing.T) {
	config := []*TagMatchRuleConfig{
		{
			Service: pointer.String("test-service"),
			Tags:    []string{"delete-me"},
		},
		{
			Service:   pointer.String("some*service"),
			Operation: pointer.String("sensitive*"),
			Tags:      []string{"PII", "SSN"},
		},
	}
	so, err := NewRm(config, &rmend{})
	require.NoError(t, err)
	t.Run("should remove tag from exact-match service", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("test-service", "shouldn't matter", map[string]string{"delete-me": "val"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{}, spans[0].Tags)
	})
	t.Run("should not remove tag from exact-match service as prefix", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("false-test-service", "shouldn't matter", map[string]string{"delete-me": "val"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"delete-me": "val"}, spans[0].Tags)
	})
	t.Run("should not remove tag from exact-match service as suffix", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("test-service-extra", "shouldn't matter", map[string]string{"delete-me": "val"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"delete-me": "val"}, spans[0].Tags)
	})
	t.Run("should remove tag from matching wildcard service and operation", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "sensitive-data-leak", map[string]string{"PII": "val"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{}, spans[0].Tags)
	})
	t.Run("should not remove tag with mismatched tag name", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "sensitive-data-leak", map[string]string{"delete-me": "val"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"delete-me": "val"}, spans[0].Tags)
	})
	t.Run("should not remove tag with matching service but unmatched operation", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "secure-op", map[string]string{"PII": "val"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{"PII": "val"}, spans[0].Tags)
	})
	t.Run("should remove all tags defined in the removal rule", func(t *testing.T) {
		spans := []*trace.Span{makeSpan("some-test-service", "sensitive-data-leak", map[string]string{"PII": "val", "SSN": "111-22-3333"})}
		so.AddSpans(context.Background(), spans)
		assert.Equal(t, map[string]string{}, spans[0].Tags)
	})
	t.Run("should handle an empty span", func(t *testing.T) {
		spans := []*trace.Span{{}}
		err := so.AddSpans(context.Background(), spans)
		assert.NoError(t, err)
	})
	t.Run("should handle a span with an empty service", func(t *testing.T) {
		spans := []*trace.Span{{LocalEndpoint: &trace.Endpoint{}}}
		err := so.AddSpans(context.Background(), spans)
		assert.NoError(t, err)
	})
}

func TestNewRmMissingTags(t *testing.T) {
	_, err := NewRm([]*TagMatchRuleConfig{{}}, &rmend{})
	assert.Error(t, err)
}

func TestNewRmEmptyTagName(t *testing.T) {
	_, err := NewRm([]*TagMatchRuleConfig{{Tags: []string{""}}}, &rmend{})
	assert.Error(t, err)
}

func TestNewRmEmptyTagsArray(t *testing.T) {
	_, err := NewRm([]*TagMatchRuleConfig{{Tags: []string{}}}, &rmend{})
	assert.Error(t, err)
}

func TestRmPassthrough(t *testing.T) {
	so := &SpanTagRemoval{next: &rmend{}}
	assert.NoError(t, so.AddDatapoints(context.Background(), []*datapoint.Datapoint{}))
	assert.NoError(t, so.AddEvents(context.Background(), []*event.Event{}))
}

func BenchmarkRmOne(b *testing.B) {
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
	so, _ := NewRm(config, &rmend{})
	b.ResetTimer()
	b.ReportAllocs()
	so.AddSpans(context.Background(), spans)
}

func BenchmarkRmTen(b *testing.B) {
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
	so, _ := NewRm(config, &rmend{})
	b.ResetTimer()
	b.ReportAllocs()
	so.AddSpans(context.Background(), spans)
}

func makeSpan(service string, operation string, tags map[string]string) *trace.Span {
	localEndpoint := &trace.Endpoint{
		ServiceName: pointer.String(service),
	}
	return &trace.Span{
		Name:          pointer.String(operation),
		LocalEndpoint: localEndpoint,
		Tags:          tags,
	}
}
