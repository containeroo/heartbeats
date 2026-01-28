package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	htypes "github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notify/targets"
	ntypes "github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/runner"
)

// Service provides heartbeat updates and status snapshots.
// Store exposes read access to heartbeat state.
type Store interface {
	Get(id string) (*htypes.Heartbeat, bool)
	All() []*htypes.Heartbeat
}

// ReceiverStore provides receiver configuration access.
type ReceiverStore interface {
	Receivers() []*ntypes.Receiver
}

type Service struct {
	manager   Store
	receivers ReceiverStore
	history   history.Recorder
}

// Status is the public heartbeat state payload.
type Status struct {
	ID        string        `json:"id"`              // Heartbeat identifier.
	LastSeen  time.Time     `json:"last_seen"`       // Last received heartbeat time.
	Stage     runner.Stage  `json:"stage"`           // Current stage.
	LateAfter time.Duration `json:"lateAfter"`       // Late window duration.
	Interval  time.Duration `json:"interval"`        // Expected interval.
	SinceSeen time.Duration `json:"since_last_seen"` // Time since last heartbeat.
}

// HeartbeatSummary represents a UI-friendly heartbeat payload.
type HeartbeatSummary struct {
	ID               string   `json:"id"`
	Title            string   `json:"title,omitempty"`
	Status           string   `json:"status"`
	Interval         string   `json:"interval,omitempty"`
	IntervalSeconds  int64    `json:"intervalSeconds,omitempty"`
	LateAfter        string   `json:"lateAfter,omitempty"`
	LateAfterSeconds int64    `json:"lateAfterSeconds,omitempty"`
	LastBump         string   `json:"lastBump,omitempty"`
	Receivers        []string `json:"receivers,omitempty"`
	URL              string   `json:"url,omitempty"`
	HasHistory       bool     `json:"hasHistory,omitempty"`
}

// ReceiverSummary represents a UI-friendly receiver payload.
type ReceiverSummary struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Destination string `json:"destination"`
	LastSent    string `json:"lastSent,omitempty"`
	LastErr     string `json:"lastErr,omitempty"`
}

// NewService constructs a Service from a Manager.
func NewService(manager Store, receivers ReceiverStore, historyStore history.Recorder) *Service {
	return &Service{
		manager:   manager,
		receivers: receivers,
		history:   historyStore,
	}
}

// Update records a new heartbeat payload for the given id.
func (s *Service) Update(id string, payload string, now time.Time) error {
	hb, ok := s.manager.Get(id)
	if !ok {
		return fmt.Errorf("heartbeat %q not found", id)
	}
	if ok := hb.State.UpdateSeen(now, payload); !ok {
		s.recordHeartbeatReceived(id, now, len(payload), false)
		return errors.New("heartbeat mailbox full")
	}
	s.recordHeartbeatReceived(id, now, len(payload), true)
	return nil
}

// StatusByID returns the status for a heartbeat.
func (s *Service) StatusByID(id string) (Status, error) {
	hb, ok := s.manager.Get(id)
	if !ok {
		return Status{}, fmt.Errorf("heartbeat %q not found", id)
	}
	return buildStatus(hb), nil
}

// StatusAll returns a status list for all heartbeats.
func (s *Service) StatusAll() []Status {
	heartbeats := s.manager.All()
	out := make([]Status, 0, len(heartbeats))
	for _, hb := range heartbeats {
		out = append(out, buildStatus(hb))
	}
	return out
}

