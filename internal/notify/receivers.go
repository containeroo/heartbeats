package notify

import (
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/config"

	kit "github.com/containeroo/notifykit/notify"
	"github.com/containeroo/notifykit/targets/email"
	"github.com/containeroo/notifykit/targets/webhook"
	"github.com/containeroo/notifykit/templates"
)

const (
	defaultSubjectTemplate = `[{{ .Title }}] {{ .Status }}`
	defaultWebhookTimeout  = 10 * time.Second
	defaultEmailPort       = 587
	defaultBodyLimit       = 4096
)

var templateFuncs = []templates.Option{
	templates.WithDefaultFuncs(),
	templates.WithFunc("formatDuration", formatDuration),
	templates.WithFunc("isRecent", isRecent),
	templates.WithFunc("ago", ago),
	templates.WithFunc("join", join),
}

// ReceiverRoutes maps heartbeat IDs to explicit notifykit receiver IDs.
type ReceiverRoutes map[string][]kit.ReceiverID

// ReceiversFromConfig builds notifykit receivers and heartbeat routes from config.
func ReceiversFromConfig(templateFS fs.FS, cfg *config.Config, logger *slog.Logger) (kit.Receivers, ReceiverRoutes, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}

	receivers := make(kit.Receivers)
	routes := make(ReceiverRoutes, len(cfg.Heartbeats))

	for _, heartbeatID := range sortedHeartbeatIDs(cfg.Heartbeats) {
		hb := cfg.Heartbeats[heartbeatID]
		for _, receiverName := range hb.Receivers {
			receiverCfg, ok := cfg.Receivers[receiverName]
			if !ok {
				return nil, nil, fmt.Errorf("heartbeat %q references unknown receiver %q", heartbeatID, receiverName)
			}

			receiver, err := receiverFromConfig(templateFS, heartbeatID, hb, receiverName, receiverCfg, logger)
			if err != nil {
				return nil, nil, fmt.Errorf("heartbeat %q receiver %q: %w", heartbeatID, receiverName, err)
			}
			receivers[receiver.ID] = receiver
			routes[heartbeatID] = append(routes[heartbeatID], receiver.ID)
		}
	}

	return receivers, routes, nil
}

// ReplaceReceivers swaps the contents of dst with src while keeping the same map instance.
func ReplaceReceivers(dst, src kit.Receivers) {
	for id := range dst {
		delete(dst, id)
	}
	maps.Copy(dst, src)
}

// ReplaceRoutes swaps the contents of dst with src while keeping the same map instance.
func ReplaceRoutes(dst, src ReceiverRoutes) {
	for id := range dst {
		delete(dst, id)
	}
	for id, receivers := range src {
		dst[id] = append([]kit.ReceiverID(nil), receivers...)
	}
}

// ReceiverIDs returns explicit receiver IDs for a heartbeat.
func (r ReceiverRoutes) ReceiverIDs(heartbeatID string) []kit.ReceiverID {
	if len(r) == 0 {
		return nil
	}
	return append([]kit.ReceiverID(nil), r[heartbeatID]...)
}

func receiverFromConfig(
	templateFS fs.FS,
	heartbeatID string,
	hb config.HeartbeatConfig,
	receiverName string,
	receiverCfg config.ReceiverConfig,
	logger *slog.Logger,
) (*kit.Receiver, error) {
	targets := make([]kit.Target, 0, len(receiverCfg.Webhooks)+len(receiverCfg.Emails))

	webhookTargets, err := webhookTargetsFromConfig(templateFS, hb, receiverName, receiverCfg, logger)
	if err != nil {
		return nil, err
	}
	targets = append(targets, webhookTargets...)

	emailTargets, err := emailTargetsFromConfig(templateFS, hb, receiverName, receiverCfg)
	if err != nil {
		return nil, err
	}
	targets = append(targets, emailTargets...)

	if len(targets) == 0 {
		return nil, fmt.Errorf("receiver has no targets")
	}

	return &kit.Receiver{
		ID:         receiverID(heartbeatID, receiverName),
		Name:       receiverName,
		Targets:    targets,
		CustomData: varsFromConfig(receiverCfg.Vars),
	}, nil
}

