package model

import (
	"strings"
	"time"
)

// Job represents a single job listing scraped from any source.
type Job struct {
	Title       string
	Company     string
	Location    string
	URL         string
	Description string
	Source      string
	PostedAt    time.Time
	JobType     string // full-time, part-time, estagio, freelance
	WorkModel   string // remoto, hibrido, presencial
	Level       string // junior, pleno, senior
	Salary      string // texto livre ex: "R$ 5.000 - R$ 8.000"
}

// FullText returns all searchable text fields concatenated in lowercase.
func (j Job) FullText() string {
	return strings.ToLower(
		j.Title + " " + j.Description + " " + j.JobType + " " +
			j.WorkModel + " " + j.Level + " " + j.Location + " " + j.Salary,
	)
}

// Key returns a deduplication key for this job.
// Uses URL when available, otherwise falls back to title+company.
func (j Job) Key() string {
	if j.URL != "" {
		return strings.ToLower(j.URL)
	}
	return strings.ToLower(j.Title + "|" + j.Company)
}
