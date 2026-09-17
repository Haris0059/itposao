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

const capslockCareersURL = "https://capslock.global/careers/"

// capslockItemDelim splits the server-rendered careers page (a Comeet
// embed) into one chunk per job card. Comeet has a documented JSON API,
// but it requires a per-company token that isn't exposed anywhere on this
// page, so the rendered HTML is used instead — same as authoritypartners.go.
const capslockItemDelim = `<div class="item" data-title="`

var (
	capslockHrefRe = regexp.MustCompile(`href="([^"]*)"`)
	capslockTagRe  = regexp.MustCompile(`tag__title"[^>]*>([^<]*)</span>`)
)

// CapsLock fetches job listings from CapsLock's Comeet careers page.
type CapsLock struct {
	Client *http.Client
}

func (c *CapsLock) Name() string { return "capslock" }

func (c *CapsLock) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, capslockCareersURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("capslock: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	chunks := strings.Split(string(body), capslockItemDelim)[1:]
	jobs := make([]*models.Job, 0, len(chunks))
	for _, chunk := range chunks {
		titleEnd := strings.IndexByte(chunk, '"')
		if titleEnd < 0 {
			continue
		}
		title := html.UnescapeString(chunk[:titleEnd])

		hrefMatch := capslockHrefRe.FindStringSubmatch(chunk)
		if hrefMatch == nil {
			continue
		}
		jobURL := hrefMatch[1]

		id := jobURL
		if u, err := url.Parse(jobURL); err == nil {
			segments := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(segments) > 0 {
				id = segments[len(segments)-1]
			}
		}

		// The rendered card shows two tags: department, then region. Region
		// is the closest thing to a location this page exposes.
		location := ""
		if tags := capslockTagRe.FindAllStringSubmatch(chunk, -1); len(tags) > 0 {
			location = html.UnescapeString(tags[len(tags)-1][1])
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
