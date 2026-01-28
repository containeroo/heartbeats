package manager

import (
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/notify/targets"
	"github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/templates"
)

// buildReceivers resolves receiver configurations for a heartbeat.
func buildReceivers(
	cfg *config.Config,
	sc config.HeartbeatConfig,
	defaultWebhook *templates.Template,
	defaultTitle *templates.StringTemplate,
	defaultEmail *templates.StringTemplate,
	templateFS fs.FS,
	logger *slog.Logger,
) (map[string]*types.Receiver, []string, error) {
	receivers := make(map[string]*types.Receiver, len(sc.Receivers))
	receiverNames := make([]string, 0, len(sc.Receivers))
	for _, name := range sc.Receivers {
		base, ok := cfg.Receivers[name]
		if !ok {
			return nil, nil, fmt.Errorf("unknown receiver %q", name)
		}
		rcv := &types.Receiver{
			Name: name,
			Retry: types.RetryConfig{
				Count: base.Retry.Count,
				Delay: base.Retry.Delay,
			},
			Targets: nil,
			Vars:    base.Vars,
		}
		webhookTargets, err := buildWebhookTargets(
			sc,
			base,
			rcv,
			defaultWebhook,
			defaultTitle,
			templateFS,
			logger,
		)
		if err != nil {
			return nil, nil, err
		}
		rcv.Targets = append(rcv.Targets, webhookTargets...)
		emailTargets, err := buildEmailTargets(
			sc,
			base,
			rcv,
			defaultEmail,
			defaultTitle,
			templateFS,
		)
		if err != nil {
			return nil, nil, err
		}
		rcv.Targets = append(rcv.Targets, emailTargets...)
		receivers[name] = rcv
		receiverNames = append(receiverNames, name)
	}
	return receivers, receiverNames, nil
}

// buildWebhookTargets builds webhook targets.
func buildWebhookTargets(
	sc config.HeartbeatConfig,
	base config.ReceiverConfig,
	rcv *types.Receiver,
	defaultWebhook *templates.Template,
	defaultTitle *templates.StringTemplate,
	templateFS fs.FS,
	logger *slog.Logger,
) ([]types.Target, error) {
	if len(base.Webhooks) == 0 {
		return nil, nil
	}
	out := make([]types.Target, 0, len(base.Webhooks))
	for _, webhook := range base.Webhooks {
		webhookTemplate, err := resolveWebhookTemplate(
			sc.WebhookTemplate,
			webhook.Template,
			defaultWebhook,
			templateFS,
		)
		if err != nil {
			return nil, err
		}
		titleTmpl, err := resolveTitleTemplate(
			sc.SubjectTmpl,
			webhook.SubjectTmpl,
			defaultTitle,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, targets.NewWebhookTarget(
			rcv.Name,
			webhook.URL,
			webhook.Headers,
			webhookTemplate,
			titleTmpl,
			rcv.Vars,
			logger,
		))
	}
	return out, nil
}

// buildEmailTargets builds email targets.
func buildEmailTargets(
	sc config.HeartbeatConfig,
	base config.ReceiverConfig,
	rcv *types.Receiver,
	defaultEmail *templates.StringTemplate,
	defaultTitle *templates.StringTemplate,
	templateFS fs.FS,
) ([]types.Target, error) {
	if len(base.Emails) == 0 {
		return nil, nil
	}
	out := make([]types.Target, 0, len(base.Emails))
	for _, email := range base.Emails {
		if email.Port == 0 {
			return nil, fmt.Errorf("email port is required")
		}
		emailTemplate, err := resolveEmailTemplate(
			sc.EmailTemplate,
			email.Template,
			defaultEmail,
			templateFS,
		)
		if err != nil {
			return nil, err
		}
		titleTmpl, err := resolveTitleTemplate(
			sc.SubjectTmpl,
			email.SubjectTmpl,
			defaultTitle,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, targets.NewEmailTarget(
			targets.EmailTarget{
				Host:               email.Host,
				Port:               email.Port,
				User:               email.User,
				Pass:               email.Pass,
				From:               email.From,
				To:                 email.To,
				StartTLS:           email.StartTLS,
				SSL:                email.SSL,
				InsecureSkipVerify: email.InsecureSkipVerify,
				Template:           emailTemplate,
				TitleTmpl:          titleTmpl,
				Receiver:           rcv.Name,
				Vars:               rcv.Vars,
			}),
		)
	}
	return out, nil
}

// resolveTitleTemplate picks and parses the title template for a target.
func resolveTitleTemplate(
	heartbeatTemplate, receiverTemplate string,
	fallback *templates.StringTemplate,
) (*templates.StringTemplate, error) {
	templateValue := strings.TrimSpace(receiverTemplate)
	if templateValue == "" {
		templateValue = strings.TrimSpace(heartbeatTemplate)
	}
	if templateValue == "" {
		return fallback, nil
	}
	return templates.ParseStringTemplate("title", templateValue)
}

// resolveWebhookTemplate picks and loads the webhook payload template.
func resolveWebhookTemplate(
	heartbeatPath, receiverPath string,
	fallback *templates.Template,
	templateFS fs.FS,
) (*templates.Template, error) {
	path := strings.TrimSpace(receiverPath)
	if path == "" {
		path = strings.TrimSpace(heartbeatPath)
	}
	if path == "" {
		return fallback, nil
	}
	if builtin, ok := builtinWebhookTemplates[strings.ToLower(path)]; ok {
		return templates.LoadFromFS(templateFS, builtin)
	}
	return templates.Load(path)
}

// resolveEmailTemplate picks and loads the email body template.
func resolveEmailTemplate(
	heartbeatPath, receiverPath string,
	fallback *templates.StringTemplate,
	templateFS fs.FS,
) (*templates.StringTemplate, error) {
	path := strings.TrimSpace(receiverPath)
	if path == "" {
		path = strings.TrimSpace(heartbeatPath)
	}
	if path == "" {
		return fallback, nil
	}
	if builtin, ok := builtinEmailTemplates[strings.ToLower(path)]; ok {
		return templates.LoadStringFromFS(templateFS, builtin)
	}
	return templates.LoadString(path)
}
