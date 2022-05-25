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
	. "github.com/smartystreets/goconvey/convey"
)

func TestValueToValue(t *testing.T) {
	testVal := func(toSend interface{}, expected string) datapoint.Value {
		dv, err := ValueToValue(toSend)
		So(err, ShouldBeNil)
		So(expected, ShouldEqual, dv.String())
		return dv
	}
	Convey("v2v conversion", t, func() {
		Convey("test basic conversions", func() {
			testVal(int64(1), "1")
			testVal(float64(.2), "0.2")
			testVal(int(3), "3")
			testVal("4", "4")
			_, err := ValueToValue(errors.New("testing"))
			So(err, ShouldNotBeNil)
			_, err = ValueToValue(nil)
			So(err, ShouldNotBeNil)
		})
		Convey("show that maxfloat64 is too large to be a long", func() {
			dv := testVal(math.MaxFloat64, "179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			_, ok := dv.(datapoint.FloatValue)
			So(ok, ShouldBeTrue)
		})
		Convey("show that maxint32 will be a long", func() {
			dv := testVal(math.MaxInt32, "2147483647")
			_, ok := dv.(datapoint.IntValue)
			So(ok, ShouldBeTrue)
		})
		Convey("show that float(maxint64) will be a float due to edgyness of conversions", func() {
			dv := testVal(float64(math.MaxInt64), "9223372036854776000")
			_, ok := dv.(datapoint.FloatValue)
			So(ok, ShouldBeTrue)
		})
	})
}

func TestNewProtobufDataPointWithType(t *testing.T) {
	Convey("A nil datapoint value", t, func() {
		dp := sfxmodel.DataPoint{}
		Convey("should error when converted", func() {
			_, err := NewProtobufDataPointWithType(&dp, sfxmodel.MetricType_COUNTER)
			So(err, ShouldEqual, errDatapointValueNotSet)
		})
		Convey("with a value", func() {
			dp.Value = sfxmodel.Datum{
				IntValue: pointer.Int64(1),
			}
			Convey("source should set", func() {
				dp.Source = "hello"
				dp2, err := NewProtobufDataPointWithType(&dp, sfxmodel.MetricType_COUNTER)
				So(err, ShouldBeNil)
				So(dp2.Dimensions["sf_source"], ShouldEqual, "hello")
			})
		})
	})
}

func TestPropertyAsRawType(t *testing.T) {
	Convey("With value raw test PropertyAsRawType", t, func() {
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
			So(PropertyAsRawType(c.v), ShouldEqual, c.exp)
		}
	})
}

func TestBodySendFormatV2(t *testing.T) {
	Convey("BodySendFormatV2 should String()-ify", t, func() {
		x := signalfxformat.BodySendFormatV2{
			Metric: "hi",
		}
		So(x.String(), ShouldContainSubstring, "hi")
	})
}

func TestNewProtobufEvent(t *testing.T) {
	Convey("given a protobuf event with a nil property value", t, func() {
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
		Convey("should error when converted", func() {
			_, err := NewProtobufEvent(protoEvent)
			So(err, ShouldEqual, errPropertyValueNotSet)
		})
	})
}

func TestFromMT(t *testing.T) {
	Convey("invalid fromMT types should panic", t, func() {
		So(func() {
			fromMT(sfxmodel.MetricType(1001))
		}, ShouldPanic)
	})
}

func TestNewDatumValue(t *testing.T) {
	Convey("datum values should convert", t, func() {
		Convey("string should convert", func() {
			s1 := "abc"
			So(s1, ShouldEqual, NewDatumValue(sfxmodel.Datum{StrValue: &s1}).(datapoint.StringValue).String())
		})
		Convey("floats should convert", func() {
			f1 := 1.2
			So(f1, ShouldEqual, NewDatumValue(sfxmodel.Datum{DoubleValue: &f1}).(datapoint.FloatValue).Float())
		})
		Convey("int should convert", func() {
			i1 := int64(3)
			So(i1, ShouldEqual, NewDatumValue(sfxmodel.Datum{IntValue: &i1}).(datapoint.IntValue).Int())
		})
	})
}
