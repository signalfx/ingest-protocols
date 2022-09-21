// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otlp

import (
	"math"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	. "github.com/smartystreets/goconvey/convey"
	metricsservicev1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
)

const (
	unixSecs  = int64(1574092046)
	unixNSecs = int64(11 * time.Millisecond)
)

var ts = time.Unix(unixSecs, unixNSecs)

func Test_FromMetrics(t *testing.T) {
	labelMap := map[string]string{
		"k0": "v0",
		"k1": "v1",
	}

	const doubleVal = 1234.5678
	makeDoublePt := func() *metricsv1.NumberDataPoint {
		return &metricsv1.NumberDataPoint{
			TimeUnixNano: uint64(ts.UnixNano()),
			Value:        &metricsv1.NumberDataPoint_AsDouble{AsDouble: doubleVal},
		}
	}

	makeDoublePtWithLabels := func() *metricsv1.NumberDataPoint {
		pt := makeDoublePt()
		pt.Attributes = stringMapToAttributeMap(labelMap)
		return pt
	}

	const int64Val = int64(123)
	makeInt64Pt := func() *metricsv1.NumberDataPoint {
		return &metricsv1.NumberDataPoint{
			TimeUnixNano: uint64(ts.UnixNano()),
			Value:        &metricsv1.NumberDataPoint_AsInt{AsInt: int64Val},
		}
	}

	makeInt64PtWithLabels := func() *metricsv1.NumberDataPoint {
		pt := makeInt64Pt()
		pt.Attributes = stringMapToAttributeMap(labelMap)
		return pt
	}

	makeNilValuePt := func() *metricsv1.NumberDataPoint {
		return &metricsv1.NumberDataPoint{
			TimeUnixNano: uint64(ts.UnixNano()),
			Value:        nil,
		}
	}

	histBounds := []float64{1, 2, 4}
	histCounts := []uint64{4, 2, 3, 7}

	makeDoubleHistDP := func() *metricsv1.HistogramDataPoint {
		return &metricsv1.HistogramDataPoint{
			TimeUnixNano:   uint64(ts.UnixNano()),
			Count:          16,
			Sum:            float64Pointer(100.0),
			ExplicitBounds: histBounds,
			BucketCounts:   histCounts,
			Attributes:     stringMapToAttributeMap(labelMap),
		}
	}
	doubleHistDP := makeDoubleHistDP()

	makeDoubleHistDPBadCounts := func() *metricsv1.HistogramDataPoint {
		return &metricsv1.HistogramDataPoint{
			TimeUnixNano:   uint64(ts.UnixNano()),
			Count:          16,
			Sum:            float64Pointer(100.0),
			ExplicitBounds: histBounds,
			BucketCounts:   []uint64{4},
			Attributes:     stringMapToAttributeMap(labelMap),
		}
	}

	makeIntHistDP := func() *metricsv1.HistogramDataPoint {
		return &metricsv1.HistogramDataPoint{
			TimeUnixNano:   uint64(ts.UnixNano()),
			Count:          16,
			Sum:            float64Pointer(100),
			ExplicitBounds: histBounds,
			BucketCounts:   histCounts,
			Attributes:     stringMapToAttributeMap(labelMap),
		}
	}
	intHistDP := makeIntHistDP()

	makeIntHistDPBadCounts := func() *metricsv1.HistogramDataPoint {
		return &metricsv1.HistogramDataPoint{
			TimeUnixNano:   uint64(ts.UnixNano()),
			Count:          16,
			Sum:            float64Pointer(100),
			ExplicitBounds: histBounds,
			BucketCounts:   []uint64{4},
			Attributes:     stringMapToAttributeMap(labelMap),
		}
	}

	makeHistDPNoBuckets := func() *metricsv1.HistogramDataPoint {
		return &metricsv1.HistogramDataPoint{
			Count:        2,
			Sum:          float64Pointer(10),
			TimeUnixNano: uint64(ts.UnixNano()),
			Attributes:   stringMapToAttributeMap(labelMap),
		}
	}
	histDPNoBuckets := makeHistDPNoBuckets()

	const summarySumVal = 123.4
	const summaryCountVal = 111

	makeSummaryDP := func() *metricsv1.SummaryDataPoint {
		summaryDP := &metricsv1.SummaryDataPoint{
			TimeUnixNano: uint64(ts.UnixNano()),
			Sum:          summarySumVal,
			Count:        summaryCountVal,
			Attributes:   stringMapToAttributeMap(labelMap),
		}
		for i := 0; i < 4; i++ {
			summaryDP.QuantileValues = append(summaryDP.QuantileValues, &metricsv1.SummaryDataPoint_ValueAtQuantile{
				Quantile: 0.25 * float64(i+1),
				Value:    float64(i),
			})
		}
		return summaryDP
	}

	makeEmptySummaryDP := func() *metricsv1.SummaryDataPoint {
		return &metricsv1.SummaryDataPoint{
			TimeUnixNano: uint64(ts.UnixNano()),
			Sum:          summarySumVal,
			Count:        summaryCountVal,
			Attributes:   stringMapToAttributeMap(labelMap),
		}
	}

	createRMS := func() (*metricsv1.ResourceMetrics, *[]*metricsv1.Metric) {
		out := &metricsv1.ResourceMetrics{}
		var metrics *[]*metricsv1.Metric
		ilm := &metricsv1.ScopeMetrics{}
		out.ScopeMetrics = append(out.ScopeMetrics, ilm)
		metrics = &ilm.Metrics

		return out, metrics
	}
	tests := []struct {
		name              string
		metricsFn         func() []*metricsv1.ResourceMetrics
		wantSfxDataPoints []*datapoint.Datapoint
	}{
		{
			name: "nil_node_nil_resources_no_dims",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()

				*metrics = []*metricsv1.Metric{
					{
						Name: "gauge_double_with_no_dims",
						Data: &metricsv1.Metric_Gauge{
							Gauge: &metricsv1.Gauge{
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePt(),
								},
							},
						},
					},
					{
						Name: "gauge_int_with_no_dims",
						Data: &metricsv1.Metric_Gauge{
							Gauge: &metricsv1.Gauge{
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64Pt(),
								},
							},
						},
					},
					{
						Name: "cumulative_double_with_no_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic:            true,
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePt(),
								},
							},
						},
					},
					{
						Name: "cumulative_int_with_no_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic:            true,
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64Pt(),
								},
							},
						},
					},
					{
						Name: "delta_double_with_no_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic:            true,
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePt(),
								},
							},
						},
					},
					{
						Name: "delta_int_with_no_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic:            true,
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64Pt(),
								},
							},
						},
					},
					{
						Name: "gauge_sum_double_with_no_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic: false,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePt(),
								},
							},
						},
					},
					{
						Name: "gauge_sum_int_with_no_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic: false,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64Pt(),
								},
							},
						},
					},
					{
						Name: "gauge_sum_int_with_nil_value",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic: false,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeNilValuePt(),
								},
							},
						},
					},
					{
						Name: "nil_data",
						Data: nil,
					},
				}

				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: []*datapoint.Datapoint{
				doubleSFxDataPoint("gauge_double_with_no_dims", datapoint.Gauge, nil, doubleVal),
				int64SFxDataPoint("gauge_int_with_no_dims", datapoint.Gauge, nil, int64Val),
				doubleSFxDataPoint("cumulative_double_with_no_dims", datapoint.Counter, nil, doubleVal),
				int64SFxDataPoint("cumulative_int_with_no_dims", datapoint.Counter, nil, int64Val),
				doubleSFxDataPoint("delta_double_with_no_dims", datapoint.Count, nil, doubleVal),
				int64SFxDataPoint("delta_int_with_no_dims", datapoint.Count, nil, int64Val),
				doubleSFxDataPoint("gauge_sum_double_with_no_dims", datapoint.Gauge, nil, doubleVal),
				int64SFxDataPoint("gauge_sum_int_with_no_dims", datapoint.Gauge, nil, int64Val),
				{
					Metric:     "gauge_sum_int_with_nil_value",
					Timestamp:  ts,
					Value:      nil,
					MetricType: datapoint.Gauge,
					Dimensions: map[string]string{},
				},
			},
		},
		{
			name: "nil_node_and_resources_with_dims",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()

				*metrics = []*metricsv1.Metric{
					{
						Name: "gauge_double_with_dims",
						Data: &metricsv1.Metric_Gauge{
							Gauge: &metricsv1.Gauge{
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePtWithLabels(),
								},
							},
						},
					},
					{
						Name: "gauge_int_with_dims",
						Data: &metricsv1.Metric_Gauge{
							Gauge: &metricsv1.Gauge{
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64PtWithLabels(),
								},
							},
						},
					},
					{
						Name: "cumulative_double_with_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic:            true,
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePtWithLabels(),
								},
							},
						},
					},
					{
						Name: "cumulative_int_with_dims",
						Data: &metricsv1.Metric_Sum{
							Sum: &metricsv1.Sum{
								IsMonotonic:            true,
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64PtWithLabels(),
								},
							},
						},
					},
				}

				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: []*datapoint.Datapoint{
				doubleSFxDataPoint("gauge_double_with_dims", datapoint.Gauge, labelMap, doubleVal),
				int64SFxDataPoint("gauge_int_with_dims", datapoint.Gauge, labelMap, int64Val),
				doubleSFxDataPoint("cumulative_double_with_dims", datapoint.Counter, labelMap, doubleVal),
				int64SFxDataPoint("cumulative_int_with_dims", datapoint.Counter, labelMap, int64Val),
			},
		},
		{
			name: "with_node_resources_dims",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()
				out.Resource = &resourcev1.Resource{
					Attributes: []*commonv1.KeyValue{
						{
							Key:   "k_r0",
							Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "v_r0"}},
						},
						{
							Key:   "k_r1",
							Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "v_r1"}},
						},
						{
							Key:   "k_n0",
							Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "v_n0"}},
						},
						{
							Key:   "k_n1",
							Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "v_n1"}},
						},
					},
				}

				*metrics = []*metricsv1.Metric{
					{
						Name: "gauge_double_with_dims",
						Data: &metricsv1.Metric_Gauge{
							Gauge: &metricsv1.Gauge{
								DataPoints: []*metricsv1.NumberDataPoint{
									makeDoublePtWithLabels(),
								},
							},
						},
					},
					{
						Name: "gauge_int_with_dims",
						Data: &metricsv1.Metric_Gauge{
							Gauge: &metricsv1.Gauge{
								DataPoints: []*metricsv1.NumberDataPoint{
									makeInt64PtWithLabels(),
								},
							},
						},
					},
				}
				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: []*datapoint.Datapoint{
				doubleSFxDataPoint(
					"gauge_double_with_dims",
					datapoint.Gauge,
					mergeStringMaps(map[string]string{
						"k_n0": "v_n0",
						"k_n1": "v_n1",
						"k_r0": "v_r0",
						"k_r1": "v_r1",
					}, labelMap),
					doubleVal),
				int64SFxDataPoint(
					"gauge_int_with_dims",
					datapoint.Gauge,
					mergeStringMaps(map[string]string{
						"k_n0": "v_n0",
						"k_n1": "v_n1",
						"k_r0": "v_r0",
						"k_r1": "v_r1",
					}, labelMap),
					int64Val),
			},
		},
		{
			name: "histograms",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()

				*metrics = []*metricsv1.Metric{
					{
						Name: "int_histo",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeIntHistDP(),
								},
							},
						},
					},
					{
						Name: "int_delta_histo",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeIntHistDP(),
								},
							},
						},
					},
					{
						Name: "double_histo",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeDoubleHistDP(),
								},
							},
						},
					},
					{
						Name: "double_delta_histo",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeDoubleHistDP(),
								},
							},
						},
					},
					{
						Name: "double_histo_bad_counts",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeDoubleHistDPBadCounts(),
								},
							},
						},
					},
					{
						Name: "int_histo_bad_counts",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeIntHistDPBadCounts(),
								},
							},
						},
					},
				}
				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: mergeDPs(
				expectedFromHistogram("int_histo", labelMap, intHistDP, false),
				expectedFromHistogram("int_delta_histo", labelMap, intHistDP, true),
				expectedFromHistogram("double_histo", labelMap, doubleHistDP, false),
				expectedFromHistogram("double_delta_histo", labelMap, doubleHistDP, true),
				[]*datapoint.Datapoint{
					int64SFxDataPoint("double_histo_bad_counts_count", datapoint.Counter, labelMap, int64(doubleHistDP.Count)),
					doubleSFxDataPoint("double_histo_bad_counts_sum", datapoint.Counter, labelMap, *doubleHistDP.Sum),
				},
				[]*datapoint.Datapoint{
					int64SFxDataPoint("int_histo_bad_counts_count", datapoint.Counter, labelMap, int64(intHistDP.Count)),
					doubleSFxDataPoint("int_histo_bad_counts_sum", datapoint.Counter, labelMap, *intHistDP.Sum),
				},
			),
		},
		{
			name: "distribution_no_buckets",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()

				*metrics = []*metricsv1.Metric{
					{
						Name: "no_bucket_histo",
						Data: &metricsv1.Metric_Histogram{
							Histogram: &metricsv1.Histogram{
								DataPoints: []*metricsv1.HistogramDataPoint{
									makeHistDPNoBuckets(),
								},
							},
						},
					},
				}
				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: expectedFromHistogram("no_bucket_histo", labelMap, histDPNoBuckets, false),
		},
		{
			name: "summaries",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()

				*metrics = []*metricsv1.Metric{
					{
						Name: "summary",
						Data: &metricsv1.Metric_Summary{
							Summary: &metricsv1.Summary{
								DataPoints: []*metricsv1.SummaryDataPoint{
									makeSummaryDP(),
								},
							},
						},
					},
				}
				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: expectedFromSummary("summary", labelMap, summaryCountVal, summarySumVal),
		},
		{
			name: "empty_summary",
			metricsFn: func() []*metricsv1.ResourceMetrics {
				out, metrics := createRMS()

				*metrics = []*metricsv1.Metric{
					{
						Name: "empty_summary",
						Data: &metricsv1.Metric_Summary{
							Summary: &metricsv1.Summary{
								DataPoints: []*metricsv1.SummaryDataPoint{
									makeEmptySummaryDP(),
								},
							},
						},
					},
				}
				return []*metricsv1.ResourceMetrics{out}
			},
			wantSfxDataPoints: expectedFromEmptySummary("empty_summary", labelMap, summaryCountVal, summarySumVal),
		},
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			rms := tt.metricsFn()
			gotSfxDataPoints := FromOTLPMetricRequest(&metricsservicev1.ExportMetricsServiceRequest{ResourceMetrics: rms})
			So(tt.wantSfxDataPoints, ShouldResemble, gotSfxDataPoints)

			firstMetric := rms[0].GetScopeMetrics()[0].Metrics[0]
			dpsFromMetric := FromMetric(firstMetric)
			So(dpsFromMetric, ShouldNotBeEmpty)
		})
	}
}

