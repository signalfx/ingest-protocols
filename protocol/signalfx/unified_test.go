package signalfx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/datapoint/dpsink"
	"github.com/signalfx/golib/v3/datapoint/dptest"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/log"
	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
)

type expect struct {
	count     int
	forwardTo Sink
}

func (e *expect) AddEvents(ctx context.Context, events []*event.Event) error {
	if len(events) != e.count {
		panic("NOPE")
	}
	if e.forwardTo != nil {
		events = append(events, nil)
		log.IfErr(log.Panic, e.forwardTo.AddEvents(ctx, events))
	}
	return nil
}

func (e *expect) AddDatapoints(ctx context.Context, points []*datapoint.Datapoint) error {
	if len(points) != e.count {
		panic("NOPE")
	}
	if e.forwardTo != nil {
		points = append(points, nil)
		log.IfErr(log.Panic, e.forwardTo.AddDatapoints(ctx, points))
	}
	return nil
}

func (e *expect) AddSpans(ctx context.Context, spans []*trace.Span) error {
	if len(spans) != e.count {
		panic("NOPE")
	}
	if e.forwardTo != nil {
		spans = append(spans, nil)
		log.IfErr(log.Panic, e.forwardTo.AddSpans(ctx, spans))
	}
	return nil
}

func (e *expect) next(sendTo Sink) Sink {
	return &expect{
		count:     e.count,
		forwardTo: sendTo,
	}
}

func TestFromChain(t *testing.T) {
	e2 := expect{count: 2}
	e1 := expect{count: 1}
	e0 := expect{count: 0}

	chain := FromChain(&e2, e0.next, e1.next)
	assert.NoError(t, chain.AddDatapoints(context.TODO(), []*datapoint.Datapoint{}))
	assert.NoError(t, chain.AddEvents(context.TODO(), []*event.Event{}))
}

func TestIncludingDimensions(t *testing.T) {
	end := dptest.NewBasicSink()
	end.Resize(1)
	addInto := IncludingDimensions(map[string]string{"name": "jack"}, end)
	ctx := context.Background()
	t.Run("no dimensions should be identity function", func(t *testing.T) {
		identityAddInto := IncludingDimensions(nil, end)
		assert.Equal(t, end, identityAddInto)
	})

	t.Run("appending dims should work for datapoints", func(t *testing.T) {
		dp := dptest.DP()
		assert.NoError(t, addInto.AddDatapoints(ctx, []*datapoint.Datapoint{dp}))
		dpOut := end.Next()
		assert.Equal(t, "jack", dpOut.Dimensions["name"])

		w := &WithDimensions{}
		assert.Equal(t, 0, len(w.appendDimensions(nil)))
	})
	t.Run("appending dims should work for events", func(t *testing.T) {
		e := dptest.E()
		assert.NoError(t, addInto.AddEvents(ctx, []*event.Event{e}))
		eOut := end.NextEvent()
		assert.Equal(t, "jack", eOut.Dimensions["name"])

		w := &WithDimensions{}
		assert.Equal(t, 0, len(w.appendDimensionsEvents(nil)))
	})
	t.Run("appending dims should be a pass through for traces", func(t *testing.T) {
		assert.NoError(t, addInto.AddSpans(ctx, []*trace.Span{{}}))
	})
}

type boolFlagCheck bool

func (b *boolFlagCheck) HasFlag(context.Context) bool {
	return bool(*b)
}

func setupFilterTest() (*dpsink.ItemFlagger, Sink, []*datapoint.Datapoint, []*event.Event) {
	flagCheck := boolFlagCheck(false)
	i := &dpsink.ItemFlagger{
		CtxFlagCheck:        &flagCheck,
		EventMetaName:       "my_events",
		MetricDimensionName: "sf_metric",
		Logger:              log.Discard,
	}
	dp1 := datapoint.New("mname", map[string]string{"org": "mine", "type": "prod"}, nil, datapoint.Gauge, time.Time{})
	dp2 := datapoint.New("mname2", map[string]string{"org": "another", "type": "prod"}, nil, datapoint.Gauge, time.Time{})
	ev1 := event.New("mname", event.USERDEFINED, map[string]string{"org": "mine", "type": "prod"}, time.Time{})
	ev2 := event.New("mname2", event.USERDEFINED, map[string]string{"org": "another", "type": "prod"}, time.Time{})
	chain := FromChain(dpsink.Discard, NextWrap(UnifyNextSinkWrap(i)))
	return i, chain, []*datapoint.Datapoint{dp1, dp2}, []*event.Event{ev1, ev2}
}

