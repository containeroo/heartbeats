package history

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitHistory(t *testing.T) {
	t.Parallel()

	t.Run("Ring backend returns store", func(t *testing.T) {
		t.Parallel()

		store, err := InitializeHistory(5)
		assert.NoError(t, err)
		assert.NotNil(t, store)
	})
}
