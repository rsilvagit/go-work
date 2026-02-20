package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rsilvagit/go-work/internal/model"
)

type Gupy struct {
	client  *http.Client
	baseURL string
}

func NewGupy() *Gupy {
	return &Gupy{
		client:  &http.Client{},
		baseURL: "https://portal.gupy.io/job-search/term",
	}
}

func (g *Gupy) Name() string {
	return "Gupy"
}

func (g *Gupy) Search(ctx context.Context, query string, location string) ([]model.Job, error) {
	searchURL := fmt.Sprintf("%s=%s", g.baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gupy: building request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko)")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gupy: executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gupy: unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gupy: parsing HTML: %w", err)
	}

	// TODO: Atualizar seletores CSS conforme a estrutura real da Gupy.
	var jobs []model.Job
	doc.Find("[data-testid='job-list-item']").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("h2").Text())
		company := strings.TrimSpace(s.Find("[data-testid='company-name']").Text())
		loc := strings.TrimSpace(s.Find("[data-testid='job-location']").Text())
		link, _ := s.Find("a").Attr("href")

		if title != "" {
			jobs = append(jobs, model.Job{
				Title:    title,
				Company:  company,
				Location: loc,
				URL:      link,
				Source:   "gupy",
			})
		}
	})

	return jobs, nil
}
