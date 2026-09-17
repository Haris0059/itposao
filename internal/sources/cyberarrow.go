package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const cyberarrowJobsURL = "https://career.recruitee.com/api/c/109963/widget/?widget=true"

// CyberArrow fetches job listings from CyberArrow's Recruitee careers board.
type CyberArrow struct {
	Client *http.Client
}

func (c *CyberArrow) Name() string { return "cyberarrow" }

type cyberarrowResponse struct {
	Offers []cyberarrowJob `json:"offers"`
}

type cyberarrowJob struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Location   string `json:"location"`
	City       string `json:"city"`
	Country    string `json:"country"`
	CareersURL string `json:"careers_url"`
}

func (c *CyberArrow) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cyberarrowJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cyberarrow: unexpected status %s", resp.Status)
	}

	var body cyberarrowResponse
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
			Source:     c.Name(),
		})
	}
	return jobs, nil
}
