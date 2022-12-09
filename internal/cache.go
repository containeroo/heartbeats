package internal

import (
	"fmt"
	"sync"
	"time"
)

// Enum for Event
type Event int16

const (
	Ping Event = iota
	Grace
	Failed
	Send
)

func (s Event) String() string {
	return [...]string{"PING", "GRACE", "FAILED", "SEND"}[s]
}

type History struct {
	Time    time.Time `mapstructure:"time"`
	Event   Event     `mapstructure:"event"`
	Message string    `mapstructure:"message"`
}

type cachedHistory struct {
	Histories []History `mapstructure:"history"`
}

type LocalCache struct {
	wg      sync.WaitGroup       `mapstructure:"waitgroup"`
	mu      sync.RWMutex         `mapstructure:"mutex"`
	maxSize int                  `mapstructure:"max_size"`
	reduce  int                  `mapstructure:"reduce"`
	History map[string][]History `mapstructure:"history"`
}

func NewLocalCache(maxSize int, reduce int) *LocalCache {
	lc := &LocalCache{
		History: make(map[string][]History),
		maxSize: maxSize,
		reduce:  reduce,
	}

	lc.wg.Add(1)

	return lc
}

func reduceCache(maxSize int, reduce int, history map[string][]History) {
	for k, v := range history {
		if len(v) > maxSize {
			history[k] = v[len(v)-reduce:]
		}
	}
}

func (lc *LocalCache) Add(heartbeatName string, h History) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if h.Time.IsZero() {
		h.Time = time.Now()
	}

	if len(lc.History[heartbeatName]) > lc.maxSize {
		reduceCache(lc.maxSize, lc.reduce, lc.History)
	}

	lc.History[heartbeatName] = append(lc.History[heartbeatName], h)

}

func (lc *LocalCache) Get(heartbeatName string) ([]History, error) {
	if _, ok := lc.History[heartbeatName]; !ok {
		return nil, fmt.Errorf("History for Heartbeat %s does not exist", heartbeatName)
	}

	lc.mu.RLock()
	defer lc.mu.RUnlock()

	return lc.History[heartbeatName], nil
}