func TestFilterShouldNotFlagByDefault(t *testing.T) {
	i, chain, dps, events := setupFilterTest()

	assert.Equal(t, 4, len(i.Datapoints()))
	assert.NoError(t, chain.AddDatapoints(context.Background(), dps))
	assert.False(t, i.HasDatapointFlag(dps[0]))
	assert.False(t, i.HasDatapointFlag(dps[1]))

	assert.NoError(t, chain.AddEvents(context.Background(), events))
	assert.False(t, i.HasEventFlag(events[0]))
	assert.False(t, i.HasEventFlag(events[1]))
	assert.NoError(t, i.AddSpans(context.Background(), []*trace.Span{}, dpsink.Discard))
}

func TestFilterShouldFlagIfContextFlagged(t *testing.T) {
	i, chain, dps, events := setupFilterTest()
	fc := boolFlagCheck(true)
	i.CtxFlagCheck = &fc

	assert.NoError(t, chain.AddDatapoints(context.Background(), dps))
	assert.True(t, i.HasDatapointFlag(dps[0]))
	assert.True(t, i.HasDatapointFlag(dps[1]))

	assert.NoError(t, chain.AddEvents(context.Background(), events))
	assert.True(t, i.HasEventFlag(events[0]))
	assert.True(t, i.HasEventFlag(events[1]))
}

func TestFilterShouldFlagIfDimensionsFlagged(t *testing.T) {
	i, chain, dps, events := setupFilterTest()

	i.SetDimensions(map[string]string{"org": "mine"})
	assert.Contains(t, i.Var().String(), "mine")
	assert.NoError(t, chain.AddDatapoints(context.Background(), dps))
	assert.True(t, i.HasDatapointFlag(dps[0]))
	assert.False(t, i.HasDatapointFlag(dps[1]))

	assert.NoError(t, chain.AddEvents(context.Background(), events))
	assert.True(t, i.HasEventFlag(events[0]))
	assert.False(t, i.HasEventFlag(events[1]))
}

func TestFilterShouldFlagIfMetricFlagged(t *testing.T) {
	i, chain, dps, _ := setupFilterTest()

	i.SetDimensions(map[string]string{"sf_metric": "mname2"})
	assert.NoError(t, chain.AddDatapoints(context.Background(), dps))
	assert.False(t, i.HasDatapointFlag(dps[0]))
	assert.True(t, i.HasDatapointFlag(dps[1]))
}

func TestFilterInvalidPOSTShouldReturnError(t *testing.T) {
	i, _, _, _ := setupFilterTest()

	req, err := http.NewRequestWithContext(context.Background(), "POST", "", strings.NewReader(`_INVALID_JSON`))
	assert.NoError(t, err)
	rw := httptest.NewRecorder()
	i.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestFilterPOSTShouldChangeDimensions(t *testing.T) {
	i, _, _, _ := setupFilterTest()

	req, err := http.NewRequestWithContext(context.Background(), "POST", "", strings.NewReader(`{"name":"jack"}`))
	assert.NoError(t, err)
	rw := httptest.NewRecorder()
	i.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, map[string]string{"name": "jack"}, i.GetDimensions())
}

func TestFilterGETShouldReturnDimensions(t *testing.T) {
	i, _, _, _ := setupFilterTest()

	req, err := http.NewRequestWithContext(context.Background(), "POST", "", strings.NewReader(`{"name":"jack"}`))
	assert.NoError(t, err)
	rw := httptest.NewRecorder()
	i.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	req, err = http.NewRequestWithContext(context.Background(), "GET", "", nil)
	assert.NoError(t, err)
	rw = httptest.NewRecorder()
	i.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, `{"name":"jack"}`+"\n", rw.Body.String())
}

func TestFilterPATCHShould404(t *testing.T) {
	i, _, _, _ := setupFilterTest()

	req, err := http.NewRequestWithContext(context.Background(), "PATCH", "", strings.NewReader(`{"name":"jack"}`))
	assert.NoError(t, err)
	rw := httptest.NewRecorder()
	i.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusNotFound, rw.Code)
}

