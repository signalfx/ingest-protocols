package filtering

import (
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/datapoint/dptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestFilteredForwarder struct {
	FilteredForwarder
}

func Test(t *testing.T) {
	var (
		imNotAMatcher = "im.not.a.matcher"
		cpuIdleString = "cpu.idle"
	)

	t.Run("bad regexes throw errors", func(t *testing.T) {
		filters := FilterObj{
			Deny: []string{"["},
		}
		forwarder := TestFilteredForwarder{}
		err := forwarder.Setup(&filters)
		assert.NotNil(t, err)
		filters = FilterObj{
			Allow: []string{"["},
		}
		err = forwarder.Setup(&filters)
		assert.NotNil(t, err)
	})
	t.Run("good regexes don't throw errors and work together", func(t *testing.T) {
		filters := FilterObj{
			Deny:  []string{"^cpu.*"},
			Allow: []string{"^cpu.idle"},
		}
		forwarder := TestFilteredForwarder{}
		err := forwarder.Setup(&filters)
		require.NoError(t, err)
		dp := dptest.DP()
		t.Run("cpu.idle is allowed even though cpu.* is denied", func(t *testing.T) {
			dp.Metric = cpuIdleString
			datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
			assert.Equal(t, 1, len(datapoints))
			assert.Equal(t, int64(0), forwarder.FilteredDatapoints)
			t.Run("metrics that match deny but not allow are denied", func(t *testing.T) {
				dp.Metric = "cpu.user"
				assert.Equal(t, int64(0), forwarder.FilteredDatapoints)
				datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
				assert.Equal(t, 0, len(datapoints))
				assert.Equal(t, int64(1), forwarder.FilteredDatapoints)
				t.Run("other metrics that do not explicitly pass allow are denied", func(t *testing.T) {
					dp.Metric = "i.should.be.denied"
					datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
					assert.Equal(t, 0, len(datapoints))
					assert.Equal(t, int64(2), forwarder.FilteredDatapoints)
				})
			})
		})
	})
	t.Run("allow by itself denies what it doesn't match", func(t *testing.T) {
		filters := FilterObj{
			Allow: []string{"^cpu.*"},
		}
		forwarder := TestFilteredForwarder{}
		err := forwarder.Setup(&filters)
		require.NoError(t, err)
		dp := dptest.DP()
		t.Run("metrics starting with cpu get accepted", func(t *testing.T) {
			dp.Metric = cpuIdleString
			datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
			assert.Equal(t, 1, len(datapoints))
			assert.Equal(t, int64(0), forwarder.FilteredDatapoints)
			t.Run("metrics that don't match allow are denied", func(t *testing.T) {
				dp.Metric = imNotAMatcher
				datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
				assert.Equal(t, 0, len(datapoints))
				assert.Equal(t, int64(1), forwarder.FilteredDatapoints)
			})
		})
	})
	t.Run("deny by itself accepts what it doesn't match", func(t *testing.T) {
		filters := FilterObj{
			Deny: []string{"^cpu.*"},
		}
		forwarder := TestFilteredForwarder{}
		err := forwarder.Setup(&filters)
		require.NoError(t, err)
		dp := dptest.DP()
		t.Run("metrics starting with cpu get denied", func(t *testing.T) {
			dp.Metric = cpuIdleString
			datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
			assert.Equal(t, 0, len(datapoints))
			assert.Equal(t, int64(1), forwarder.FilteredDatapoints)
			t.Run("metrics that don't match cpu are accepted", func(t *testing.T) {
				dp.Metric = imNotAMatcher
				datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
				assert.Equal(t, 1, len(datapoints))
				assert.Equal(t, int64(1), forwarder.FilteredDatapoints)
				assert.Equal(t, 1, len(forwarder.GetFilteredDatapoints()))
			})
		})
	})
	t.Run("no rules let everything through", func(t *testing.T) {
		filters := FilterObj{}
		forwarder := TestFilteredForwarder{}
		err := forwarder.Setup(&filters)
		require.NoError(t, err)
		dp := dptest.DP()
		t.Run("metrics starting with cpu get accepted", func(t *testing.T) {
			dp.Metric = cpuIdleString
			datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
			assert.Equal(t, 1, len(datapoints))
			assert.Equal(t, int64(0), forwarder.FilteredDatapoints)
			t.Run("metrics that don't match cpu are accepted", func(t *testing.T) {
				dp.Metric = imNotAMatcher
				datapoints := forwarder.FilterDatapoints([]*datapoint.Datapoint{dp})
				assert.Equal(t, 1, len(datapoints))
				assert.Equal(t, int64(0), forwarder.FilteredDatapoints)
				assert.Equal(t, 1, len(forwarder.GetFilteredDatapoints()))
			})
		})
	})
}
