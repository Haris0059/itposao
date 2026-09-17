package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"itposao/internal/models"
)

const (
	alemSistemJobsURL = "https://alem-sistem.breezy.hr/json"
)

// AlemSistem fetches job listings from Alem Sistem's Breezy HR careers board.
type AlemSistem struct {
	Client *http.Client
}

func (a *AlemSistem) Name() string { return "alemsistem" }

type alemSistemJob struct {
	ID          string `json:"id"`
	FriendlyID  string `json:"friendly_id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	PublishedAt string `json:"published_date"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
}

func (a *AlemSistem) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, alemSistemJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("alemsistem: unexpected status %s", resp.Status)
	}

	var body []alemSistemJob
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body))
	for _, j := range body {
		job := &models.Job{
			ExternalID: j.ID,
			Title:      j.Name,
			Location:   j.Location.Name,
			URL:        j.URL,
			Source:     a.Name(),
		}
		if t, err := time.Parse(time.RFC3339, j.PublishedAt); err == nil {
			job.PostedAt = t
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
