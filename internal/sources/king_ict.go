package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const kingICTJobsURL = "https://api.talentlyft.com/v2/public/king-ict/jobs"

// KingICT fetches job listings from KING ICT's TalentLyft board.
type KingICT struct {
	Client *http.Client
}

func (k *KingICT) Name() string { return "king-ict" }

type kingICTResponse struct {
	Results []kingICTJob `json:"Results"`
}

type kingICTJob struct {
	ID           int64  `json:"Id"`
	Title        string `json:"Title"`
	ShortlinkURL string `json:"ShortlinkUrl"`
	Country      string `json:"Country"`
	City         string `json:"City"`
}

func (k *KingICT) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := k.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, kingICTJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("king-ict: unexpected status %s", resp.Status)
	}

	var body kingICTResponse
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
			Source:     k.Name(),
		})
	}
	return jobs, nil
}

func (k *KingICT) FetchDescription(ctx context.Context, j *models.Job) (string, error) {
	return jobPostingDescription(ctx, k.Client, j.URL)
}
