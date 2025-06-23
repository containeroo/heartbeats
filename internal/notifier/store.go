package notifier

import (
	"log/slog"

	"github.com/containeroo/heartbeats/pkg/notify/email"
	"github.com/containeroo/heartbeats/pkg/notify/msteams"
	"github.com/containeroo/heartbeats/pkg/notify/msteamsgraph"
	"github.com/containeroo/heartbeats/pkg/notify/slack"
)

// ReceiverStore holds a mapping from receiver ID to a slice of Notifiers.
type ReceiverStore struct {
	notifiers map[string][]Notifier // map of notifiers by receiver ID
}

// NewReceiverStore creates an empty ReceiverStore.
func NewReceiverStore() *ReceiverStore {
	return &ReceiverStore{notifiers: make(map[string][]Notifier)}
}

// Register adds a Notifier to the list for a given receiver ID.
func (s *ReceiverStore) Register(receiverID string, n Notifier) {
	s.notifiers[receiverID] = append(s.notifiers[receiverID], n)
}

// List returns all configured notifiers.
func (s *ReceiverStore) List() map[string][]Notifier {
	return s.notifiers
}

// getNotifiers retrieves the Notifiers for a given receiver ID.
func (s *ReceiverStore) getNotifiers(receiverID string) []Notifier {
	return s.notifiers[receiverID]
}

// InitializeStore builds a store from the receiver configuration.
func InitializeStore(cfg map[string]ReceiverConfig, globalSkipTLS bool, version string, logger *slog.Logger) *ReceiverStore {
	baseHeaders := map[string]string{
		"User-Agent":   "heartbeats/" + version,
		"Content-Type": "application/json",
	}

	store := NewReceiverStore()
	for id, rc := range cfg {
		for _, sc := range rc.SlackConfigs {
			effectiveSkipTLS := resolveSkipTLS(sc.SkipTLS, globalSkipTLS)
			store.Register(id,
				NewSlackNotifier(id, sc, logger,
					slack.NewWithToken(
						sc.Token,
						slack.WithHeaders(baseHeaders),
						slack.WithInsecureTLS(effectiveSkipTLS),
					),
				),
			)
		}
		for _, ec := range rc.EmailConfigs {
			effectiveSkipTLS := resolveSkipTLS(ec.SMTPConfig.SkipInsecureVerify, globalSkipTLS)
			ec.SMTPConfig.SkipInsecureVerify = &effectiveSkipTLS
			store.Register(id,
				NewEmailNotifier(id, ec, logger,
					email.New(ec.SMTPConfig)))
		}
		for _, tc := range rc.MSTeamsConfigs {
			effectiveSkipTLS := resolveSkipTLS(tc.SkipTLS, globalSkipTLS)
			store.Register(id,
				NewMSTeamsNotifier(id, tc, logger,
					msteams.New(
						msteams.WithHeaders(baseHeaders),
						msteams.WithInsecureTLS(effectiveSkipTLS),
					),
				),
			)
		}
		for _, gc := range rc.MSTeamsGraphConfig {
			effectiveSkipTLS := resolveSkipTLS(gc.SkipTLS, globalSkipTLS)
			store.Register(id,
				NewMSTeamsGraphNotifier(id, gc, logger,
					msteamsgraph.NewWithToken(
						gc.Token,
						msteamsgraph.WithHeaders(baseHeaders),
						msteamsgraph.WithInsecureTLS(effectiveSkipTLS),
					)),
			)
		}
	}
	return store
}
