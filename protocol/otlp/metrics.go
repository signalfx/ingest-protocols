// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This code is copied and modified directly from the OTEL Collector:
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/1d6309bb62264cc7e2dda076ed95385b1ddef28a/pkg/translator/signalfx/from_metrics.go

package otlp

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	metricsservicev1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
)

var (
	// Some standard dimension keys.
	// upper bound dimension key for histogram buckets.
	upperBoundDimensionKey = "le"

	// infinity bound dimension value is used on all histograms.
	infinityBoundSFxDimValue = float64ToDimValue(math.Inf(1))
)

// SignalFxMetric is a single NumberDataPoint paired with a metric name such that it contains all of
// the information needed to convert it to a SignalFx datapoint.  It serves as an intermediate
// object between an OTLP DataPoint and the SignalFx datapoint.Datapoint type.  Atttribute values
// must be made into strings and attributes from the resource should be added to the attributes of
// the DP.
type SignalFxMetric struct {
	Name string
	Type datapoint.MetricType
	DP   metricsv1.NumberDataPoint
}

// ToDatapoint converts the SignalFxMetric to a datapoint.Datapoint instance
func (s *SignalFxMetric) ToDatapoint() *datapoint.Datapoint {
	return &datapoint.Datapoint{
		Metric:     s.Name,
		MetricType: s.Type,
		Timestamp:  time.Unix(0, int64(s.DP.GetTimeUnixNano())),
		Dimensions: StringAttributesToDimensions(s.DP.GetAttributes()),
		Value:      numberToSignalFxValue(&s.DP),
	}
}

// StringAttributesToDimensions converts a list of string KVs into a map.
func StringAttributesToDimensions(attributes []*commonv1.KeyValue) map[string]string {
	dimensions := make(map[string]string, len(attributes))
	if len(attributes) == 0 {
		return dimensions
	}
	for _, kv := range attributes {
		if kv.GetValue().GetValue() == nil {
			continue
		}
		if v, ok := kv.GetValue().GetValue().(*commonv1.AnyValue_StringValue); ok {
			if v.StringValue == "" {
				// Don't bother setting things that serialize to nothing
				continue
			}

			dimensions[kv.Key] = v.StringValue
		} else {
			panic("attributes must be converted to string before using in SignalFxMetric")
		}
	}
	return dimensions
}

// FromOTLPMetricRequest converts the ResourceMetrics in an incoming request to SignalFx datapoints
func FromOTLPMetricRequest(md *metricsservicev1.ExportMetricsServiceRequest) []*datapoint.Datapoint {
	return FromOTLPResourceMetrics(md.GetResourceMetrics())
}

// FromOTLPResourceMetrics converts OTLP ResourceMetrics to SignalFx datapoints.
func FromOTLPResourceMetrics(rms []*metricsv1.ResourceMetrics) []*datapoint.Datapoint {
	return datapointsFromMetrics(SignalFxMetricsFromOTLPResourceMetrics(rms))
}

// FromMetric converts a OTLP Metric to SignalFx datapoint(s).
func FromMetric(m *metricsv1.Metric) []*datapoint.Datapoint {
	return datapointsFromMetrics(SignalFxMetricsFromOTLPMetric(m))
}

func datapointsFromMetrics(sfxMetrics []SignalFxMetric) []*datapoint.Datapoint {
	sfxDps := make([]*datapoint.Datapoint, len(sfxMetrics))
	for i := range sfxMetrics {
		sfxDps[i] = sfxMetrics[i].ToDatapoint()
	}
	return sfxDps
}

// SignalFxMetricsFromOTLPResourceMetrics creates the intermediate SignalFxMetric from OTLP metrics
// instead of going all the way to datapoint.Datapoint.
func SignalFxMetricsFromOTLPResourceMetrics(rms []*metricsv1.ResourceMetrics) []SignalFxMetric {
	var sfxDps []SignalFxMetric

	for _, rm := range rms {
		for _, ilm := range rm.GetScopeMetrics() {
			for _, m := range ilm.GetMetrics() {
				sfxDps = append(sfxDps, SignalFxMetricsFromOTLPMetric(m)...)
			}
		}

		resourceAttrs := stringifyAttributes(rm.GetResource().GetAttributes())
		for i := range sfxDps {
			sfxDps[i].DP.Attributes = append(sfxDps[i].DP.Attributes, resourceAttrs...)
		}
	}

	return sfxDps
}

// SignalFxMetricsFromOTLPMetric converts an OTLP Metric to a SignalFxMetric
func SignalFxMetricsFromOTLPMetric(m *metricsv1.Metric) []SignalFxMetric {
	var sfxMetrics []SignalFxMetric

	data := m.GetData()
	switch data.(type) {
	case *metricsv1.Metric_Gauge:
		sfxMetrics = convertNumberDataPoints(m.GetGauge().GetDataPoints(), deriveSignalFxMetricType(m), m.GetName())
	case *metricsv1.Metric_Sum:
		sfxMetrics = convertNumberDataPoints(m.GetSum().GetDataPoints(), deriveSignalFxMetricType(m), m.GetName())
	case *metricsv1.Metric_Histogram:
		sfxMetrics = convertHistogram(m.GetHistogram().GetDataPoints(), deriveSignalFxMetricType(m), m.GetName())
	case *metricsv1.Metric_ExponentialHistogram:
		// TODO: Add support for these
	case *metricsv1.Metric_Summary:
		sfxMetrics = convertSummaryDataPoints(m.GetSummary().GetDataPoints(), m.GetName())
	}

	return sfxMetrics
}

