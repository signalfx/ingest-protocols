package tagreplace

import (
	"context"
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type end struct{}

func (e *end) AddSpans(ctx context.Context, spans []*trace.Span) error {
	return nil
}

func (e *end) AddDatapoints(ctx context.Context, points []*datapoint.Datapoint) error {
	return nil
}

func (e *end) AddEvents(ctx context.Context, events []*event.Event) error {
	return nil
}

func TestTagReplace(t *testing.T) {
	cases := []struct {
		name       string
		rules      []string
		inputSpan  *trace.Span
		outputSpan *trace.Span
		exitEarly  bool
	}{
		{
			name:       "single replacement",
			rules:      []string{`^\/api\/v1\/document\/(?P<documentId>.*)\/update$`},
			inputSpan:  &trace.Span{Name: pointer.String("/api/v1/document/321083210/update")},
			outputSpan: &trace.Span{Name: pointer.String("/api/v1/document/{documentId}/update"), Tags: map[string]string{"documentId": "321083210"}},
			exitEarly:  false,
		},
		{
			name:       "multi replacement",
			rules:      []string{`^\/api\/(?P<version>.*)\/document\/(?P<documentId>.*)\/update$`},
			inputSpan:  &trace.Span{Name: pointer.String("/api/v1/document/321083210/update")},
			outputSpan: &trace.Span{Name: pointer.String("/api/{version}/document/{documentId}/update"), Tags: map[string]string{"documentId": "321083210", "version": "v1"}},
			exitEarly:  false,
		},
		{
			name: "exit early",
			rules: []string{
				`^\/api\/v1\/document\/(?P<documentId>.*)\/update$`,
				`^\/api\/(?P<version>.*)\/document\/(?P<documentId>.*)\/update$`,
			},
			inputSpan:  &trace.Span{Name: pointer.String("/api/v1/document/321083210/update")},
			outputSpan: &trace.Span{Name: pointer.String("/api/v1/document/{documentId}/update"), Tags: map[string]string{"documentId": "321083210"}},
			exitEarly:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &end{}
			tr, err := New(tc.rules, tc.exitEarly, e)
			require.NoError(t, err)
			require.NotNil(t, tr)
			err = tr.AddSpans(context.Background(), []*trace.Span{tc.inputSpan})
			require.NoError(t, err)

			assert.Equal(t, *tc.outputSpan.Name, *tc.inputSpan.Name)
			assert.Equal(t, tc.outputSpan.Tags, tc.inputSpan.Tags)
		})
	}
}

func Benchmark(b *testing.B) {
	spans := make([]*trace.Span, 0, b.N)
	for i := 0; i < b.N; i++ {
		spans = append(spans, &trace.Span{Name: pointer.String("/api/v1/document/321083210/update"), Tags: map[string]string{}})
	}
	tr, _ := New([]string{
		`^\/api\/(?P<version>.*)\/document\/(?P<documentId>.*)\/update$`,
	}, false, &end{})
	b.ResetTimer()
	b.ReportAllocs()
	tr.AddSpans(context.Background(), spans)
}

func TestTagReplaceBadRegex(t *testing.T) {
	_, err := New([]string{`ntId>.*)\/update$`}, false, &end{})
	assert.Error(t, err)
}

func TestTagReplaceNoSubExpRegex(t *testing.T) {
	_, err := New([]string{`^\/api\/version\/document\/documentId\/update$`}, false, &end{})
	assert.Error(t, err)
}

func TestTagReplaceNonNamedSubExpRegex(t *testing.T) {
	_, err := New([]string{`^\/api\/version\/document\/(.*)\/update$`}, false, &end{})
	assert.Error(t, err)
}

func TestTagReplacePassthroughs(t *testing.T) {
	tr := &TagReplace{next: &end{}}
	assert.NoError(t, tr.AddDatapoints(context.Background(), []*datapoint.Datapoint{}))
	assert.NoError(t, tr.AddEvents(context.Background(), []*event.Event{}))
}
