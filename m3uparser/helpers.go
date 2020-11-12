package m3uparser

import (
	"net/url"
	"regexp"
)

func isPresent(regex string, content string) string {
	re := regexp.MustCompile(regex)
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
