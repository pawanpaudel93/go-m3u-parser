package m3uparser

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

func compileRegex(regex string) *regexp.Regexp {
	return regexp.MustCompile(regex)
}

func getByRegex(re *regexp.Regexp, content string) string {
	matches := re.FindStringSubmatch(content)
	if len(matches) > 0 {
		return matches[1]
	}
	return ""
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// Get - requests
func Get(URL string, userAgent string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Warnln(err)
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	return resp, err
}
