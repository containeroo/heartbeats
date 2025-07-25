package history

import "context"

type MockStore struct {
	RecordEventFunc   func(context.Context, Event) error
	GetEventsFunc     func() []Event
	GetEventsByIDFunc func(string) []Event
	ByteSizeFunc      func() int
}

func (m *MockStore) Append(ctx context.Context, e Event) error {
	if m.RecordEventFunc != nil {
		return m.RecordEventFunc(ctx, e)
	}
	return nil
}

func (m *MockStore) List() []Event {
	if m.GetEventsFunc != nil {
		return m.GetEventsFunc()
	}
	return nil
}

func (m *MockStore) ListByID(id string) []Event {
	if m.GetEventsByIDFunc != nil {
		return m.GetEventsByIDFunc(id)
	}
	return nil
}

func (m *MockStore) ByteSize() int {
	if m.ByteSizeFunc != nil {
		return m.ByteSizeFunc()
	}
	return 0
}