func TestMetricTypeDerive(t *testing.T) {
	Convey("deriveSignalFxMetricType panics if invalid metric", t, func() {
		So(func() { deriveSignalFxMetricType(&metricsv1.Metric{}) }, ShouldPanic)
	})
}

func TestAttributesToDimensions(t *testing.T) {
	Convey("stringifyAttributes", t, func() {
		attrs := []*commonv1.KeyValue{
			{
				Key:   "a",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "s"}},
			},
			{
				Key:   "b",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: ""}},
			},
			{
				Key:   "c",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: true}},
			},
			{
				Key:   "d",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: 44}},
			},
			{
				Key:   "e",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_DoubleValue{DoubleValue: 45.1}},
			},
			{
				Key: "f",
				Value: &commonv1.AnyValue{
					Value: &commonv1.AnyValue_ArrayValue{
						ArrayValue: &commonv1.ArrayValue{
							Values: []*commonv1.AnyValue{
								{Value: &commonv1.AnyValue_StringValue{StringValue: "n1"}},
								{Value: &commonv1.AnyValue_StringValue{StringValue: "n2"}},
							},
						},
					},
				},
			},
			{
				Key: "g",
				Value: &commonv1.AnyValue{
					Value: &commonv1.AnyValue_KvlistValue{
						KvlistValue: &commonv1.KeyValueList{
							Values: []*commonv1.KeyValue{
								{Key: "k1", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "n1"}}},
								{Key: "k2", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: false}}},
								{Key: "k3", Value: nil},
								{Key: "k4", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_DoubleValue{DoubleValue: 40.3}}},
								{Key: "k5", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: 41}}},
								{Key: "k6", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_BytesValue{BytesValue: []byte("n2")}}},
							},
						},
					},
				},
			},
			{
				Key:   "h",
				Value: nil,
			},
			{
				Key:   "i",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_DoubleValue{DoubleValue: 0}},
			},
			{
				Key:   "j",
				Value: &commonv1.AnyValue{Value: nil},
			},
			{
				Key:   "k",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_BytesValue{BytesValue: []byte("to boldly go")}},
			},
		}

		dimKVs := stringifyAttributes(attrs)
		var m SignalFxMetric
		m.DP.Attributes = dimKVs
		So(m.ToDatapoint().Dimensions, ShouldResemble, map[string]string{
			"a": "s",
			"c": "true",
			"d": "44",
			"e": "45.1",
			"f": `["n1","n2"]`,
			"g": `{"k1":"n1","k2":false,"k3":null,"k4":40.3,"k5":41,"k6":"bjI="}`,
			"i": "0",
			// No entry for "j" because it's nil and would be skipped by ToDatapoint()
			"k": "dG8gYm9sZGx5IGdv",
		})
	})

	Convey("Non-string attributes in SignalFxMetric panic", t, func() {
		attrs := []*commonv1.KeyValue{
			{
				Key:   "a",
				Value: nil,
			},
			{
				Key:   "c",
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: true}},
			},
		}
		var m SignalFxMetric
		m.DP.Attributes = attrs
		So(func() { m.ToDatapoint() }, ShouldPanic)
	})
}

