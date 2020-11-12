package m3uparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	country_mapper "github.com/pirsquare/country-mapper"
)

type channel map[string]interface{}

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
	p.StreamsInfoBackup = p.StreamsInfo
}

func (p *M3uParser) parseLines() {
	initClient()
	for lineNumber := range p.lines {
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
				"name":     title,
				"logo":     logo,
				"url":      streamLink,
				"category": category,
				"language": language,
				"country":  map[string]string{"code": countryCode, "name": countryName},
				"tvg":      map[string]string{"id": tvgID, "name": tvgName, "url": tvgURL},
				"status":   statusString,
			})
		}
	}
}

// FilterBy :
func (p *M3uParser) FilterBy(key string, filters []string, retrieve bool, nestedKey bool) {
	var key0, key1 string
	var filteredStreams []channel
	if nestedKey {
		splittedKey := strings.Split(key, "-")
		key0, key1 = splittedKey[0], splittedKey[1]
	}
	if len(filters) == 0 {
		fmt.Println("Filter word/s missing!!!")
		return
	}

	switch nestedKey {
	case false:
		for _, stream := range p.StreamsInfo {
			if val, ok := stream[key]; ok {
				for _, filter := range filters {
					if retrieve {
						if strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), strings.ToLower(filter)) {
							filteredStreams = append(filteredStreams, stream)
						}
					} else {
						if !strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), strings.ToLower(filter)) {
							filteredStreams = append(filteredStreams, stream)
						}
					}
				}
			}
		}
	case true:
		for _, stream := range p.StreamsInfo {
			if val, ok := stream[key0]; ok {
				switch v := val.(type) {
				case map[string]string:
					if val, ok := v[key1]; ok {
						for _, filter := range filters {
							if retrieve {
								if strings.Contains(strings.ToLower(val), strings.ToLower(filter)) {
									filteredStreams = append(filteredStreams, stream)
								}
							} else {
								if !strings.Contains(strings.ToLower(val), strings.ToLower(filter)) {
									filteredStreams = append(filteredStreams, stream)
								}
							}
						}
					}
				}
			}
		}
	}

	p.StreamsInfo = filteredStreams
}

func (p *M3uParser) ResetOperations() {
	p.StreamsInfo = p.StreamsInfoBackup
}

func (p *M3uParser) RemoveByExtension(extension []string) {
	p.FilterBy("url", extension, false, false)
}

func (p *M3uParser) RetrieveByExtension(extension []string) {
	p.FilterBy("url", extension, true, false)
}

func (p *M3uParser) RemoveByCategory(extension []string) {
	p.FilterBy("category", extension, false, false)
}

func (p *M3uParser) RetrieveByCategory(extension []string) {
	p.FilterBy("category", extension, true, false)
}

func (p *M3uParser) SortBy(key string, asc bool, nestedKey bool) {
	var key0, key1 string
	if nestedKey {
		splittedKey := strings.Split(key, "-")
		key0, key1 = splittedKey[0], splittedKey[1]
	}
	switch nestedKey {
	case true:
		value, ok := p.StreamsInfo[0][key0].(map[string]string)
		if ok {
			if _, ok := value[key1]; ok {
				sort.Slice(p.StreamsInfo, func(i, j int) bool {
					val1, _ := p.StreamsInfo[i][key0].(map[string]string)
					val2, _ := p.StreamsInfo[j][key0].(map[string]string)
					if asc {
						return val1[key1] < val2[key1]
					} else {
						return val1[key1] > val2[key1]
					}
				})
			}

		}
	case false:
		if _, ok := p.StreamsInfo[0][key]; ok {
			sort.Slice(p.StreamsInfo, func(i, j int) bool {
				val1, _ := p.StreamsInfo[i][key].(string)
				val2, _ := p.StreamsInfo[j][key].(string)
				if asc {
					return val1 < val2
				} else {
					return val1 > val2
				}
			})
		}
	}
}

func (p *M3uParser) GetStreamsSlice() []channel {
	return p.StreamsInfo
}

func (p *M3uParser) GetStreamsJson() string {
	fmt.Println(p.StreamsInfo)
	jsonByte, err := json.Marshal(p.StreamsInfo)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return string(jsonByte)
}

func (p *M3uParser) GetRandomStream(shuffle bool) channel {
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Shuffle(len(p.StreamsInfo), func(i, j int) {
		p.StreamsInfo[i], p.StreamsInfo[j] = p.StreamsInfo[j], p.StreamsInfo[i]
	})
	return p.StreamsInfo[rand.Intn(len(p.StreamsInfo))]
}

func (p *M3uParser) SaveJsonToFile(filename string) {
	json, err := json.MarshalIndent(p.StreamsInfo, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filename, json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
