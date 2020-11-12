package m3uparser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	country_mapper "github.com/pirsquare/country-mapper"
)

type tvg struct {
	id   string
	name string
	url  string
}

type country struct {
	code string
	name string
}

type channel struct {
	name     string
	logo     string
	url      string
	category string
	language string
	country  country
	tvg      tvg
	status   string
}

// M3uParser - A parser for m3u files.
type M3uParser struct {
	StreamsInfo       []channel
	StreamsInfoBackup []channel
	lines             []string
	timeout           int
	userAgent         string
	checkLive         bool
	content           string
}

var countryClient *country_mapper.CountryInfoClient
var wg sync.WaitGroup

func initClient() {
	client, err := country_mapper.Load()
	if err != nil {
		panic(err)
	}

	countryClient = client
}

func errorLogger(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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

func (p *M3uParser) isLive(url string, status *bool) {
	_, err := http.Get(url)
	if err != nil {
		*status = false
	} else {
		*status = true
	}
	fmt.Println(*status, url)
	if p.checkLive {
		wg.Done()
	}
}

// ParseM3u -
func (p *M3uParser) ParseM3u(path string, checkLive bool) {
	p.checkLive = checkLive
	if isValidURL(path) {
		resp, err := http.Get(path)
		errorLogger(err)
		body, err := ioutil.ReadAll(resp.Body)
		p.content = string(body)
	} else {
		body, err := ioutil.ReadFile(path)
		errorLogger(err)
		p.content = string(body)
	}
	if p.content != "" {
		for _, line := range strings.Split(p.content, "\n") {
			if strings.TrimSpace(line) != "" {
				p.lines = append(p.lines, strings.TrimSpace(line))
			}
		}
	}
	if len(p.lines) > 0 {
		p.parseLines()
	} else {
		fmt.Println("No content to parse!!!")
	}
	if p.checkLive {
		wg.Wait()
	}
	fmt.Println(p.StreamsInfo)
}

func (p *M3uParser) parseLines() {
	initClient()
	for lineNumber := range p.lines {
		fmt.Println(lineNumber)
		if regexp.MustCompile("#EXTINF").Match([]byte(p.lines[lineNumber])) {
			go p.parseLine(lineNumber)
		}
	}

}

func (p *M3uParser) parseLine(lineNumber int) {
	var status bool
	lineInfo := p.lines[lineNumber]
	streamLink := ""
	streamsLink := []string{}

	for i := range [2]int{1, 2} {
		_, err := url.ParseRequestURI(p.lines[lineNumber+i])
		if err == nil {
			streamsLink = append(streamsLink, p.lines[lineNumber+i])
			break
		}
	}
	if len(streamsLink) != 0 {
		streamLink = streamsLink[0]
		if lineInfo != "" && streamLink != "" {
			tvgName := isPresent("tvg-name=\"(.*?)\"", lineInfo)
			tvgID := isPresent("tvg-id=\"(.*?)\"", lineInfo)
			logo := isPresent("tvg-logo=\"(.*?)\"", lineInfo)
			category := isPresent("group-title=\"(.*?)\"", lineInfo)
			title := isPresent("[,](.*?)$", lineInfo)
			countryCode := isPresent("tvg-country=\"(.*?)\"", lineInfo)
			language := isPresent("tvg-language=\"(.*?)\"", lineInfo)
			tvgURL := isPresent("tvg-url=\"(.*?)\"", lineInfo)
			countryName := countryClient.MapByAlpha2(strings.ToUpper(countryCode)).Name
			statusString := "NOT CHECKED"
			if p.checkLive {
				wg.Add(1)
				go p.isLive(streamLink, &status)
				statusString = "GOOD"
				if !status {
					statusString = "BAD"
				}
			} else {
				statusString = "NOT CHECKED"
			}

			p.StreamsInfo = append(p.StreamsInfo, channel{
				name:     title,
				logo:     logo,
				url:      streamLink,
				category: category,
				language: language,
				country:  country{code: countryCode, name: countryName},
				tvg:      tvg{id: tvgID, name: tvgName, url: tvgURL},
				status:   statusString,
			})
		}
	}
}
