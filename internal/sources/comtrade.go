package sources

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"

	"itposao/internal/models"
)

const comtradeJobsURL = "https://www.comtrade.com/jobs/"

// comtradeCard matches one job card in Comtrade's WordPress job listing:
// <a class="job-content" href="{url}">
//
//	<p class="country">{location}</p>
//	<h3 class="link-after">{title}</h3>
//
// No JSON API is exposed for this listing (it's a plain WP archive page
// with server-side query-string filtering), so the rendered HTML is used.
var comtradeCard = regexp.MustCompile(
	`(?s)<a class="job-content" href="([^"]*)">\s*<p class="country">\s*([^<]*?)\s*</p>\s*<h3 class="link-after">([^<]*)</h3>`,
)

// Comtrade fetches job listings from Comtrade's WordPress careers page.
type Comtrade struct {
	Client *http.Client
}

func (c *Comtrade) Name() string { return "comtrade" }

func (c *Comtrade) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, comtradeJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("comtrade: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := comtradeCard.FindAllStringSubmatch(string(body), -1)
	jobs := make([]*models.Job, 0, len(matches))
	for _, m := range matches {
		jobURL, location, title := m[1], html.UnescapeString(m[2]), html.UnescapeString(m[3])

		segments := strings.Split(strings.Trim(jobURL, "/"), "/")
		id := title
		if len(segments) > 0 {
			id = segments[len(segments)-1]
		}

		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      title,
			Location:   location,
			URL:        jobURL,
			Source:     c.Name(),
		})
	}
	return jobs, nil
}