func deriveSignalFxMetricType(m *metricsv1.Metric) datapoint.MetricType {
	switch m.GetData().(type) {
	case *metricsv1.Metric_Gauge:
		return datapoint.Gauge

	case *metricsv1.Metric_Sum:
		if !m.GetSum().GetIsMonotonic() {
			return datapoint.Gauge
		}
		if m.GetSum().GetAggregationTemporality() == metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA {
			return datapoint.Count
		}
		return datapoint.Counter

	case *metricsv1.Metric_Histogram:
		if m.GetHistogram().GetAggregationTemporality() == metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA {
			return datapoint.Count
		}
		return datapoint.Counter
	}
	panic("invalid metric type")
}

func convertNumberDataPoints(dps []*metricsv1.NumberDataPoint, typ datapoint.MetricType, name string) []SignalFxMetric {
	out := make([]SignalFxMetric, len(dps))
	for i, dp := range dps {
		out[i].Name = name
		out[i].Type = typ
		out[i].DP = metricsv1.NumberDataPoint{
			Attributes:        stringifyAttributes(dp.Attributes),
			StartTimeUnixNano: dp.StartTimeUnixNano,
			TimeUnixNano:      dp.TimeUnixNano,
			Value:             dp.Value,
		}
	}
	return out
}

func convertHistogram(histDPs []*metricsv1.HistogramDataPoint, typ datapoint.MetricType, name string) []SignalFxMetric {
	biggestCount := 0
	for _, histDP := range histDPs {
		c := len(histDP.GetBucketCounts())
		if c > biggestCount {
			biggestCount = c
		}
	}

	// ensure we are big enough to fit everything
	out := make([]SignalFxMetric, len(histDPs)*(2+biggestCount))

	i := 0
	for _, histDP := range histDPs {
		stringAttrs := stringifyAttributes(histDP.GetAttributes())

		countPt := &out[i]
		countPt.Name = name + "_count"
		countPt.Type = typ
		c := int64(histDP.GetCount())
		countPt.DP.Attributes = stringAttrs
		countPt.DP.TimeUnixNano = histDP.GetTimeUnixNano()
		countPt.DP.Value = &metricsv1.NumberDataPoint_AsInt{AsInt: c}
		i++

		sumPt := &out[i]
		sumPt.Name = name + "_sum"
		sumPt.Type = typ
		sum := histDP.GetSum()
		sumPt.DP.Attributes = stringAttrs
		sumPt.DP.TimeUnixNano = histDP.GetTimeUnixNano()
		sumPt.DP.Value = &metricsv1.NumberDataPoint_AsDouble{AsDouble: sum}
		i++

		bounds := histDP.GetExplicitBounds()
		counts := histDP.GetBucketCounts()

		// Spec says counts is optional but if present it must have one more
		// element than the bounds array.
		if len(counts) > 0 && len(counts) != len(bounds)+1 {
			continue
		}

		var le int64
		for j, c := range counts {
			bound := infinityBoundSFxDimValue
			if j < len(bounds) {
				bound = float64ToDimValue(bounds[j])
			}

			dp := &out[i]
			dp.Name = name + "_bucket"
			dp.Type = typ
			dp.DP.Attributes = append(stringAttrs, &commonv1.KeyValue{
				Key: upperBoundDimensionKey,
				Value: &commonv1.AnyValue{
					Value: &commonv1.AnyValue_StringValue{
						StringValue: bound,
					},
				},
			})
			dp.DP.TimeUnixNano = histDP.GetTimeUnixNano()
			le += int64(c)
			dp.DP.Value = &metricsv1.NumberDataPoint_AsInt{AsInt: le}
			i++
		}
	}

	return out[:i]
}

