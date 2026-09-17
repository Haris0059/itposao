package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"itposao/internal/models"
)

const klikaJobsURL = "https://albatross.recruitcrm.io/v1/external-pages/jobs-by-account/get?account=Klika_1_jobs&batch=true"

// Klika fetches job listings from Klika's RecruitCRM careers board.
type Klika struct {
	Client *http.Client
}

func (k *Klika) Name() string { return "klika" }

type klikaResponse struct {
	Status string `json:"status"`
	Data   struct {
		Jobs []klikaJob `json:"jobs"`
	} `json:"data"`
}

type klikaJob struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	City string `json:"city"`
}

func (k *Klika) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := k.Client
	if client == nil {
		client = http.DefaultClient
	}

	reqBody, err := json.Marshal(map[string]any{
		"limit": 100, "offset": 0, "search_data": "", "onlyJobs": true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, klikaJobsURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// The API rejects requests without these; it's embedded as an iframe on
	// klika.ba/careers but only accepts recruitcrm.io as the caller.
	req.Header.Set("Referer", "https://recruitcrm.io/")
	req.Header.Set("Origin", "https://recruitcrm.io")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("klika: unexpected status %s", resp.Status)
	}

	var body klikaResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Status != "success" {
		return nil, fmt.Errorf("klika: API returned status %q", body.Status)
	}

	jobs := make([]*models.Job, 0, len(body.Data.Jobs))
	for _, j := range body.Data.Jobs {
		jobs = append(jobs, &models.Job{
			ExternalID: j.Slug,
			Title:      j.Name,
			Location:   j.City,
			URL:        "https://recruitcrm.io/apply/" + j.Slug,
			Source:     k.Name(),
		})
	}
	return jobs, nil
}
