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

const systemVerificationCareersURL = "https://career.systemverification.com/"

// systemVerificationCardRe matches one job card from the HubSpot-hosted
// careers page (a blog-style listing, not a dedicated ATS): each posting is
// a static <a class="blog-card"> block linking to its own page, with the
// title in an <h3> and zero or more location tags in <div class="card-tag">.
var (
	systemVerificationCardRe  = regexp.MustCompile(`(?s)<a href="([^"]+)" class="blog-card[^>]*>(.*?)</a>`)
	systemVerificationTagRe   = regexp.MustCompile(`card-tag">([^<]*)</div>`)
	systemVerificationTitleRe = regexp.MustCompile(`blog-title">([^<]*)</h3>`)
)

// SystemVerification fetches job listings from System Verification's
// HubSpot-hosted careers page. There is no ATS: postings are rendered
// server-side as a blog listing, so the HTML is scraped directly. The site
// sits behind Cloudflare bot management, which 403s the default Go
// User-Agent, so a browser-like one is set.
type SystemVerification struct {
	Client *http.Client
}

func (s *SystemVerification) Name() string { return "systemverification" }

func (s *SystemVerification) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, systemVerificationCareersURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("systemverification: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := systemVerificationCardRe.FindAllStringSubmatch(string(body), -1)
	jobs := make([]*models.Job, 0, len(matches))
	for _, m := range matches {
		jobURL := html.UnescapeString(m[1])
		chunk := m[2]

		titleMatch := systemVerificationTitleRe.FindStringSubmatch(chunk)
		if titleMatch == nil {
			continue
		}
		title := strings.TrimSpace(html.UnescapeString(titleMatch[1]))

		var locParts []string
		for _, t := range systemVerificationTagRe.FindAllStringSubmatch(chunk, -1) {
			locParts = append(locParts, strings.TrimSpace(html.UnescapeString(t[1])))
		}
		location := strings.Join(locParts, ", ")

		id := jobURL
		if u, err := url.Parse(jobURL); err == nil {
			id = strings.Trim(u.Path, "/")
		}

		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      title,
			Location:   location,
			URL:        jobURL,
			Source:     s.Name(),
		})
	}
	return jobs, nil
}
