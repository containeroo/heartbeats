package history

import "sort"

// ListRecent returns events sorted newest-first. If limit <= 0, returns all.
func ListRecent(store Store, limit int) []Event {
	events := store.List()
	sort.Slice(events, func(i, j int) bool {
		return events[j].Timestamp.Before(events[i].Timestamp)
	})
	if limit > 0 && len(events) > limit {
		return events[:limit]
	}
	return events
}

// ListByType returns events matching the given type.
func ListByType(store Store, t EventType) []Event {
	events := store.List()
	filtered := make([]Event, 0, len(events))
	for _, e := range events {
		if e.Type == t {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
