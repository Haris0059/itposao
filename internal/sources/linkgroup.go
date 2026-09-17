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

const linkGroupCareersURL = "https://www.link-group.eu/currently-open-positions"

// linkGroupContentStart/End bound the hand-edited CMS content block on the
// page, so parsing doesn't wander into nav/footer markup that also contains
// "<h2>" and "<a href=" tags.
const (
	linkGroupContentStart = `<div id="node-275"`
	linkGroupContentEnd   = `<div id="after-content"`
)

// linkGroupSectionRe splits the content block into one chunk per "<h2>
// Country</h2> ... " section. There's no ATS here: this is a single static
// page manually maintained per-country in a Drupal CMS, with job links as
// plain <a> tags inside <p> tags (or "There are currently no open
// positions." text when a country has none).
var linkGroupSectionRe = regexp.MustCompile(`<h2>([^<]*)</h2>`)

// linkGroupLinkRe matches one job link within a country section. Link text
// usually ends with a literal "&gt;&gt;" (rendered ">>"); the page also has
// a stray second anchor on one entry that repeats ">>" pointing at the
// wrong href, which is filtered out below because its title is empty once
// the ">>" suffix is stripped.
var linkGroupLinkRe = regexp.MustCompile(`<a href="([^"]+)">([^<]*)</a>`)

// LinkGroup fetches job listings from LINK group's "currently open
// positions" page, which lists openings for all of its countries (Serbia,
// Bosnia and Herzegovina, Romania, Ukraine, Moldova) on one hand-maintained
// HTML page.
type LinkGroup struct {
	Client *http.Client
}

func (l *LinkGroup) Name() string { return "linkgroup" }

func (l *LinkGroup) Fetch(ctx context.Context) ([]*models.Job, error) {
	client := l.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, linkGroupCareersURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkgroup: unexpected status %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(raw)

	start := strings.Index(body, linkGroupContentStart)
	if start < 0 {
		return nil, fmt.Errorf("linkgroup: content block not found")
	}
	end := strings.Index(body[start:], linkGroupContentEnd)
	if end < 0 {
		end = len(body)
	} else {
		end += start
	}
	content := body[start:end]

	headings := linkGroupSectionRe.FindAllStringSubmatchIndex(content, -1)
	jobs := make([]*models.Job, 0, len(headings))
	for i, h := range headings {
		country := html.UnescapeString(strings.TrimSpace(content[h[2]:h[3]]))
		sectionStart := h[1]
		sectionEnd := len(content)
		if i+1 < len(headings) {
			sectionEnd = headings[i+1][0]
		}
		section := content[sectionStart:sectionEnd]

		for _, m := range linkGroupLinkRe.FindAllStringSubmatch(section, -1) {
			jobURL := html.UnescapeString(m[1])
			title := html.UnescapeString(m[2])
			title = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(title), ">>"))
			if title == "" {
				continue
			}

			jobs = append(jobs, &models.Job{
				ExternalID: jobURL,
				Title:      title,
				Location:   country,
				URL:        jobURL,
				Source:     l.Name(),
			})
		}
	}

	return jobs, nil
}
