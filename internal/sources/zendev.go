package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const zendevJobsURL = "https://api.talentlyft.com/v2/public/zendev/jobs"

// ZenDev fetches job listings from ZenDev's TalentLyft board.
type ZenDev struct {
	Client *http.Client
}

func (z *ZenDev) Name() string { return "zendev" }

type zendevResponse struct {
	Results []zendevJob `json:"Results"`
}

type zendevJob struct {
	ID           int64  `json:"Id"`
	Title        string `json:"Title"`
	ShortlinkURL string `json:"ShortlinkUrl"`
	Country      string `json:"Country"`
	City         string `json:"City"`
}

func (z *ZenDev) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := z.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zendevJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zendev: unexpected status %s", resp.Status)
	}

	var body zendevResponse
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
			Source:     z.Name(),
		})
	}
	return jobs, nil
}

func (z *ZenDev) FetchDescription(ctx context.Context, j *models.Job) (string, error) {
	return jobPostingDescription(ctx, z.Client, j.URL)
}
