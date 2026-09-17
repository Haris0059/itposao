package sources

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"

	"itposao/internal/models"
)

const authorityPartnersJobsURL = "https://authoritypartnersinc.catsone.com/careers/86212-General/"

// authorityPartnersRow matches one job row on the CATS ATS career portal's
// server-rendered listing table:
// <a class="table-row" href="/careers/86212/jobs/{id}-{slug}" ...>
//
//	<div class="data-cell title-cell">{title}</div>
//	<div class="data-cell" data-label="Location">{location}</div>
//
// The portal has no JSON API (its data endpoint uses React Router's
// turbo-stream wire format, not plain JSON), so the rendered HTML is the
// stable contract instead.
var authorityPartnersRow = regexp.MustCompile(
	`<a class="table-row" href="(/careers/\d+/jobs/(\d+)-[^"]*)"[^>]*><div class="data-cell title-cell">([^<]*)</div><div class="data-cell" data-label="Location">([^<]*)</div>`,
)

// AuthorityPartners fetches job listings from Authority Partners' CATS ATS
// careers page.
type AuthorityPartners struct {
	Client *http.Client
}

func (a *AuthorityPartners) Name() string { return "authoritypartners" }

func (a *AuthorityPartners) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authorityPartnersJobsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authoritypartners: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := authorityPartnersRow.FindAllStringSubmatch(string(body), -1)
	jobs := make([]*models.Job, 0, len(matches))
	for _, m := range matches {
		path, id, title, location := m[1], m[2], m[3], m[4]
		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      html.UnescapeString(title),
			Location:   html.UnescapeString(location),
			URL:        "https://authoritypartnersinc.catsone.com" + path,
			Source:     a.Name(),
		})
	}
	return jobs, nil
}