func doubleSFxDataPoint(
	metric string,
	metricType datapoint.MetricType,
	dims map[string]string,
	val float64,
) *datapoint.Datapoint {
	return &datapoint.Datapoint{
		Metric:     metric,
		Timestamp:  ts,
		Value:      datapoint.NewFloatValue(val),
		MetricType: metricType,
		Dimensions: cloneStringMap(dims),
	}
}

func int64SFxDataPoint(
	metric string,
	metricType datapoint.MetricType,
	dims map[string]string,
	val int64,
) *datapoint.Datapoint {
	return &datapoint.Datapoint{
		Metric:     metric,
		Timestamp:  ts,
		Value:      datapoint.NewIntValue(val),
		MetricType: metricType,
		Dimensions: cloneStringMap(dims),
	}
}

func expectedFromHistogram(
	metricName string,
	dims map[string]string,
	histDP *metricsv1.HistogramDataPoint,
	isDelta bool,
) []*datapoint.Datapoint {
	buckets := histDP.GetBucketCounts()

	dps := make([]*datapoint.Datapoint, 0)

	typ := datapoint.Counter
	if isDelta {
		typ = datapoint.Count
	}

	dps = append(dps,
		int64SFxDataPoint(metricName+"_count", typ, dims, int64(histDP.GetCount())),
		doubleSFxDataPoint(metricName+"_sum", typ, dims, histDP.GetSum()))

	explicitBounds := histDP.GetExplicitBounds()
	if explicitBounds == nil {
		return dps
	}
	var le int64
	for i := 0; i < len(explicitBounds); i++ {
		dimsCopy := cloneStringMap(dims)
		dimsCopy[upperBoundDimensionKey] = float64ToDimValue(explicitBounds[i])
		le += int64(buckets[i])
		dps = append(dps, int64SFxDataPoint(metricName+"_bucket", typ, dimsCopy, le))
	}
	dimsCopy := cloneStringMap(dims)
	dimsCopy[upperBoundDimensionKey] = float64ToDimValue(math.Inf(1))
	le += int64(buckets[len(buckets)-1])
	dps = append(dps, int64SFxDataPoint(metricName+"_bucket", typ, dimsCopy, le))
	return dps
}

