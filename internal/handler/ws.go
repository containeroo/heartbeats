package handler

import "net/http"

// WS streams heartbeat updates over a websocket connection.
func (a *API) WS() http.HandlerFunc {
	if a.wsHub == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "websocket hub disabled", http.StatusServiceUnavailable)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		a.wsHub.Handle(w, r)
	}
}
