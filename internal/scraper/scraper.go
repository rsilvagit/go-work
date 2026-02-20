package scraper

import (
	"context"

	"github.com/rsilvagit/go-work/internal/httpclient"
	"github.com/rsilvagit/go-work/internal/model"
)

// Scraper defines the contract every job site scraper must satisfy.
type Scraper interface {
	// Name returns a human-readable identifier for this scraper.
	Name() string

	// Search queries the job site and returns matching listings.
	Search(ctx context.Context, query string, location string) ([]model.Job, error)
}

// Registry returns all available scrapers using the shared HTTP client.
func Registry(client *httpclient.Client) []Scraper {
	return []Scraper{
		NewGupy(client),
	}
}
