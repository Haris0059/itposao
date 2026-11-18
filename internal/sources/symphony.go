package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"itposao/internal/models"
)

// Symphony moved from Breezy HR to Lever's EU region; the global
// api.lever.co host returns "Document not found" for this slug.
const symphonyJobsURL = "https://api.eu.lever.co/v0/postings/symphony?mode=json"

// Symphony fetches job listings from Symphony's Lever careers board.
type Symphony struct {
	Client *http.Client
}

func (s *Symphony) Name() string { return "symphony" }

type symphonyJob struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	HostedURL   string `json:"hostedUrl"`
	CreatedAt   int64  `json:"createdAt"`
	Description string `json:"description"`
	Additional  string `json:"additional"`
	Categories  struct {
		Location string `json:"location"`
	} `json:"categories"`
	Lists []struct {
		Text    string `json:"text"`
		Content string `json:"content"`
	} `json:"lists"`
}

func (s *Symphony) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, symphonyJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("symphony: unexpected status %s", resp.Status)
	}

	var body []symphonyJob
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(body))
	for _, j := range body {
		// Lever splits a posting into intro, titled lists (<li> items), and
		// a closing section; stitch them back into one HTML description.
		var desc strings.Builder
		desc.WriteString(j.Description)
		for _, l := range j.Lists {
			fmt.Fprintf(&desc, "<h3>%s</h3><ul>%s</ul>", l.Text, l.Content)
		}
		desc.WriteString(j.Additional)

		jobs = append(jobs, &models.Job{
			ExternalID:  j.ID,
			Title:       j.Text,
			Location:    j.Categories.Location,
			URL:         j.HostedURL,
			Description: desc.String(),
			PostedAt:    time.UnixMilli(j.CreatedAt),
			Source:      s.Name(),
		})
	}
	return jobs, nil
}
