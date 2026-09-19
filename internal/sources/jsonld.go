package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
)

var jsonLDRe = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)

// jobPostingDescription returns the description from the schema.org
// JobPosting JSON-LD block on a job's public page (TalentLyft pages have one).
func jobPostingDescription(ctx context.Context, client *http.Client, pageURL string) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: unexpected status %s", pageURL, resp.Status)
	}

	page, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	for _, m := range jsonLDRe.FindAllSubmatch(page, -1) {
		var p struct {
			Type        string `json:"@type"`
			Description string `json:"description"`
		}
		if json.Unmarshal(m[1], &p) == nil && p.Type == "JobPosting" {
			return html.UnescapeString(p.Description), nil // the JSON-LD value is entity-escaped HTML
		}
	}
	return "", fmt.Errorf("%s: no JobPosting JSON-LD", pageURL)
}
