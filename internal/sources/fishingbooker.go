package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"itposao/internal/models"
)

const fishingBookerJobsURL = "https://apply.workable.com/api/v1/widget/accounts/fishingbooker?details=true"

// FishingBooker fetches job listings from FishingBooker's Workable careers board.
type FishingBooker struct {
	Client *http.Client
}

func (f *FishingBooker) Name() string { return "fishingbooker" }

type fishingBookerResponse struct {
	Jobs []fishingBookerJob `json:"jobs"`
}

type fishingBookerJob struct {
	Title     string `json:"title"`
	Shortcode string `json:"shortcode"`
	URL       string `json:"url"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

func (f *FishingBooker) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fishingBookerJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fishingbooker: unexpected status %s", resp.Status)
	}

	var body fishingBookerResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Jobs))
	for _, j := range body.Jobs {
		jobs = append(jobs, &models.Job{
			ExternalID: j.Shortcode,
			Title:      j.Title,
			Location:   j.City,
			URL:        j.URL,
			Source:     f.Name(),
		})
	}
	return jobs, nil
}
