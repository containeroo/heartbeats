package history

import "context"

type MockStore struct {
	RecordEventFunc   func(context.Context, Event) error
	GetEventsFunc     func() []Event
	GetEventsByIDFunc func(string) []Event
}

func (m *MockStore) RecordEvent(ctx context.Context, e Event) error {
	if m.RecordEventFunc != nil {
		return m.RecordEventFunc(ctx, e)
	}
	return nil
}

func (m *MockStore) GetEvents() []Event {
	if m.GetEventsFunc != nil {
		return m.GetEventsFunc()
	}
	return nil
}

func (m *MockStore) GetEventsByID(id string) []Event {
	if m.GetEventsByIDFunc != nil {
		return m.GetEventsByIDFunc(id)
	}
	return nil
}
