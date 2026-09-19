package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"itposao/internal/models"
)

const atlantbhJobsURL = "https://apply.workable.com/api/v1/widget/accounts/atlantbh?details=true"

// Atlantbh fetches job listings from Atlantbh's Workable careers board.
type Atlantbh struct {
	Client *http.Client
}

func (a *Atlantbh) Name() string { return "atlantbh" }

type atlantbhResponse struct {
	Jobs []atlantbhJob `json:"jobs"`
}

type atlantbhJob struct {
	Title       string `json:"title"`
	Shortcode   string `json:"shortcode"`
	URL         string `json:"url"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Description string `json:"description"`
}

func (a *Atlantbh) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, atlantbhJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("atlantbh: unexpected status %s", resp.Status)
	}

	var body atlantbhResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Jobs))
	for _, j := range body.Jobs {
		jobs = append(jobs, &models.Job{
			ExternalID:  j.Shortcode,
			Title:       j.Title,
			Location:    j.City,
			URL:         j.URL,
			Description: j.Description,
			Source:      a.Name(),
		})
	}
	return jobs, nil
}
