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

const qssCareersURL = "https://qss.ba/company/careers"

// qssItemDelim splits the server-rendered careers page (a Weblium site
// builder page, no ATS or JSON API found) into one chunk per job card.
// Each card's title lives ahead of its "More about vacancy" button, so
// items are split on that shared marker class instead.
const qssItemDelim = `item-inner__content__`

var (
	qssTitleRe = regexp.MustCompile(`ui-heading[^"]*"[^>]*>.*?w-text-content">(?:<span[^>]*>)?([^<]+)`)
	qssHrefRe  = regexp.MustCompile(`href="(/company/jobs[^"]*)"`)
	qssTextRe  = regexp.MustCompile(`w-text-content">([^<]+)</span>`)
)

// QSS fetches job listings from QSS's careers page. The page shows a
// "More about vacancy" button for every card, but unpublished vacancies
// link it to "/" instead of a real posting URL — those are skipped since
// there is no public link to give out.
type QSS struct {
	Client *http.Client
}

func (q *QSS) Name() string { return "qss" }

func (q *QSS) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := q.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, qssCareersURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qss: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	chunks := strings.Split(string(body), qssItemDelim)[1:]
	jobs := make([]*models.Job, 0, len(chunks))
	for _, chunk := range chunks {
		titleMatch := qssTitleRe.FindStringSubmatch(chunk)
		if titleMatch == nil {
			continue
		}
		title := html.UnescapeString(strings.TrimSpace(titleMatch[1]))

		hrefMatch := qssHrefRe.FindStringSubmatch(chunk)
		if hrefMatch == nil {
			// Unpublished vacancy card: no real posting URL to give out.
			continue
		}
		jobURL := "https://qss.ba" + hrefMatch[1]
		id := strings.TrimPrefix(hrefMatch[1], "/company/jobs-")

		// The card lists two tags in a fixed order: employment type, then
		// location. Location is the last one.
		location := ""
		if texts := qssTextRe.FindAllStringSubmatch(chunk, -1); len(texts) > 0 {
			location = html.UnescapeString(strings.TrimSpace(texts[len(texts)-1][1]))
		}

		jobs = append(jobs, &models.Job{
			ExternalID: id,
			Title:      title,
			Location:   location,
			URL:        jobURL,
			Source:     q.Name(),
		})
	}
	return jobs, nil
}
