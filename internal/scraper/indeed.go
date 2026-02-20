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

type Indeed struct {
	client  *httpclient.Client
	baseURL string
}

func NewIndeed(client *httpclient.Client) *Indeed {
	return &Indeed{
		client:  client,
		baseURL: "https://www.indeed.com/jobs",
	}
}

func (in *Indeed) Name() string {
	return "Indeed"
}

func (in *Indeed) Search(ctx context.Context, query string, location string) ([]model.Job, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("l", location)
	searchURL := fmt.Sprintf("%s?%s", in.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("indeed: building request: %w", err)
	}
	resp, err := in.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("indeed: executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("indeed: unexpected status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("indeed: parsing HTML: %w", err)
	}

	// TODO: Atualizar seletores CSS conforme a estrutura real do Indeed.
	var jobs []model.Job
	doc.Find(".job_seen_beacon").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".jobTitle span").Text())
		company := strings.TrimSpace(s.Find(".companyName").Text())
		loc := strings.TrimSpace(s.Find(".companyLocation").Text())
		link, exists := s.Find("a").Attr("href")
		if exists && !strings.HasPrefix(link, "http") {
			link = "https://www.indeed.com" + link
		}

		if title != "" {
			jobs = append(jobs, model.Job{
				Title:    title,
				Company:  company,
				Location: loc,
				URL:      link,
				Source:   "indeed",
			})
		}
	})

	return jobs, nil
}
