package notifier

import (
	"log/slog"

	"github.com/containeroo/heartbeats/pkg/notify/email"
	"github.com/containeroo/heartbeats/pkg/notify/msteams"
	"github.com/containeroo/heartbeats/pkg/notify/slack"
)

// ReceiverStore holds a mapping from receiver ID to a slice of Notifiers.
type ReceiverStore struct {
	notifiers map[string][]Notifier // map of notifiers by receiver ID
}

// newStore creates an empty ReceiverStore.
func newStore() *ReceiverStore {
	return &ReceiverStore{notifiers: make(map[string][]Notifier)}
}

// addNotifier registers a Notifier for a given receiver ID.
func (s *ReceiverStore) addNotifier(receiverID string, n Notifier) {
	s.notifiers[receiverID] = append(s.notifiers[receiverID], n)
}

// getNotifiers retrieves the Notifiers for a given receiver ID.
func (s *ReceiverStore) getNotifiers(receiverID string) []Notifier {
	return s.notifiers[receiverID]
}

// InitializeStore builds a store from the receiver configuration.
func InitializeStore(cfg map[string]ReceiverConfig, globalSkipTLS bool, logger *slog.Logger) *ReceiverStore {
	store := newStore()
	for id, rc := range cfg {
		for _, sc := range rc.SlackConfigs {
			effectiveSkipTLS := resolveSkipTLS(sc.SkipTLS, globalSkipTLS)
			store.addNotifier(id, NewSlackNotifier(id, sc, logger, slack.NewWithToken(sc.Token, effectiveSkipTLS)))
		}
		for _, ec := range rc.EmailConfigs {
			effectiveSkipTLS := resolveSkipTLS(ec.SMTPConfig.SkipInsecureVerify, globalSkipTLS)
			ec.SMTPConfig.SkipInsecureVerify = &effectiveSkipTLS
			store.addNotifier(id, NewEmailNotifier(id, ec, logger, email.New(ec.SMTPConfig)))
		}
		for _, tc := range rc.MSTeamsConfigs {
			effectiveSkipTLS := resolveSkipTLS(tc.SkipTLS, globalSkipTLS)
			store.addNotifier(id, NewMSTeamsNotifier(id, tc, logger, msteams.New(nil, effectiveSkipTLS)))
		}
	}
	return store
}