func webhookTargetsFromConfig(
	templateFS fs.FS,
	hb config.HeartbeatConfig,
	receiverName string,
	receiverCfg config.ReceiverConfig,
	logger *slog.Logger,
) ([]kit.Target, error) {
	if len(receiverCfg.Webhooks) == 0 {
		return nil, nil
	}

	out := make([]kit.Target, 0, len(receiverCfg.Webhooks))
	for idx, cfg := range receiverCfg.Webhooks {
		tmpl, err := templates.LoadSource(
			templateFS,
			webhookTemplateSource(firstNonEmpty(cfg.Template, hb.WebhookTemplate)),
			templateFuncs...,
		)
		if err != nil {
			return nil, fmt.Errorf("load webhook[%d] template: %w", idx, err)
		}

		subjectTmpl, err := templates.ParseStringTemplate(
			"webhook-subject",
			subjectTemplate(hb.SubjectTmpl, cfg.SubjectTmpl),
			templateFuncs...,
		)
		if err != nil {
			return nil, fmt.Errorf("parse webhook[%d] subject template: %w", idx, err)
		}

		method := strings.TrimSpace(cfg.Method)
		if method == "" {
			method = http.MethodPost
		}

		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultWebhookTimeout
		}

		bodyLimit := cfg.ResponseBodyLimit
		if bodyLimit <= 0 {
			bodyLimit = defaultBodyLimit
		}

		logResponse := cfg.LogResponse
		if logResponse == "" {
			logResponse = webhook.LogResponseSummary
		}

		clientOptions := []webhook.ClientOption{webhook.WithProxyFromEnvironment()}
		if cfg.InsecureSkipVerify {
			clientOptions = append(clientOptions, webhook.WithSkipTLSVerify())
		}

		target := webhook.NewFromTarget(
			webhook.Target{
				Name:              indexedTargetName(receiverName, idx, len(receiverCfg.Webhooks)),
				URL:               cfg.URL,
				Method:            method,
				Headers:           maps.Clone(cfg.Headers),
				Template:          tmpl,
				TitleTmpl:         subjectTmpl,
				ValidateJSON:      true,
				LogResponse:       logResponse,
				ResponseBodyLimit: bodyLimit,
			},
			webhook.WithClient(webhook.NewClient(timeout, clientOptions...)),
			webhook.WithLogger(logger),
		)

		if err := validateWebhookTarget(target, receiverName, receiverCfg.Vars); err != nil {
			return nil, fmt.Errorf("validate webhook[%d] template: %w", idx, err)
		}

		out = append(out, target)
	}

	return out, nil
}

func emailTargetsFromConfig(
	templateFS fs.FS,
	hb config.HeartbeatConfig,
	receiverName string,
	receiverCfg config.ReceiverConfig,
) ([]kit.Target, error) {
	if len(receiverCfg.Emails) == 0 {
		return nil, nil
	}

	out := make([]kit.Target, 0, len(receiverCfg.Emails))
	for idx, cfg := range receiverCfg.Emails {
		tmpl, err := templates.LoadSource(
			templateFS,
			emailTemplateSource(firstNonEmpty(cfg.Template, hb.EmailTemplate)),
			templateFuncs...,
		)
		if err != nil {
			return nil, fmt.Errorf("load email[%d] template: %w", idx, err)
		}

		subjectTmpl, err := templates.ParseStringTemplate(
			"email-subject",
			subjectTemplate(hb.SubjectTmpl, cfg.SubjectTmpl),
			templateFuncs...,
		)
		if err != nil {
			return nil, fmt.Errorf("parse email[%d] subject template: %w", idx, err)
		}

		port := cfg.Port
		if port == 0 {
			port = defaultEmailPort
		}

		target := email.NewFromTarget(email.Target{
			Name:          indexedTargetName(receiverName, idx, len(receiverCfg.Emails)),
			Host:          cfg.Host,
			Port:          port,
			User:          cfg.User,
			Pass:          cfg.Pass,
			From:          cfg.From,
			To:            append([]string(nil), cfg.To...),
			Headers:       maps.Clone(cfg.Headers),
			SkipTLSVerify: cfg.InsecureSkipVerify,
			Template:      tmpl,
			SubjectTmpl:   subjectTmpl,
		})

		if err := validateEmailTarget(target, receiverName, receiverCfg.Vars); err != nil {
			return nil, fmt.Errorf("validate email[%d] template: %w", idx, err)
		}

		out = append(out, target)
	}

	return out, nil
}

func receiverID(heartbeatID, receiverName string) kit.ReceiverID {
	return kit.ReceiverID("heartbeat." + heartbeatID + ".receiver." + receiverName)
}

func indexedTargetName(receiverName string, idx, total int) string {
	if total <= 1 {
		return receiverName
	}
	return fmt.Sprintf("%s.%d", receiverName, idx+1)
}

