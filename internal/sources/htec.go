package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"itposao/internal/models"
)

const htecJobsURL = "https://api.talentlyft.com/v2/public/htecgroup/jobs"

// HTEC fetches job listings from HTEC Group's TalentLyft board.
type HTEC struct {
	Client *http.Client
}

func (h *HTEC) Name() string { return "htec" }

type htecResponse struct {
	Results []htecJob `json:"Results"`
}

type htecJob struct {
	ID           int64  `json:"Id"`
	Title        string `json:"Title"`
	ShortlinkURL string `json:"ShortlinkUrl"`
	Country      string `json:"Country"`
	City         string `json:"City"`
}

func (h *HTEC) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, htecJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("htec: unexpected status %s", resp.Status)
	}

	var body htecResponse
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
			Source:     h.Name(),
		})
	}
	return jobs, nil
}