const numTests = 10

func TestCounterSink(t *testing.T) {
	dps := []*datapoint.Datapoint{
		{},
		{},
	}
	ctx := context.Background()
	bs := dptest.NewBasicSink()
	count := &dpsink.Counter{
		Logger: log.Discard,
	}
	middleSink := NextWrap(UnifyNextSinkWrap(count))(bs)
	go func() {
		// Allow time for us to get in the middle of a call
		time.Sleep(time.Millisecond)
		assert.Equal(t, int64(1), atomic.LoadInt64(&count.CallsInFlight), "After a sleep, should be in flight")
		datas := <-bs.PointsChan
		assert.Equal(t, 2, len(datas), "Original datas should be sent")
	}()
	log.IfErr(log.Panic, middleSink.AddDatapoints(ctx, dps))
	assert.Equal(t, int64(0), atomic.LoadInt64(&count.CallsInFlight), "Call is finished")
	assert.Equal(t, int64(0), atomic.LoadInt64(&count.TotalProcessErrors), "No errors so far (see above)")
	assert.Equal(t, numTests, len(count.Datapoints()), "Just checking stats len()")

	bs.RetError(errors.New("nope"))
	if err := middleSink.AddDatapoints(ctx, dps); err == nil {
		t.Fatal("Expected an error!")
	}
	assert.Equal(t, int64(1), atomic.LoadInt64(&count.TotalProcessErrors), "Error should be sent through")
}

func TestCounterSinkEvent(t *testing.T) {
	es := []*event.Event{
		{},
		{},
	}
	ctx := context.Background()
	bs := dptest.NewBasicSink()
	count := &dpsink.Counter{}
	middleSink := NextWrap(UnifyNextSinkWrap(count))(bs)
	go func() {
		// Allow time for us to get in the middle of a call
		time.Sleep(time.Millisecond)
		assert.Equal(t, int64(1), atomic.LoadInt64(&count.CallsInFlight), "After a sleep, should be in flight")
		datas := <-bs.EventsChan
		assert.Equal(t, 2, len(datas), "Original datas should be sent")
	}()
	log.IfErr(log.Panic, middleSink.AddEvents(ctx, es))
	assert.Equal(t, int64(0), atomic.LoadInt64(&count.CallsInFlight), "Call is finished")
	assert.Equal(t, int64(0), atomic.LoadInt64(&count.TotalProcessErrors), "No errors so far (see above)")
	assert.Equal(t, numTests, len(count.Datapoints()), "Just checking stats len()")

	bs.RetError(errors.New("nope"))
	if err := middleSink.AddEvents(ctx, es); err == nil {
		t.Fatal("Expected an error!")
	}
	assert.Equal(t, int64(1), atomic.LoadInt64(&count.TotalProcessErrors), "Error should be sent through")
}

func TestCounterSinkTrace(t *testing.T) {
	es := []*trace.Span{
		{},
		{},
	}
	dcount := &dpsink.Counter{}
	ctx := context.Background()
	sink := dptest.NewBasicSink()
	counter := UnifyNextSinkWrap(dcount)
	finalSink := FromChain(sink, NextWrap(counter))

	go func() {
		// Allow time for us to get in the middle of a call
		time.Sleep(time.Millisecond)
		assert.Equal(t, int64(1), atomic.LoadInt64(&dcount.CallsInFlight), "After a sleep, should be in flight")
		spans := <-sink.TracesChan
		assert.Equal(t, 2, len(spans), "Original spans should be sent")
	}()
	log.IfErr(log.Panic, finalSink.AddSpans(ctx, es))
	assert.Equal(t, int64(0), atomic.LoadInt64(&dcount.CallsInFlight), "Call is finished")
	assert.Equal(t, int64(0), atomic.LoadInt64(&dcount.TotalProcessErrors), "No errors so far (see above)")
	assert.Equal(t, numTests, len(dcount.Datapoints()), "Just checking stats len()")

	sink.RetError(errors.New("nope"))
	if err := finalSink.AddSpans(ctx, es); err == nil {
		t.Fatal("Expected an error!")
	}
	assert.Equal(t, int64(1), atomic.LoadInt64(&dcount.TotalProcessErrors), "Error should be sent through")
}
