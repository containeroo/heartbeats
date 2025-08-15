package server

import "net/http"

// mountUnderPrefix mounts h under the given route prefix, adding a redirect from bare prefix → prefix/.
func mountUnderPrefix(h http.Handler, prefix string) http.Handler {
	if prefix == "" {
		return h // serve at root
	}
	mux := http.NewServeMux()

	// Redirect bare "/tiledash" to "/tiledash/" so subtree handlers match.
	mux.HandleFunc("GET "+prefix, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
	})

	// Mount everything under prefix and strip it so internal routes live at "/".
	mux.Handle(prefix+"/", http.StripPrefix(prefix, h))

	// Not mounting at "/" ensures non-prefixed URLs 404, which is desirable when hosting under a subpath.
	return mux
}
