package notify

import "time"

const customDataVar = "CustomData"

// Data is the template context for heartbeat notifications.
type Data struct {
	HeartbeatID string
	Title       string
	Status      string
	Subject     string
	Payload     string
	Timestamp   time.Time
	Interval    time.Duration
	LateAfter   time.Duration
	Receiver    string
	Vars        map[string]any
	CustomData  map[string]string
	Since       time.Duration
}

// NewData builds template data from a heartbeat event and notifykit receiver vars.
func NewData(event Event, receiver string, vars map[string]any, subject string) Data {
	return Data{
		HeartbeatID: event.Heartbeat,
		Title:       event.TitleValue,
		Status:      event.StatusValue,
		Subject:     subject,
		Payload:     event.Body,
		Timestamp:   event.Time,
		Interval:    event.Interval,
		LateAfter:   event.LateAfter,
		Receiver:    receiver,
		Vars:        publicVars(vars),
		CustomData:  customData(vars),
		Since:       event.SinceValue,
	}
}

func varsFromConfig(vars map[string]any) map[string]any {
	if len(vars) == 0 {
		return nil
	}

	out := make(map[string]any, len(vars)+1)
	custom := make(map[string]string)
	for key, value := range vars {
		out[key] = value
		if text, ok := value.(string); ok {
			custom[key] = text
		}
	}
	if len(custom) > 0 {
		out[customDataVar] = custom
	}
	return out
}

func customData(vars map[string]any) map[string]string {
	if len(vars) == 0 {
		return nil
	}
	if custom, ok := vars[customDataVar].(map[string]string); ok {
		return cloneStringMap(custom)
	}

	custom := make(map[string]string, len(vars))
	for key, value := range vars {
		text, ok := value.(string)
		if ok {
			custom[key] = text
		}
	}
	if len(custom) == 0 {
		return nil
	}
	return custom
}

func publicVars(vars map[string]any) map[string]any {
	if len(vars) == 0 {
		return nil
	}
	out := make(map[string]any, len(vars))
	for key, value := range vars {
		if key == customDataVar {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
