package signalfx

import (
	"errors"
	"math"
	"testing"

	"github.com/gogo/protobuf/proto"
	sfxmodel "github.com/signalfx/com_signalfx_metrics_protobuf/model"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/pointer"
	signalfxformat "github.com/signalfx/ingest-protocols/protocol/signalfx/format"
	"github.com/stretchr/testify/assert"
)

func TestValueToValue(t *testing.T) {
	testVal := func(toSend interface{}, expected string) datapoint.Value {
		dv, err := ValueToValue(toSend)
		assert.NoError(t, err)
		assert.Equal(t, expected, dv.String())
		return dv
	}
	t.Run("test basic conversions", func(t *testing.T) {
		testVal(int64(1), "1")
		testVal(float64(.2), "0.2")
		testVal(int(3), "3")
		testVal("4", "4")
		_, err := ValueToValue(errors.New("testing"))
		assert.NotNil(t, err)
		_, err = ValueToValue(nil)
		assert.NotNil(t, err)
	})
	t.Run("show that maxfloat64 is too large to be a long", func(t *testing.T) {
		dv := testVal(math.MaxFloat64, "179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		_, ok := dv.(datapoint.FloatValue)
		assert.True(t, ok)
	})
	t.Run("show that maxint32 will be a long", func(t *testing.T) {
		dv := testVal(math.MaxInt32, "2147483647")
		_, ok := dv.(datapoint.IntValue)
		assert.True(t, ok)
	})
	t.Run("show that float(maxint64) will be a float due to edgyness of conversions", func(t *testing.T) {
		dv := testVal(float64(math.MaxInt64), "9223372036854776000")
		_, ok := dv.(datapoint.FloatValue)
		assert.True(t, ok)
	})
}

func TestNewProtobufDataPointWithType(t *testing.T) {
	dp := sfxmodel.DataPoint{}
	t.Run("nil datapoint value should error when converted", func(t *testing.T) {
		_, err := NewProtobufDataPointWithType(&dp, sfxmodel.MetricType_COUNTER)
		assert.Equal(t, errDatapointValueNotSet, err)
	})
	t.Run("source should set", func(t *testing.T) {
		dp.Value = sfxmodel.Datum{
			IntValue: pointer.Int64(1),
		}
		dp.Source = "hello"
		dp2, err := NewProtobufDataPointWithType(&dp, sfxmodel.MetricType_COUNTER)
		assert.NoError(t, err)
		assert.Equal(t, "hello", dp2.Dimensions["sf_source"])
	})
}

func TestPropertyAsRawType(t *testing.T) {
	type testCase struct {
		v   *sfxmodel.PropertyValue
		exp interface{}
	}
	cases := []testCase{
		{
			v:   nil,
			exp: nil,
		},
		{
			v: &sfxmodel.PropertyValue{
				BoolValue: proto.Bool(false),
			},
			exp: false,
		},
		{
			v: &sfxmodel.PropertyValue{
				IntValue: proto.Int64(123),
			},
			exp: 123,
		},
		{
			v: &sfxmodel.PropertyValue{
				DoubleValue: proto.Float64(2.0),
			},
			exp: 2.0,
		},
		{
			v: &sfxmodel.PropertyValue{
				StrValue: proto.String("hello"),
			},
			exp: "hello",
		},
		{
			v:   &sfxmodel.PropertyValue{},
			exp: nil,
		},
	}
	for _, c := range cases {
		assert.EqualValues(t, c.exp, PropertyAsRawType(c.v))
	}
}

func TestBodySendFormatV2(t *testing.T) {
	x := signalfxformat.BodySendFormatV2{
		Metric: "hi",
	}
	assert.Contains(t, x.String(), "hi")
}

func TestNewProtobufEvent(t *testing.T) {
	protoEvent := &sfxmodel.Event{
		EventType:  "mwp.test2",
		Dimensions: []*sfxmodel.Dimension{},
		Properties: []*sfxmodel.Property{
			{
				Key:   "version",
				Value: &sfxmodel.PropertyValue{},
			},
		},
	}
	_, err := NewProtobufEvent(protoEvent)
	assert.Equal(t, errPropertyValueNotSet, err)
}

func TestFromMT(t *testing.T) {
	assert.Panics(t, func() {
		fromMT(sfxmodel.MetricType(1001))
	})
}

func TestNewDatumValue(t *testing.T) {
	t.Run("string should convert", func(t *testing.T) {
		s1 := "abc"
		assert.Equal(t, s1, NewDatumValue(sfxmodel.Datum{StrValue: &s1}).(datapoint.StringValue).String())
	})
	t.Run("floats should convert", func(t *testing.T) {
		f1 := 1.2
		assert.Equal(t, f1, NewDatumValue(sfxmodel.Datum{DoubleValue: &f1}).(datapoint.FloatValue).Float())
	})
	t.Run("int should convert", func(t *testing.T) {
		i1 := int64(3)
		assert.Equal(t, i1, NewDatumValue(sfxmodel.Datum{IntValue: &i1}).(datapoint.IntValue).Int())
	})
}
