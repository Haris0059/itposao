package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const mistralJobsURL = "https://api.talentlyft.com/v2/public/mistral/jobs"

// Mistral fetches job listings from Mistral's TalentLyft board.
type Mistral struct {
	Client *http.Client
}

func (m *Mistral) Name() string { return "mistral" }

type mistralResponse struct {
	Results []mistralJob `json:"Results"`
}

type mistralJob struct {
	ID           int64  `json:"Id"`
	Title        string `json:"Title"`
	ShortlinkURL string `json:"ShortlinkUrl"`
	Country      string `json:"Country"`
	City         string `json:"City"`
}

func (m *Mistral) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := m.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mistralJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mistral: unexpected status %s", resp.Status)
	}

	var body mistralResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body.Results))
	for _, r := range body.Results {
		jobs = append(jobs, &models.Job{
			ExternalID: strconv.FormatInt(r.ID, 10),
			Title:      r.Title,
			Location:   r.City,
			URL:        r.ShortlinkURL,
			Source:     m.Name(),
		})
	}
	return jobs, nil
}
