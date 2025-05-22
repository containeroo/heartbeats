package notifier

import (
	"context"
	"sync"
	"time"
)

// MockNotifier is a test implementation of the Notifier interface.
type MockNotifier struct {
	mu         sync.Mutex                                             // guards access to called and last
	called     bool                                                   // true if Notify was called
	last       NotificationData                                       // holds the last NotificationData passed to Notify
	TypeName   string                                                 // optional custom type name returned by Type()
	Sent       time.Time                                              // mock timestamp returned by LastSent()
	Success    *bool                                                  // mock success status returned by Successful()
	FormatFunc func(NotificationData) (NotificationData, error)       // optional override for Format behavior
	NotifyFunc func(ctx context.Context, data NotificationData) error // optional override for Notify behavior
}

func (m *MockNotifier) Notify(ctx context.Context, data NotificationData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.last = data

	if m.NotifyFunc != nil {
		return m.NotifyFunc(ctx, data)
	}

	return nil
}

func (m *MockNotifier) Format(data NotificationData) (NotificationData, error) {
	if m.FormatFunc != nil {
		return m.FormatFunc(data)
	}
	return data, nil
}

func (m *MockNotifier) Validate() error {
	return nil
}

func (m *MockNotifier) Resolve() error {
	return nil
}

func (m *MockNotifier) Type() string {
	if m.TypeName != "" {
		return m.TypeName
	}
	return "mock"
}

func (m *MockNotifier) LastSent() time.Time {
	return m.Sent
}

func (m *MockNotifier) Successful() *bool {
	return m.Success
}
