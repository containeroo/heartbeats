package history

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitHistory(t *testing.T) {
	t.Parallel()

	t.Run("Ring backend returns store", func(t *testing.T) {
		t.Parallel()

		store, err := InitHistory(BackendTypeRingStore, "", 5)
		assert.NoError(t, err)
		assert.NotNil(t, store)
	})

	t.Run("Badger backend without path returns error", func(t *testing.T) {
		t.Parallel()

		store, err := InitHistory(BackendTypeBadger, "", 0)
		assert.Nil(t, store)
		assert.ErrorContains(t, err, "badger backend requires a path")
	})

	t.Run("Badger backend with path returns store", func(t *testing.T) {
		t.Parallel()

		store, err := InitHistory(BackendTypeBadger, t.TempDir(), 0)
		assert.NoError(t, err)
		assert.NotNil(t, store)
	})

	t.Run("Unknown backend returns error", func(t *testing.T) {
		t.Parallel()

		store, err := InitHistory("invalid-backend", "", 0)
		assert.Nil(t, store)
		assert.ErrorContains(t, err, "unknown history backend")
	})
}
