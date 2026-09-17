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

const lanacoCareersURL = "https://www.lanaco.com/en/otvorene-pozicije"

// lanacoPositionRe matches one job card in LANACO's server-rendered careers
// page. There is no ATS behind it (no Greenhouse/Lever/etc fingerprint in the
// HTML or JS), so the listing is scraped directly the way capslock.go and
// authoritypartners.go do.
var lanacoPositionRe = regexp.MustCompile(`(?s)<a href="pozicija\?id=([^"]+)" class="open-position">(.*?)</a>`)

var (
	lanacoTitleRe = regexp.MustCompile(`<h3>([^<]*)</h3>`)
	lanacoLocRe   = regexp.MustCompile(`data-loc="[^"]*">([^<]*)</div>`)
)

// Lanaco fetches job listings from LANACO's careers page.
type Lanaco struct {
	Client *http.Client
}

func (l *Lanaco) Name() string { return "lanaco" }

func (l *Lanaco) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := l.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lanacoCareersURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lanaco: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := lanacoPositionRe.FindAllStringSubmatch(string(body), -1)
	jobs := make([]*models.Job, 0, len(matches))
	for _, m := range matches {
		id := m[1]
		chunk := m[2]

		titleMatch := lanacoTitleRe.FindStringSubmatch(chunk)
		if titleMatch == nil {
			continue
		}
		title := html.UnescapeString(strings.TrimSpace(titleMatch[1]))

		var locs []string
		for _, lm := range lanacoLocRe.FindAllStringSubmatch(chunk, -1) {
			locs = append(locs, html.UnescapeString(strings.TrimSpace(lm[1])))
		}

		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      title,
			Location:   strings.Join(locs, ", "),
			URL:        "https://www.lanaco.com/en/pozicija?id=" + id,
			Source:     l.Name(),
		})
	}
	return jobs, nil
}