func subjectTemplate(heartbeatTemplate, receiverTemplate string) string {
	if strings.TrimSpace(receiverTemplate) != "" {
		return receiverTemplate
	}
	if strings.TrimSpace(heartbeatTemplate) != "" {
		return heartbeatTemplate
	}
	return defaultSubjectTemplate
}

func webhookTemplateSource(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default":
		return "builtin:default"
	case "slack":
		return "builtin:slack"
	default:
		return value
	}
}

func emailTemplateSource(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "email":
		return "builtin:email"
	default:
		return value
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func validateWebhookTarget(target *webhook.Target, receiverName string, vars map[string]any) error {
	for _, event := range validationEvents() {
		if err := target.Validate(kit.Payload{
			Notification: event,
			Receiver:     receiverName,
			CustomData:   varsFromConfig(vars),
		}); err != nil {
			return err
		}
	}
	return nil
}

func validateEmailTarget(target *email.Target, receiverName string, vars map[string]any) error {
	for _, event := range validationEvents() {
		if err := target.Validate(kit.Payload{
			Notification: event,
			Receiver:     receiverName,
			CustomData:   varsFromConfig(vars),
		}); err != nil {
			return err
		}
	}
	return nil
}

func validationEvents() []*Event {
	now := time.Now()
	return []*Event{
		NewEvent("validation", "Validation", "missing", "payload", 8*time.Second, now, 5*time.Second, 3*time.Second, nil),
		NewEvent("validation", "Validation", "recovered", "payload", 0, now, 5*time.Second, 3*time.Second, nil),
	}
}

// formatDuration renders a duration with heartbeat-specific rounding behavior.
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Second {
		return d.Truncate(time.Millisecond).String()
	}
	return d.Truncate(time.Second).String()
}

// isRecent reports whether a time value is within maxAge.
//
// It supports both direct and pipeline usage:
//
//	{{ isRecent .Timestamp "5m" }}
//	{{ .Timestamp | isRecent "5m" }}
func isRecent(first, second any) bool {
	if maxAge, ok := durationValue(first); ok {
		t, ok := timeValue(second)
		return isTimeRecent(t, maxAge, ok)
	}
	if maxAge, ok := durationValue(second); ok {
		t, ok := timeValue(first)
		return isTimeRecent(t, maxAge, ok)
	}
	return false
}

func isTimeRecent(t time.Time, maxAge time.Duration, ok bool) bool {
	if !ok || t.IsZero() || maxAge <= 0 {
		return false
	}

	age := time.Since(t)
	return age >= 0 && age <= maxAge
}

// ago renders how long ago a time value happened.
//
// Zero, nil, empty, or unsupported values render as "never".
func ago(value any) string {
	t, ok := timeValue(value)
	if !ok || t.IsZero() {
		return "never"
	}

	d := time.Since(t)
	if d < 0 {
		return "in " + formatDuration(d)
	}
	return formatDuration(d) + " ago"
}

// join renders a slice or array as a string joined by sep.
//
// It supports both direct and pipeline usage:
//
//	{{ join .Tags "," }}
//	{{ .Tags | join "," }}
func join(first, second any) string {
	if sep, ok := first.(string); ok {
		return joinValues(sep, second)
	}
	if sep, ok := second.(string); ok {
		return joinValues(sep, first)
	}
	return joinValues("", first)
}

func joinValues(sep string, values any) string {
	if values == nil {
		return ""
	}

	switch v := values.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, sep)
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, strings.TrimSpace(fmt.Sprint(item)))
		}
		return strings.Join(parts, sep)
	}

	rv := reflect.ValueOf(values)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		parts := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts = append(parts, strings.TrimSpace(fmt.Sprint(rv.Index(i).Interface())))
		}
		return strings.Join(parts, sep)
	default:
		return fmt.Sprint(values)
	}
}

func timeValue(value any) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case *time.Time:
		if v == nil {
			return time.Time{}, false
		}
		return *v, true
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return time.Time{}, false
		}
		t, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	default:
		return time.Time{}, false
	}
}

func durationValue(value any) (time.Duration, bool) {
	switch v := value.(type) {
	case time.Duration:
		return v, true
	case *time.Duration:
		if v == nil {
			return 0, false
		}
		return *v, true
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return 0, false
		}
		d, err := time.ParseDuration(text)
		if err != nil {
			return 0, false
		}
		return d, true
	default:
		return 0, false
	}
}

func sortedHeartbeatIDs(values map[string]config.HeartbeatConfig) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