func convertSummaryDataPoints(
	in []*metricsv1.SummaryDataPoint,
	name string,
) []SignalFxMetric {
	biggestCount := 0
	for _, sumDP := range in {
		c := len(sumDP.GetQuantileValues())
		if c > biggestCount {
			biggestCount = c
		}
	}
	out := make([]SignalFxMetric, len(in)*(2+biggestCount))

	i := 0
	for _, inDp := range in {
		stringAttrs := stringifyAttributes(inDp.GetAttributes())

		countPt := &out[i]
		countPt.Name = name + "_count"
		countPt.Type = datapoint.Counter
		c := int64(inDp.GetCount())
		countPt.DP.Attributes = stringAttrs
		countPt.DP.TimeUnixNano = inDp.GetTimeUnixNano()
		countPt.DP.Value = &metricsv1.NumberDataPoint_AsInt{AsInt: c}
		i++

		sumPt := &out[i]
		sumPt.Name = name + "_sum"
		sumPt.Type = datapoint.Counter
		sumPt.DP.Attributes = stringAttrs
		sumPt.DP.TimeUnixNano = inDp.GetTimeUnixNano()
		sumPt.DP.Value = &metricsv1.NumberDataPoint_AsDouble{AsDouble: inDp.GetSum()}
		i++

		qvs := inDp.GetQuantileValues()
		for _, qv := range qvs {
			qPt := &out[i]
			qPt.Name = name + "_quantile"
			qPt.Type = datapoint.Gauge
			qPt.DP.Attributes = append(stringAttrs, &commonv1.KeyValue{
				Key: "quantile",
				Value: &commonv1.AnyValue{
					Value: &commonv1.AnyValue_StringValue{
						StringValue: strconv.FormatFloat(qv.GetQuantile(), 'f', -1, 64),
					},
				},
			})
			qPt.DP.TimeUnixNano = inDp.GetTimeUnixNano()
			qPt.DP.Value = &metricsv1.NumberDataPoint_AsDouble{AsDouble: qv.GetValue()}
			i++
		}
	}
	return out[:i]
}

func stringifyAttributes(attributes []*commonv1.KeyValue) []*commonv1.KeyValue {
	out := make([]*commonv1.KeyValue, len(attributes))
	for i, kv := range attributes {
		out[i] = &commonv1.KeyValue{
			Key: attributes[i].Key,
			Value: &commonv1.AnyValue{
				Value: &commonv1.AnyValue_StringValue{
					StringValue: StringifyAnyValue(kv.GetValue()),
				},
			},
		}
	}
	return out
}

// StringifyAnyValue converts an AnyValue to a string.  KVLists and Arrays get recursively JSON
// marshalled.
func StringifyAnyValue(a *commonv1.AnyValue) string {
	var v string
	if a == nil {
		return ""
	}
	switch a.GetValue().(type) {
	case *commonv1.AnyValue_StringValue:
		v = a.GetStringValue()

	case *commonv1.AnyValue_BytesValue:
		v = base64.StdEncoding.EncodeToString(a.GetBytesValue())

	case *commonv1.AnyValue_BoolValue:
		v = strconv.FormatBool(a.GetBoolValue())

	case *commonv1.AnyValue_DoubleValue:
		v = float64ToDimValue(a.GetDoubleValue())

	case *commonv1.AnyValue_IntValue:
		v = strconv.FormatInt(a.GetIntValue(), 10)

	case *commonv1.AnyValue_KvlistValue, *commonv1.AnyValue_ArrayValue:
		jsonStr, _ := json.Marshal(anyValueToRaw(a))
		v = string(jsonStr)
	}

	return v
}

// nolint:gocyclo
func anyValueToRaw(a *commonv1.AnyValue) interface{} {
	var v interface{}
	if a == nil {
		return nil
	}
	switch a.GetValue().(type) {
	case *commonv1.AnyValue_StringValue:
		v = a.GetStringValue()

	case *commonv1.AnyValue_BytesValue:
		v = a.GetBytesValue()

	case *commonv1.AnyValue_BoolValue:
		v = a.GetBoolValue()

	case *commonv1.AnyValue_DoubleValue:
		v = a.GetDoubleValue()

	case *commonv1.AnyValue_IntValue:
		v = a.GetIntValue()

	case *commonv1.AnyValue_KvlistValue:
		kvl := a.GetKvlistValue()
		tv := make(map[string]interface{}, len(kvl.Values))
		for _, kv := range kvl.Values {
			tv[kv.Key] = anyValueToRaw(kv.Value)
		}
		v = tv

	case *commonv1.AnyValue_ArrayValue:
		av := a.GetArrayValue()
		tv := make([]interface{}, len(av.Values))
		for i := range av.Values {
			tv[i] = anyValueToRaw(av.Values[i])
		}
		v = tv
	}
	return v
}

// Is equivalent to strconv.FormatFloat(f, 'g', -1, 64), but hardcodes a few common cases for increased efficiency.
func float64ToDimValue(f float64) string {
	// Parameters below are the same used by Prometheus
	// see https://github.com/prometheus/common/blob/b5fe7d854c42dc7842e48d1ca58f60feae09d77b/expfmt/text_create.go#L450
	// SignalFx agent uses a different pattern
	// https://github.com/signalfx/signalfx-agent/blob/5779a3de0c9861fa07316fd11b3c4ff38c0d78f0/internal/monitors/prometheusexporter/conversion.go#L77
	// The important issue here is consistency with the exporter, opting for the
	// more common one used by Prometheus.
	switch {
	case f == 0:
		return "0"
	case f == 1:
		return "1"
	case math.IsInf(f, +1):
		return "+Inf"
	default:
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
}

func numberToSignalFxValue(in *metricsv1.NumberDataPoint) datapoint.Value {
	v := in.GetValue()
	switch n := v.(type) {
	case *metricsv1.NumberDataPoint_AsDouble:
		return datapoint.NewFloatValue(n.AsDouble)
	case *metricsv1.NumberDataPoint_AsInt:
		return datapoint.NewIntValue(n.AsInt)
	}
	return nil
}

func mergeStringMaps(ms ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, m := range ms {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
