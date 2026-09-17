package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"itposao/internal/models"
)

const rubiconJobsURL = "https://rubicon.bamboohr.com/careers/list?version=2"

// Rubicon fetches job listings from RUBICON's BambooHR careers board.
type Rubicon struct {
	Client *http.Client
}

func (r *Rubicon) Name() string { return "rubicon" }

type rubiconResponse struct {
	Result []rubiconJob `json:"result"`
}

type rubiconJob struct {
	ID       string `json:"id"`
	Name     string `json:"jobOpeningName"`
	Location struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"location"`
}

func (r *Rubicon) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rubiconJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rubicon: unexpected status %s", resp.Status)
	}

	var body rubiconResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Result))
	for _, j := range body.Result {
		jobs = append(jobs, &models.Job{
			ExternalID: j.ID,
			Title:      j.Name,
			Location:   j.Location.City,
			URL:        "https://rubicon.bamboohr.com/careers/" + j.ID,
			Source:     r.Name(),
		})
	}
	return jobs, nil
}
