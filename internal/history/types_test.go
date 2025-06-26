package history

import (
	"testing"

	"github.com/containeroo/heartbeats/internal/flag"

	"github.com/stretchr/testify/assert"
)

func TestInitHistory(t *testing.T) {
	t.Parallel()

	t.Run("Ring backend returns store", func(t *testing.T) {
		t.Parallel()

		flags := flag.Options{
			HistorySize: 5,
		}
		store, err := InitializeHistory(flags)
		assert.NoError(t, err)
		assert.NotNil(t, store)
	})
}
