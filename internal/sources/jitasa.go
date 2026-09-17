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

const jitasaCareersURL = "https://www.jitasagroup.com/about/careers/"

// jitasaJobRe matches the open-positions list on the careers page: each
// role is a link to a static PDF posting, followed by a span with the
// location. There's no ATS or JSON API — just this hand-written list.
var jitasaJobRe = regexp.MustCompile(`<a href="(/files/careers/([^"/]+)\.pdf)">([^<]*)</a>\s*<span[^>]*>\s*-?\s*([^<]*)</span>`)

// Jitasa scrapes the static open-positions list on Jitasa Group's careers
// page. Postings are plain PDF links (no ATS), so the page is scraped
// directly with regexp instead of hitting a JSON API.
type Jitasa struct {
	Client *http.Client
}

func (j *Jitasa) Name() string { return "jitasa" }

func (j *Jitasa) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := j.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jitasaCareersURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jitasa: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := jitasaJobRe.FindAllStringSubmatch(string(body), -1)
	jobs := make([]*models.Job, 0, len(matches))
	for _, m := range matches {
		path, id, title, location := m[1], m[2], m[3], m[4]

		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      strings.TrimSpace(html.UnescapeString(title)),
			Location:   strings.TrimSpace(html.UnescapeString(location)),
			URL:        "https://www.jitasagroup.com" + path,
			Source:     j.Name(),
		})
	}
	return jobs, nil
}
