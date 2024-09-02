package pattern

import (
	"errors"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type MultiplePatternsError struct {
	URLs []string
}

func (m MultiplePatternsError) Error() string {
	return "multiple patterns found: " + strings.Join(m.URLs, ", ")
}

var ErrNoPatternLinks = errors.New("no pattern links found")

func FindHrefPatterns(resp *http.Response) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var matches []string
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if href, found := s.Attr("href"); found {
			if slices.Contains(Extensions(), path.Ext(href)) {
				matches = append(matches, href)
			}
		}
	})

	slices.Sort(matches)
	slices.Reverse(matches) // Reverse so that .rle comes before .cells
	matches = slices.CompactFunc(matches, func(s1 string, s2 string) bool {
		// Remove instances of the same path with different extensions
		return strings.TrimSuffix(s1, path.Ext(s1)) == strings.TrimSuffix(s2, path.Ext(s2))
	})
	slices.Reverse(matches)

	if len(matches) == 0 {
		return nil, ErrNoPatternLinks
	}
	return matches, nil
}
