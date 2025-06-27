package history

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHistoryMetrics_Describe(t *testing.T) {
	t.Parallel()

	store := &MockStore{}
	collector := NewHistoryMetrics(store)

	ch := make(chan *prometheus.Desc, 1)
	collector.Describe(ch)

	desc := <-ch
	assert.Contains(t, desc.String(), "heartbeats_history_byte_size")
}

func TestHistoryMetrics_Collect(t *testing.T) {
	t.Parallel()

	store := &MockStore{}
	store.ByteSizeFunc = func() int {
		return 1234
	}
	collector := NewHistoryMetrics(store)

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	expected := `
# HELP heartbeats_history_byte_size Current size of the history store in bytes
# TYPE heartbeats_history_byte_size gauge
heartbeats_history_byte_size 1234
`
	err := testutil.GatherAndCompare(reg, strings.NewReader(expected), "heartbeats_history_byte_size")
	assert.NoError(t, err)
}
