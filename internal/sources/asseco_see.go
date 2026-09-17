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

const assecoSEECareersURL = "https://asee.peopleforce.io/careers"

// assecoSEEMaxPages caps pagination as a safety net. The site serves ~10
// jobs per page and silently clamps out-of-range page numbers to the last
// page (re-serving the same jobs), so pagination stops as soon as a page
// contributes no new job IDs.
const assecoSEEMaxPages = 20

// assecoSEEJobRe matches one job card link: href, then title text.
var assecoSEEJobRe = regexp.MustCompile(`<a class="stretched-link tw-text-black" data-turbo-frame="_top" href="([^"]+)">([^<]*)</a>`)

// assecoSEELocationRe matches the location line (marked by a map-pin icon)
// that follows a job card's title in the same card block.
var assecoSEELocationRe = regexp.MustCompile(`fa-map-marker-alt[^>]*></i>([\s\S]*?)</div>`)

// AssecoSEE fetches job listings from Asseco SEE's PeopleForce careers
// board (asee.peopleforce.io). The board is server-rendered (Rails +
// Turbo), so plain HTML scraping is used instead of PeopleForce's JS
// widget API, which isn't exposed on this page.
type AssecoSEE struct {
	Client *http.Client
}

func (a *AssecoSEE) Name() string { return "asseco-see" }

func (a *AssecoSEE) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}

	seen := make(map[string]bool)
	var jobs []*models.Job

	for page := 1; page <= assecoSEEMaxPages; page++ {
		body, err := assecoSEEFetchPage(ctx, client, page)
		if err != nil {
			return nil, err
		}

		newOnPage := 0
		matches := assecoSEEJobRe.FindAllStringSubmatchIndex(body, -1)
		for i, m := range matches {
			href := body[m[2]:m[3]]
			title := html.UnescapeString(strings.TrimSpace(body[m[4]:m[5]]))

			id := strings.TrimPrefix(href, "/careers/v/")
			if seen[id] {
				continue
			}
			seen[id] = true
			newOnPage++

			// Search for the location within this card's slice of the
			// body: from the end of this match to the start of the next
			// one (or end of string for the last card on the page).
			end := len(body)
			if i+1 < len(matches) {
				end = matches[i+1][0]
			}
			location := ""
			if locMatch := assecoSEELocationRe.FindStringSubmatch(body[m[1]:end]); locMatch != nil {
				location = html.UnescapeString(strings.Join(strings.Fields(locMatch[1]), " "))
			}

			jobs = append(jobs, &models.Job{
				ExternalID: id,
				Title:      title,
				Location:   location,
				URL:        assecoSEECareersURL + "/v/" + id,
				Source:     a.Name(),
			})
		}

		if newOnPage == 0 {
			break
		}
	}

	return jobs, nil
}

func assecoSEEFetchPage(ctx context.Context, client *http.Client, page int) (string, error) {
	url := fmt.Sprintf("%s?page=%d", assecoSEECareersURL, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("asseco-see: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
