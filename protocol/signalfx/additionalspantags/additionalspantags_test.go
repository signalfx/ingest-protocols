package additionalspantags

import (
	"context"
	"strconv"
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type end struct{}

func (e *end) AddSpans(context.Context, []*trace.Span) error {
	return nil
}

func (e *end) AddDatapoints(context.Context, []*datapoint.Datapoint) error {
	return nil
}

func (e *end) AddEvents(context.Context, []*event.Event) error {
	return nil
}

func TestAdditionalSpanTags(t *testing.T) {
	cases := []struct {
		desc       string
		tags       map[string]string
		inputSpan  *trace.Span
		outputSpan *trace.Span
	}{
		{
			desc: "test add single tag",
			tags: map[string]string{
				"tagKey": "tagValue",
			},
			inputSpan:  &trace.Span{Tags: map[string]string{}},
			outputSpan: &trace.Span{Tags: map[string]string{"tagKey": "tagValue"}},
		},
		{
			desc: "test add single tag to nil Tags",
			tags: map[string]string{
				"tagKey": "tagValue",
			},
			inputSpan:  &trace.Span{},
			outputSpan: &trace.Span{Tags: map[string]string{"tagKey": "tagValue"}},
		},
		{
			desc: "test add multiple tags",
			tags: map[string]string{
				"tagKey":    "tagValue",
				"secondTag": "secondValue",
			},
			inputSpan:  &trace.Span{},
			outputSpan: &trace.Span{Tags: map[string]string{"tagKey": "tagValue", "secondTag": "secondValue"}},
		},
		{
			desc: "test add single tag to already existing Tags",
			tags: map[string]string{
				"tagKey": "tagValue",
			},
			inputSpan:  &trace.Span{Tags: map[string]string{"existingKey": "existingValue"}},
			outputSpan: &trace.Span{Tags: map[string]string{"existingKey": "existingValue", "tagKey": "tagValue"}},
		},
		{
			desc: "test overwrite existing tag",
			tags: map[string]string{
				"tagKey": "tagValue",
			},
			inputSpan:  &trace.Span{Tags: map[string]string{"tagKey": "wrongValue"}},
			outputSpan: &trace.Span{Tags: map[string]string{"tagKey": "tagValue"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := &end{}
			at := New(tc.tags, e)
			require.NotNil(t, at)
			err := at.AddSpans(context.Background(), []*trace.Span{tc.inputSpan})
			require.NoError(t, err)

			assert.Equal(t, tc.outputSpan.Tags, tc.inputSpan.Tags)
		})
	}
}

func TestPassthroughs(t *testing.T) {
	at := &AdditionalSpanTags{next: &end{}}
	assert.NoError(t, at.AddDatapoints(context.Background(), []*datapoint.Datapoint{}))
	assert.NoError(t, at.AddEvents(context.Background(), []*event.Event{}))
}

func Benchmark(b *testing.B) {
	spans := make([]*trace.Span, 0, b.N)
	for i := 0; i < b.N; i++ {
		spans = append(spans, &trace.Span{})
	}
	tagsConfig := make(map[string]string, 100)
	var i int64
	for i = 0; i < 100; i++ {
		kv := strconv.FormatInt(i, 10)
		tagsConfig[kv] = kv
	}
	addTags := AdditionalSpanTags{
		tags: tagsConfig,
		next: &end{},
	}
	b.ResetTimer()
	b.ReportAllocs()
	addTags.AddSpans(context.Background(), spans)
}
