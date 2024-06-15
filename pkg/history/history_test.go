package history

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistory_NewHistory(t *testing.T) {
	_, err := NewHistory(5, 120)
	assert.Error(t, err)
}

func TestHistory_Add(t *testing.T) {
	t.Run("Add Entries", func(t *testing.T) {
		h, err := NewHistory(5, 20)
		assert.NoError(t, err)

		h.Add(Beat, "Beat message", nil)
		assert.Equal(t, 1, len(h.GetAllEntries()))

		h.Add(Interval, "Interval message", nil)
		h.Add(Grace, "Grace message", nil)
		h.Add(Expired, "Expired message", nil)
		h.Add(Send, "Send message", nil)
		assert.Equal(t, 5, len(h.GetAllEntries()))

		h.Add(Beat, "New Beat message", nil)
		assert.Equal(t, 4, len(h.GetAllEntries())) // Reduced by 20% (5 * 0.8 = 4)
	})

	t.Run("Add Entries, reduce odd", func(t *testing.T) {
		h, err := NewHistory(5, 25)
		assert.NoError(t, err)

		h.Add(Beat, "Beat message", nil)
		assert.Equal(t, 1, len(h.GetAllEntries()))

		h.Add(Interval, "Interval message", nil)
		h.Add(Grace, "Grace message", nil)
		h.Add(Expired, "Expired message", nil)
		h.Add(Send, "Send message", nil)
		assert.Equal(t, 5, len(h.GetAllEntries()))

		h.Add(Beat, "New Beat message", nil)
		assert.Equal(t, 4, len(h.GetAllEntries())) // Reduced by 25% (5 * 0.75 ~= 4)
	})
}

func TestHistory_GetAllEntries(t *testing.T) {
	h, err := NewHistory(5, 20)
	assert.NoError(t, err)

	t.Run("AddEntries", func(t *testing.T) {
		h.Add(Beat, "Beat message", nil)
		h.Add(Interval, "Interval message", nil)
	})

	t.Run("VerifyEntries", func(t *testing.T) {
		entries := h.GetAllEntries()
		assert.Equal(t, 2, len(entries), "Expected 2 entries")
		assert.Equal(t, Beat, entries[0].Event, "Expected first entry to be a Beat event")
		assert.Equal(t, "Beat message", entries[0].Message, "Expected first entry message to be 'Beat message'")
		assert.Equal(t, Interval, entries[1].Event, "Expected second entry to be an Interval event")
		assert.Equal(t, "Interval message", entries[1].Message, "Expected second entry message to be 'Interval message'")
	})
}

func TestStore_Add_Get_Delete(t *testing.T) {
	store := NewStore()
	h, err := NewHistory(5, 20)
	assert.NoError(t, err)

	t.Run("Add", func(t *testing.T) {
		err = store.Add("test", h)
		assert.NoError(t, err, "Expected no error when adding a history")
	})

	t.Run("Get", func(t *testing.T) {
		retrieved := store.Get("test")
		assert.NotNil(t, retrieved, "Expected to retrieve the added history")
	})

	t.Run("Delete", func(t *testing.T) {
		store.Delete("test")
		retrieved := store.Get("test")
		assert.Nil(t, retrieved, "Expected history to be deleted")
	})
}

func TestStore_AddDuplicate(t *testing.T) {
	store := NewStore()
	h, err := NewHistory(5, 20)
	assert.NoError(t, err)

	t.Run("Add", func(t *testing.T) {
		err = store.Add("test", h)
		assert.NoError(t, err)
	})

	t.Run("Duplicate", func(t *testing.T) {
		err = store.Add("test", h)
		assert.Error(t, err)
	})
}

func TestEvent_String(t *testing.T) {
	t.Run("TestBeat", func(t *testing.T) {
		assert.Equal(t, "BEAT", Beat.String())
	})

	t.Run("TestInterval", func(t *testing.T) {
		assert.Equal(t, "INTERVAL", Interval.String())
	})

	t.Run("TestGrace", func(t *testing.T) {
		assert.Equal(t, "GRACE", Grace.String())
	})

	t.Run("TestExpired", func(t *testing.T) {
		assert.Equal(t, "EXPIRED", Expired.String())
	})

	t.Run("TestSend", func(t *testing.T) {
		assert.Equal(t, "SEND", Send.String())
	})
}

func TestHistoryEntry(t *testing.T) {
	entry := HistoryEntry{
		Time:    time.Now(),
		Event:   Beat,
		Message: "Test message",
		Details: map[string]string{"key": "value"},
	}

	t.Run("TestEvent", func(t *testing.T) {
		assert.Equal(t, Beat, entry.Event)
	})

	t.Run("TestMessage", func(t *testing.T) {
		assert.Equal(t, "Test message", entry.Message)
	})

	t.Run("TestDetails", func(t *testing.T) {
		assert.Equal(t, "value", entry.Details["key"])
	})
}

func TestMarshalYAML(t *testing.T) {
	store := NewStore()
	h, err := NewHistory(5, 20)

	t.Run("TestNewHistory", func(t *testing.T) {
		assert.NoError(t, err)
	})

	t.Run("TestAddHistory", func(t *testing.T) {
		err = store.Add("test", h)
		assert.NoError(t, err)
	})

	t.Run("TestMarshalStore", func(t *testing.T) {
		data, err := store.MarshalYAML()
		assert.NoError(t, err)
		assert.NotNil(t, data)
	})
}
