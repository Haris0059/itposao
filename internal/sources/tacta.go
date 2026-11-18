package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const tactaJobsURL = "https://api.talentlyft.com/v2/public/tacta-bv/jobs"

// Tacta fetches job listings from Tacta's TalentLyft board.
type Tacta struct {
	Client *http.Client
}

func (t *Tacta) Name() string { return "tacta" }

type tactaResponse struct {
	Results []tactaJob `json:"Results"`
}

type tactaJob struct {
	ID           int64  `json:"Id"`
	Title        string `json:"Title"`
	ShortlinkURL string `json:"ShortlinkUrl"`
	Country      string `json:"Country"`
	City         string `json:"City"`
}

func (t *Tacta) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := t.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tactaJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tacta: unexpected status %s", resp.Status)
	}

	var body tactaResponse
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
			Source:     t.Name(),
		})
	}
	return jobs, nil
}

func (t *Tacta) FetchDescription(ctx context.Context, j *models.Job) (string, error) {
	return jobPostingDescription(ctx, t.Client, j.URL)
}
