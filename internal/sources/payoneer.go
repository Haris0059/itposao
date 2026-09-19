package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"itposao/internal/models"
)

const payoneerJobsURL = "https://boards-api.greenhouse.io/v1/boards/payoneer/jobs?content=true"

// Payoneer fetches job listings from Payoneer's Greenhouse board API.
type Payoneer struct {
	Client *http.Client
}

func (p *Payoneer) Name() string { return "payoneer" }

type payoneerResponse struct {
	Jobs []payoneerJob `json:"jobs"`
}

type payoneerJob struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	AbsoluteURL    string `json:"absolute_url"`
	FirstPublished string `json:"first_published"`
	Content        string `json:"content"` // HTML, entity-escaped by Greenhouse
	Location       struct {
		Name string `json:"name"`
	} `json:"location"`
}

func (p *Payoneer) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, payoneerJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payoneer: unexpected status %s", resp.Status)
	}

	var body payoneerResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Jobs))
	for _, j := range body.Jobs {
		job := &models.Job{
			ExternalID:  strconv.FormatInt(j.ID, 10),
			Title:       j.Title,
			Location:    j.Location.Name,
			URL:         j.AbsoluteURL,
			Description: html.UnescapeString(j.Content),
			Source:      p.Name(),
		}
		if t, err := time.Parse(time.RFC3339, j.FirstPublished); err == nil {
			job.PostedAt = t
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
