package processdebug

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

func (e *end) AddSpans(context.Context, []*trace.Span) error {
	return nil
}

func (e *end) AddDatapoints(context.Context, []*datapoint.Datapoint) error {
	return nil
}

func (e *end) AddEvents(context.Context, []*event.Event) error {
	return nil
}

func Test(t *testing.T) {
	cases := []struct {
		desc       string
		inputSpan  *trace.Span
		outputSpan *trace.Span
	}{
		{
			desc:       "no debug or sampling priority tag",
			inputSpan:  &trace.Span{},
			outputSpan: &trace.Span{},
		},
		{
			desc:       "debug set with no sampling priority tag",
			inputSpan:  &trace.Span{Debug: pointer.Bool(true)},
			outputSpan: &trace.Span{Debug: pointer.Bool(true), Tags: map[string]string{"sampling.priority": "1"}},
		},
		{
			desc:       "no debug with sampling priority tag set",
			inputSpan:  &trace.Span{Tags: map[string]string{"sampling.priority": "1"}},
			outputSpan: &trace.Span{Debug: pointer.Bool(true), Tags: map[string]string{"sampling.priority": "1"}},
		},
		{
			desc:       "debug set with non-true sampling priority tag",
			inputSpan:  &trace.Span{Debug: pointer.Bool(true), Tags: map[string]string{"sampling.priority": "0"}},
			outputSpan: &trace.Span{Debug: pointer.Bool(true), Tags: map[string]string{"sampling.priority": "1"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			e := &end{}
			pd := New(e)
			require.NotNil(t, pd)
			err := pd.AddSpans(context.Background(), []*trace.Span{tc.inputSpan})
			assert.NoError(t, err)

			assert.Equal(t, tc.outputSpan.Debug, tc.inputSpan.Debug)
			assert.Equal(t, tc.outputSpan.Tags, tc.inputSpan.Tags)
		})
	}
}

func TestPassthroughs(t *testing.T) {
	at := &ProcessDebug{next: &end{}}
	assert.NoError(t, at.AddDatapoints(context.Background(), []*datapoint.Datapoint{}))
	assert.NoError(t, at.AddEvents(context.Background(), []*event.Event{}))
}