func expectedFromSummary(name string, labelMap map[string]string, count int64, sumVal float64) []*datapoint.Datapoint {
	countPt := int64SFxDataPoint(name+"_count", datapoint.Counter, labelMap, count)
	sumPt := doubleSFxDataPoint(name+"_sum", datapoint.Counter, labelMap, sumVal)
	out := []*datapoint.Datapoint{countPt, sumPt}
	quantileDimVals := []string{"0.25", "0.5", "0.75", "1"}
	for i := 0; i < 4; i++ {
		qDims := map[string]string{"quantile": quantileDimVals[i]}
		qPt := doubleSFxDataPoint(
			name+"_quantile",
			datapoint.Gauge,
			mergeStringMaps(labelMap, qDims),
			float64(i),
		)
		out = append(out, qPt)
	}
	return out
}

func expectedFromEmptySummary(name string, labelMap map[string]string, count int64, sumVal float64) []*datapoint.Datapoint {
	countPt := int64SFxDataPoint(name+"_count", datapoint.Counter, labelMap, count)
	sumPt := doubleSFxDataPoint(name+"_sum", datapoint.Counter, labelMap, sumVal)
	return []*datapoint.Datapoint{countPt, sumPt}
}

func mergeDPs(dps ...[]*datapoint.Datapoint) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint
	for i := range dps {
		out = append(out, dps[i]...)
	}
	return out
}

func cloneStringMap(m map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[k] = v
	}
	return out
}

func stringMapToAttributeMap(m map[string]string) []*commonv1.KeyValue {
	ret := make([]*commonv1.KeyValue, 0, len(m))
	for k, v := range m {
		ret = append(ret, &commonv1.KeyValue{
			Key:   k,
			Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: v}},
		})
	}
	return ret
}

func float64Pointer(f float64) *float64 { return &f }
