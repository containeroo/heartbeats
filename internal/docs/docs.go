package docs

var Documentation Docs
var Chapters []string = []string{"config", "api", "endpoints", "heartbeats", "notifications"}

// Cache is the struct for the cache
type Cache struct {
	MaxSize int `json:"maxSize"`
	Reduce  int `json:"reduce"`
}

// Docs is the struct for the documentation
type Docs struct {
	SiteRoot   string      `json:"siteRoot"`
	Cache      *Cache      `json:"cache"`
	Endpoints  []Endpoint  `json:"endpoints"`
	Examples   []Example   `json:"examples"`
	Heartbeats []Heartbeat `json:"heartbeats"`
	Services   []Service   `json:"services"`
	Defaults   []Default   `json:"defaults"`
}

// NewDocumentation creates a new documentation
func NewDocumentation(siteRoot string, cache *Cache) *Docs {
	d := Docs{
		SiteRoot: siteRoot,
		Cache:    cache,
	}
	d.endpoints()
	d.examples()
	d.heartbeats()
	d.services()
	d.defaults()

	return &d
}
