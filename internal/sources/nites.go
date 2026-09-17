package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"itposao/internal/models"
)

// nitesJobsURL is ENIT's (now branded NITES, nites.ba redirects to nites.eu)
// OneAssessment career board API, found in the AngularJS bundle served from
// the careers page (/build/assets/career-*.js references "/api/career/positions").
const nitesJobsURL = "https://nites.oneassessment.com/api/career/positions"

// Nites fetches job listings from NITES's (formerly ENIT) OneAssessment career board.
type Nites struct {
	Client *http.Client
}

func (e *Nites) Name() string { return "nites" }

type nitesResponse struct {
	Data []nitesJob `json:"data"`
}

type nitesJob struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	CreatedAt string            `json:"created_at"`
	Locations []nitesJobLocation `json:"locations"`
}

type nitesJobLocation struct {
	Name string `json:"name"`
}

func (e *Nites) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := e.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nitesJobsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nites: unexpected status %s", resp.Status)
	}

	var body nitesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Data))
	for _, j := range body.Data {
		locs := make([]string, 0, len(j.Locations))
		for _, l := range j.Locations {
			if l.Name != "" {
				locs = append(locs, l.Name)
			}
		}

		job := &models.Job{
			ExternalID: j.ID,
			Title:      strings.TrimSpace(j.Name),
			Location:   strings.Join(locs, ", "),
			URL:        j.URL,
			Source:     e.Name(),
		}
		if t, err := time.Parse("2006-01-02 15:04:05", j.CreatedAt); err == nil {
			job.PostedAt = t
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