// HeartbeatSummaries returns a summary list for the UI.
func (s *Service) HeartbeatSummaries() []HeartbeatSummary {
	if s == nil || s.manager == nil {
		return nil
	}
	historyIndex := heartbeatHistoryIndex(s.history)
	heartbeats := s.manager.All()
	out := make([]HeartbeatSummary, 0, len(heartbeats))
	for _, hb := range heartbeats {
		if hb == nil || hb.State == nil {
			continue
		}
		snap := hb.State.Snapshot()
		item := HeartbeatSummary{
			ID:               hb.ID,
			Title:            hb.Title,
			Status:           snap.Stage.String(),
			Interval:         hb.Config.Interval.String(),
			IntervalSeconds:  int64(hb.Config.Interval.Seconds()),
			LateAfter:        hb.Config.LateAfter.String(),
			LateAfterSeconds: int64(hb.Config.LateAfter.Seconds()),
			Receivers:        hb.Receivers,
			HasHistory:       historyIndex[hb.ID],
		}
		if !snap.LastSeen.IsZero() {
			item.LastBump = snap.LastSeen.UTC().Format(time.RFC3339Nano)
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

// HeartbeatSummaryByID returns a heartbeat summary for a given id.
func (s *Service) HeartbeatSummaryByID(id string) (HeartbeatSummary, bool) {
	if s == nil || s.manager == nil {
		return HeartbeatSummary{}, false
	}
	hb, ok := s.manager.Get(id)
	if !ok || hb == nil || hb.State == nil {
		return HeartbeatSummary{}, false
	}
	historyIndex := heartbeatHistoryIndex(s.history)
	snap := hb.State.Snapshot()
	item := HeartbeatSummary{
		ID:               hb.ID,
		Title:            hb.Title,
		Status:           snap.Stage.String(),
		Interval:         hb.Config.Interval.String(),
		IntervalSeconds:  int64(hb.Config.Interval.Seconds()),
		LateAfter:        hb.Config.LateAfter.String(),
		LateAfterSeconds: int64(hb.Config.LateAfter.Seconds()),
		Receivers:        hb.Receivers,
		HasHistory:       historyIndex[hb.ID],
	}
	if !snap.LastSeen.IsZero() {
		item.LastBump = snap.LastSeen.UTC().Format(time.RFC3339Nano)
	}
	return item, true
}

// ReceiverSummaries returns a summary list for the UI.
func (s *Service) ReceiverSummaries() []ReceiverSummary {
	if s == nil || s.receivers == nil {
		return nil
	}
	statusIndex := receiverHistoryIndex(s.history)
	receivers := s.receivers.Receivers()
	out := make([]ReceiverSummary, 0)
	for _, rcv := range receivers {
		if rcv == nil {
			continue
		}
		for _, target := range rcv.Targets {
			if target == nil {
				continue
			}
			dest := targetDestination(target)
			summary := ReceiverSummary{
				ID:          rcv.Name,
				Type:        target.Type(),
				Destination: dest,
			}
			if status, ok := statusIndex[receiverStatusKey(rcv.Name, target.Type(), dest)]; ok {
				if !status.time.IsZero() {
					summary.LastSent = status.time.UTC().Format(time.RFC3339Nano)
				}
				if status.err != "" {
					summary.LastErr = status.err
				}
			}
			out = append(out, summary)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID == out[j].ID {
			return out[i].Type < out[j].Type
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// ReceiverSummaryByKey returns a receiver summary by receiver/type/target.
func (s *Service) ReceiverSummaryByKey(receiver, targetType, target string) (ReceiverSummary, bool) {
	if s == nil || s.receivers == nil {
		return ReceiverSummary{}, false
	}
	statusIndex := receiverHistoryIndex(s.history)
	receivers := s.receivers.Receivers()
	for _, rcv := range receivers {
		if rcv == nil || rcv.Name != receiver {
			continue
		}
		for _, t := range rcv.Targets {
			if t == nil {
				continue
			}
			dest := targetDestination(t)
			if t.Type() != targetType || dest != target {
				continue
			}
			summary := ReceiverSummary{
				ID:          rcv.Name,
				Type:        t.Type(),
				Destination: dest,
			}
			if status, ok := statusIndex[receiverStatusKey(rcv.Name, t.Type(), dest)]; ok {
				if !status.time.IsZero() {
					summary.LastSent = status.time.UTC().Format(time.RFC3339Nano)
				}
				if status.err != "" {
					summary.LastErr = status.err
				}
			}
			return summary, true
		}
	}
	return ReceiverSummary{}, false
}

// HistorySnapshot returns all history events.
func (s *Service) HistorySnapshot() []history.Event {
	if s == nil || s.history == nil {
		return nil
	}
	return s.history.List()
}

// HistoryStream subscribes to history events when supported.
func (s *Service) HistoryStream(buffer int) (<-chan history.Event, func()) {
	if s == nil || s.history == nil {
		return nil, func() {}
	}
	streamer, ok := s.history.(history.Streamer)
	if !ok {
		return nil, func() {}
	}
	return streamer.Subscribe(buffer)
}

// buildStatus builds a status snapshot for a heartbeat.
func buildStatus(hb *htypes.Heartbeat) Status {
	snap := hb.State.Snapshot()
	now := time.Now().UTC()
	st := Status{
		ID:        hb.ID,
		LastSeen:  snap.LastSeen,
		Stage:     snap.Stage,
		LateAfter: hb.Config.LateAfter,
		Interval:  hb.Config.Interval,
	}
	if !snap.LastSeen.IsZero() {
		st.SinceSeen = now.Sub(snap.LastSeen)
	}
	return st
}

func (s *Service) recordHeartbeatReceived(id string, now time.Time, payloadBytes int, enqueued bool) {
	if s.history == nil {
		return
	}
	s.history.Add(history.Event{
		Time:        now,
		Type:        history.EventHeartbeatReceived.String(),
		HeartbeatID: id,
		Fields: map[string]any{
			"payload_bytes": payloadBytes,
			"enqueued":      enqueued,
		},
	})
}

func targetDestination(target ntypes.Target) string {
	switch t := target.(type) {
	case *targets.WebhookTarget:
		return t.URL
	case *targets.EmailTarget:
		return strings.Join(t.To, ", ")
	default:
		return ""
	}
}

func heartbeatHistoryIndex(recorder history.Recorder) map[string]bool {
	out := make(map[string]bool)
	if recorder == nil {
		return out
	}
	for _, ev := range recorder.List() {
		if ev.HeartbeatID != "" {
			out[ev.HeartbeatID] = true
		}
	}
	return out
}

type receiverStatus struct {
	time time.Time
	err  string
}

func receiverHistoryIndex(recorder history.Recorder) map[string]receiverStatus {
	out := make(map[string]receiverStatus)
	if recorder == nil {
		return out
	}
	for _, ev := range recorder.List() {
		if ev.Receiver == "" || ev.TargetType == "" {
			continue
		}
		target, _ := ev.Fields["target"].(string)
		key := receiverStatusKey(ev.Receiver, ev.TargetType, target)
		current, ok := out[key]
		if ok && !ev.Time.After(current.time) {
			continue
		}
		status := receiverStatus{time: ev.Time}
		if ev.Type == history.EventNotificationFailed.String() {
			status.err = ev.Message
		}
		out[key] = status
	}
	return out
}

func receiverStatusKey(receiver, targetType, target string) string {
	return receiver + "|" + targetType + "|" + target
}
