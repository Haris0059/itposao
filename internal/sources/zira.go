package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const ziraJobsURL = "https://zira.recruitee.com/api/offers/"

// Zira fetches job listings from ZIRA Group's Recruitee careers board.
type Zira struct {
	Client *http.Client
}

func (z *Zira) Name() string { return "zira" }

type ziraResponse struct {
	Offers []ziraJob `json:"offers"`
}

type ziraJob struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Location   string `json:"location"`
	City       string `json:"city"`
	Country    string `json:"country"`
	CareersURL string `json:"careers_url"`
}

func (z *Zira) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := z.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ziraJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zira: unexpected status %s", resp.Status)
	}

	var body ziraResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Offers))
	for _, o := range body.Offers {
		location := o.City
		if location == "" {
			location = o.Location
		}
		if o.Country != "" && o.Country != location {
			if location != "" {
				location += ", "
			}
			location += o.Country
		}

		jobs = append(jobs, &models.Job{
			ExternalID: strconv.FormatInt(o.ID, 10),
			Title:      o.Title,
			Location:   location,
			URL:        o.CareersURL,
			Source:     z.Name(),
		})
	}
	return jobs, nil
}
