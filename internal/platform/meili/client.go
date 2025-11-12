package meili

import (
	"fmt"
	"net/http"
	"time"

	meilisearch "github.com/meilisearch/meilisearch-go"
	"github.com/windy/caatsm-dashboard/config"
)

// NewClient initialises a Meilisearch service manager with sensible defaults.
func NewClient(cfg config.SearchConfig) (meilisearch.ServiceManager, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("meilisearch host is empty")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	opts := []meilisearch.Option{
		meilisearch.WithCustomClient(httpClient),
	}
	if cfg.APIKey != "" {
		opts = append(opts, meilisearch.WithAPIKey(cfg.APIKey))
	}

	manager := meilisearch.New(cfg.Host, opts...)
	return manager, nil
}
