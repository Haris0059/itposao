package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"itposao/internal/models"
)

const allopsJobsAPIURL = "https://allops.co/wp-json/wp/v2/job?per_page=100"

// allopsLocationRe pulls the location text out of a job detail page. The
// WordPress REST API for the "job" custom post type doesn't expose location
// as a field, so each listing's own page is fetched to scrape it out of the
// rendered "job-hero__location" div.
var allopsLocationRe = regexp.MustCompile(`job-hero__location"[^>]*>([^<]*)</div>`)

type allopsJob struct {
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Link  string `json:"link"`
	Slug  string `json:"slug"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
}

// AllOps fetches open positions from allOps Solutions' WordPress-based
// careers site (a custom "job" post type exposed over the WP REST API).
type AllOps struct {
	Client *http.Client
}

func (a *AllOps) Name() string { return "allops" }

func (a *AllOps) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}

	listing, err := a.fetchListing(ctx, client)
	if err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0, len(listing))
	for _, lj := range listing {
		title := html.UnescapeString(strings.TrimSpace(lj.Title.Rendered))
		location := a.fetchLocation(ctx, client, lj.Link)

		var postedAt time.Time
		if t, err := time.Parse("2006-01-02T15:04:05", lj.Date); err == nil {
			postedAt = t
		}

		jobs = append(jobs, &models.Job{
			ExternalID: strconv.Itoa(lj.ID),
			Title:      title,
			Location:   location,
			URL:        lj.Link,
			Source:     a.Name(),
			PostedAt:   postedAt,
		})
	}
	return jobs, nil
}

func (a *AllOps) fetchListing(ctx context.Context, client *http.Client) ([]allopsJob, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, allopsJobsAPIURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("allops: unexpected status %s", resp.Status)
	}

	var listing []allopsJob
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, err
	}
	return listing, nil
}

// fetchLocation best-effort scrapes the location off a job's own page. A
// failure here (network hiccup, page layout change) just leaves the
// location blank rather than failing the whole fetch.
func (a *AllOps) fetchLocation(ctx context.Context, client *http.Client, jobURL string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobURL, nil)
	if err != nil {
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	m := allopsLocationRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(string(m[1])))
}
