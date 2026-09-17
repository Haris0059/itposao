package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"itposao/internal/models"
)

const (
	infobipJobsURL    = "https://infobip.wd3.myworkdayjobs.com/wday/cxs/infobip/InfobipCareers/jobs"
	infobipJobBaseURL = "https://infobip.wd3.myworkdayjobs.com/en-US/InfobipCareers"
	infobipPageSize   = 20
)

// Infobip fetches job listings from Infobip's Workday careers site.
type Infobip struct {
	Client *http.Client
}

func (i *Infobip) Name() string { return "infobip" }

type infobipRequest struct {
	AppliedFacets struct{} `json:"appliedFacets"`
	Limit         int      `json:"limit"`
	Offset        int      `json:"offset"`
	SearchText    string   `json:"searchText"`
}

type infobipResponse struct {
	Total       int          `json:"total"`
	JobPostings []infobipJob `json:"jobPostings"`
}

type infobipJob struct {
	Title         string   `json:"title"`
	ExternalPath  string   `json:"externalPath"`
	LocationsText string   `json:"locationsText"`
	BulletFields  []string `json:"bulletFields"` // first entry is the requisition ID, e.g. "JR104897"
}

func (i *Infobip) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := i.Client
	if client == nil {
		client = http.DefaultClient
	}

	var jobs []*models.Job
	offset := 0
	total := -1 // unknown until the first page comes back
	for {
		page, pageTotal, err := i.fetchPage(ctx, client, offset)
		if err != nil {
			return nil, err
		}
		if total == -1 {
			total = pageTotal
		}
		for _, r := range page {
			externalID := r.ExternalPath
			if len(r.BulletFields) > 0 {
				externalID = r.BulletFields[0]
			}
			jobs = append(jobs, &models.Job{
				ExternalID: externalID,
				Title:      r.Title,
				Location:   r.LocationsText,
				URL:        infobipJobBaseURL + r.ExternalPath,
				Source:     i.Name(),
			})
		}

		offset += infobipPageSize
		if offset >= total || len(page) == 0 {
			break
		}
	}
	return jobs, nil
}

func (i *Infobip) fetchPage(ctx context.Context, client *http.Client, offset int) ([]infobipJob, int, error) {
	body, err := json.Marshal(infobipRequest{Limit: infobipPageSize, Offset: offset})
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, infobipJobsURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("infobip: unexpected status %s", resp.Status)
	}

	var out infobipResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, 0, err
	}
	return out.JobPostings, out.Total, nil
}
