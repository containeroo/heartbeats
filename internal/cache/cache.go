package cache

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var Local *localCache

// Enum for Event
type Event int16

// Event Enum
const (
	EventPing Event = iota
	EventGrace
	EventFailed
	EventSend
)

// String returns the string representation of the Event
func (s Event) String() string {
	return [...]string{"PING", "GRACE", "FAILED", "SEND"}[s]
}

// History is the struct for the history of a heartbeat
type History struct {
	Time    time.Time         `mapstructure:"time"`
	Event   Event             `mapstructure:"event"`
	Message string            `mapstructure:"message"`
	Details map[string]string `mapstructure:"details"`
}

// localCache is the struct for the local cache
type localCache struct {
	wg      sync.WaitGroup       `mapstructure:"waitgroup"`
	mu      sync.RWMutex         `mapstructure:"mutex"`
	maxSize int                  `mapstructure:"max_size"`
	reduce  int                  `mapstructure:"reduce"`
	History map[string][]History `mapstructure:"history"`
}

// New creates a new localCache
func New(maxSize int, reduce int) *localCache {
	lc := &localCache{
		History: make(map[string][]History),
		maxSize: maxSize,
		reduce:  reduce,
	}

	lc.wg.Add(1)

	return lc
}

// Add adds a new history item to the local cache
func (lc *localCache) Add(heartbeatName string, h History) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if h.Time.IsZero() {
		h.Time = time.Now()
	}

	if len(lc.History[heartbeatName]) > lc.maxSize {
		lc.History[heartbeatName] = lc.History[heartbeatName][lc.reduce:]
		log.Debugf("%s Reduced history to %d", heartbeatName, len(lc.History[heartbeatName]))
	}

	lc.History[heartbeatName] = append(lc.History[heartbeatName], h)

}

// Get returns the history of a heartbeat
func (lc *localCache) Get(heartbeatName string) ([]History, error) {
	if _, ok := lc.History[heartbeatName]; !ok {
		return nil, fmt.Errorf("History for Heartbeat %s does not exist", heartbeatName)
	}

	lc.mu.RLock()
	defer lc.mu.RUnlock()

	return lc.History[heartbeatName], nil
}
