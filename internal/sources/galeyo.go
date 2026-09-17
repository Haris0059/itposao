package sources

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"itposao/internal/models"
)

const galeyoOpenPositionsURL = "https://www.galeyo.com/open-positions"

// galeyoItemDelim splits the server-rendered open-positions page (a Webflow
// CMS collection) into one chunk per listing. Webflow renders the list
// server-side, so no ATS/JS endpoint is involved — plain HTML scraping.
const galeyoItemDelim = `class="collection-item-4 w-dyn-item">`

var (
	galeyoHrefRe     = regexp.MustCompile(`href="([^"]*)"`)
	galeyoHeadingRe  = regexp.MustCompile(`open-position-list-item-heading">([^<]*)</div>`)
	galeyoLocationRe = regexp.MustCompile(`job-location">([^<]*)</div>`)
)

// Galeyo fetches job listings from Galeyo's (formerly Walter Code) Webflow
// careers page.
type Galeyo struct {
	Client *http.Client
}

func (g *Galeyo) Name() string { return "galeyo" }

func (g *Galeyo) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, galeyoOpenPositionsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("galeyo: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	chunks := strings.Split(string(body), galeyoItemDelim)[1:]
	jobs := make([]*models.Job, 0, len(chunks))
	for _, chunk := range chunks {
		hrefMatch := galeyoHrefRe.FindStringSubmatch(chunk)
		if hrefMatch == nil {
			continue
		}
		jobURL := hrefMatch[1]
		if u, err := url.Parse(jobURL); err == nil {
			jobURL = (&url.URL{Scheme: "https", Host: "www.galeyo.com", Path: u.Path}).String()
		}

		headingMatch := galeyoHeadingRe.FindStringSubmatch(chunk)
		if headingMatch == nil {
			continue
		}
		title := html.UnescapeString(strings.TrimSpace(headingMatch[1]))

		location := ""
		if locMatch := galeyoLocationRe.FindStringSubmatch(chunk); locMatch != nil {
			location = html.UnescapeString(strings.TrimSpace(locMatch[1]))
		}

		id := jobURL
		if u, err := url.Parse(jobURL); err == nil {
			segments := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(segments) > 0 {
				id = segments[len(segments)-1]
			}
		}

		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      title,
			Location:   location,
			URL:        jobURL,
			Source:     g.Name(),
		})
	}
	return jobs, nil
}
