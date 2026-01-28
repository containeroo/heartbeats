package handler

import "net/http"

// configResponse describes runtime settings exposed to the SPA.
type configResponse struct {
	Version string `json:"version"` // Version is the build version string.
	Commit  string `json:"commit"`  // Commit is the build commit SHA.
	SiteURL string `json:"siteUrl"` // SiteURL is the site root URL.
}

// Config returns runtime configuration for the SPA.
func (a *API) Config() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, configResponse{
			Version: a.Version,
			Commit:  a.Commit,
			SiteURL: a.SiteURL,
		})
	}
}
