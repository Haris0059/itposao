package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"itposao/internal/models"
)

const (
	symphonyJobsURL = "https://symphony.breezy.hr/json"
	symphonyBaseURL = "https://symphony.breezy.hr/p/"
)

// Symphony fetches job listings from Symphony's Breezy HR careers board.
type Symphony struct {
	Client *http.Client
}

func (s *Symphony) Name() string { return "symphony" }

type symphonyJob struct {
	ID         string `json:"_id"`
	FriendlyID string `json:"friendly_id"`
	Name       string `json:"name"`
	Location   struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
}

func (s *Symphony) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, symphonyJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("symphony: unexpected status %s", resp.Status)
	}

	var body []symphonyJob
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body))
	for _, j := range body {
		jobs = append(jobs, &models.Job{
			ExternalID: j.ID,
			Title:      j.Name,
			Location:   j.Location.City,
			URL:        symphonyBaseURL + j.FriendlyID,
			Source:     s.Name(),
		})
	}
	return jobs, nil
}
