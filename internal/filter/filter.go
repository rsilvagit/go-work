package filter

import (
	"strings"

	"github.com/rsilvagit/go-work/internal/model"
)

// Options holds all filter criteria. Empty fields mean "no filter".
type Options struct {
	JobType   string // full-time, part-time, estagio, freelance
	WorkModel string // remoto, hibrido, presencial
	Level     string // junior, pleno, senior
	Region    string // text to match against Location
}

// Apply filters a slice of jobs, returning only those that match all criteria.
func Apply(jobs []model.Job, opts Options) []model.Job {
	if opts.isEmpty() {
		return jobs
	}

	var result []model.Job
	for _, j := range jobs {
		if matchJob(j, opts) {
			result = append(result, j)
		}
	}
	return result
}

func matchJob(j model.Job, opts Options) bool {
	text := j.FullText()

	if opts.JobType != "" && !containsAny(text, opts.JobType) {
		return false
	}
	if opts.WorkModel != "" && !containsAny(text, opts.WorkModel) {
		return false
	}
	if opts.Level != "" && !containsAny(text, opts.Level) {
		return false
	}
	if opts.Region != "" && !containsAny(text, opts.Region) {
		return false
	}
	return true
}

// containsAny checks if text contains any of the comma-separated terms.
func containsAny(text, terms string) bool {
	for _, term := range strings.Split(terms, ",") {
		term = strings.TrimSpace(strings.ToLower(term))
		if term != "" && strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func (o Options) isEmpty() bool {
	return o.JobType == "" && o.WorkModel == "" && o.Level == "" && o.Region == ""
}
