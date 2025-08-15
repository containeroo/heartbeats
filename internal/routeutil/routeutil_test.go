package routeutil

import "testing"

func TestNormalizeRoutePrefix(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if got := NormalizeRoutePrefix(""); got != "" {
			t.Fatalf("want '', got %q", got)
		}
	})
	t.Run("slash", func(t *testing.T) {
		t.Parallel()
		if got := NormalizeRoutePrefix("/"); got != "" {
			t.Fatalf("want '', got %q", got)
		}
	})
	t.Run("no leading slash", func(t *testing.T) {
		t.Parallel()
		if got := NormalizeRoutePrefix("tiledash"); got != "/tiledash" {
			t.Fatalf("want '/tiledash', got %q", got)
		}
	})
	t.Run("trailing slash", func(t *testing.T) {
		t.Parallel()
		if got := NormalizeRoutePrefix("/tiledash/"); got != "/tiledash" {
			t.Fatalf("want '/tiledash', got %q", got)
		}
	})
	t.Run("full url", func(t *testing.T) {
		t.Parallel()
		if got := NormalizeRoutePrefix("https://x/y/z/"); got != "/y/z" {
			t.Fatalf("want '/y/z', got %q", got)
		}
	})
}
