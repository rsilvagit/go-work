package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/rsilvagit/go-work/internal/httpclient"
	"github.com/rsilvagit/go-work/internal/model"
)

type LinkedIn struct {
	client  *httpclient.Client
	baseURL string
}

func NewLinkedIn(client *httpclient.Client) *LinkedIn {
	return &LinkedIn{
		client:  client,
		baseURL: "https://www.linkedin.com/jobs/search",
	}
}

func (l *LinkedIn) Name() string {
	return "LinkedIn"
}

func (l *LinkedIn) Search(ctx context.Context, query string, location string) ([]model.Job, error) {
	params := url.Values{}
	params.Set("keywords", query)
	params.Set("location", location)
	searchURL := fmt.Sprintf("%s?%s", l.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("linkedin: building request: %w", err)
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linkedin: executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin: unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("linkedin: parsing HTML: %w", err)
	}

	// TODO: Atualizar seletores CSS conforme a estrutura real do LinkedIn.
	var jobs []model.Job
	doc.Find(".base-card").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".base-search-card__title").Text())
		company := strings.TrimSpace(s.Find(".base-search-card__subtitle").Text())
		loc := strings.TrimSpace(s.Find(".job-search-card__location").Text())
		link, _ := s.Find("a").Attr("href")

		if title != "" {
			jobs = append(jobs, model.Job{
				Title:    title,
				Company:  company,
				Location: loc,
				URL:      link,
				Source:   "linkedin",
			})
		}
	})

	return jobs, nil
}
