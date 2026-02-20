package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rsilvagit/go-work/internal/httpclient"
	"github.com/rsilvagit/go-work/internal/model"
)

const gupyAPIURL = "https://employability-portal.gupy.io/api/v1/jobs"
const gupyPageSize = 20

type gupyResponse struct {
	Data []gupyJob `json:"data"`
}

type gupyJob struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	CompanyID     int    `json:"companyId"`
	CareerPage    string `json:"careerPageName"`
	CareerPageURL string `json:"careerPageUrl"`
	JobURL        string `json:"jobUrl"`
	Type          string `json:"type"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	WorkplaceType string `json:"workplaceType"`
	IsRemoteWork  bool   `json:"isRemoteWork"`
	PublishedDate string `json:"publishedDate"`
}

type Gupy struct {
	client *httpclient.Client
}

func NewGupy(client *httpclient.Client) *Gupy {
	return &Gupy{client: client}
}

func (g *Gupy) Name() string {
	return "Gupy"
}

func (g *Gupy) Search(ctx context.Context, query string, location string) ([]model.Job, error) {
	params := url.Values{}
	params.Set("jobName", query)
	params.Set("limit", fmt.Sprintf("%d", gupyPageSize))
	params.Set("offset", "0")
	searchURL := fmt.Sprintf("%s?%s", gupyAPIURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gupy: building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gupy: executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gupy: unexpected status %d", resp.StatusCode)
	}

	var result gupyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gupy: decoding response: %w", err)
	}

	var jobs []model.Job
	for _, gj := range result.Data {
		loc := buildLocation(gj.City, gj.State, gj.Country)

		// Filtrar por localização se informada.
		if location != "" && !strings.Contains(strings.ToLower(loc), strings.ToLower(location)) {
			continue
		}

		posted, _ := time.Parse(time.RFC3339, gj.PublishedDate)

		jobs = append(jobs, model.Job{
			Title:     gj.Name,
			Company:   gj.CareerPage,
			Location:  loc,
			URL:       gj.JobURL,
			Source:    "gupy",
			PostedAt:  posted,
			WorkModel: mapWorkplaceType(gj.WorkplaceType, gj.IsRemoteWork),
			JobType:   mapJobType(gj.Type),
		})
	}

	return jobs, nil
}

func buildLocation(city, state, country string) string {
	parts := make([]string, 0, 3)
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if country != "" {
		parts = append(parts, country)
	}
	return strings.Join(parts, ", ")
}

func mapWorkplaceType(wt string, isRemote bool) string {
	switch strings.ToLower(wt) {
	case "remote":
		return "remoto"
	case "hybrid":
		return "hibrido"
	case "on-site":
		return "presencial"
	default:
		if isRemote {
			return "remoto"
		}
		return ""
	}
}

func mapJobType(t string) string {
	switch t {
	case "vacancy_type_effective":
		return "full-time"
	case "vacancy_type_internship":
		return "estagio"
	case "vacancy_type_temporary":
		return "part-time"
	case "vacancy_type_freelance":
		return "freelance"
	default:
		return ""
	}
}
