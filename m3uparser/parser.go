package m3uparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	country_mapper "github.com/pirsquare/country-mapper"
)

// Channel - Map containing streams information.
type Channel map[string]interface{}

// M3uParser - A parser for m3u files.
type M3uParser struct {
	streamsInfo       []Channel
	streamsInfoBackup []Channel
	lines             []string
	timeout           int
	userAgent         string
	checkLive         bool
	content           string
}

var countryClient *country_mapper.CountryInfoClient
var wg sync.WaitGroup

func init() {
	client, err := country_mapper.Load()
	if err != nil {
		panic(err)
	}
	countryClient = client

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
	log.Infoln("Parser started")
}

func errorLogger(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
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

// ParseM3u - Parses the content of local file/URL.
func (p *M3uParser) ParseM3u(path string, checkLive bool) {
	p.checkLive = checkLive
	if isValidURL(path) {
		log.Infoln("Started parsing m3u file...")
		resp, err := http.Get(path)
		errorLogger(err)
		body, err := ioutil.ReadAll(resp.Body)
		p.content = string(body)
	} else {
		log.Infoln("Started parsing m3u URL...")
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
		log.Infoln("No content to parse!!!")
	}
	if p.checkLive {
		wg.Wait()
	}
	p.streamsInfoBackup = p.streamsInfo
}

func (p *M3uParser) parseLines() {
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

			p.streamsInfo = append(p.streamsInfo, Channel{
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

// FilterBy - Filter streams infomation.
func (p *M3uParser) FilterBy(key string, filters []string, retrieve bool, nestedKey bool) {
	var key0, key1 string
	var filteredStreams []Channel
	if nestedKey {
		splittedKey := strings.Split(key, "-")
		key0, key1 = splittedKey[0], splittedKey[1]
	}
	if len(filters) == 0 {
		log.Warnln("Filter word/s missing!!!")
		return
	}

	switch nestedKey {
	case false:
		for _, stream := range p.streamsInfo {
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
		for _, stream := range p.streamsInfo {
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
	p.streamsInfo = filteredStreams
}

//ResetOperations - Reset the stream information list to initial state before various operations.
func (p *M3uParser) ResetOperations() {
	p.streamsInfo = p.streamsInfoBackup
}

// RemoveByExtension - Remove stream information with certain extension/s.
func (p *M3uParser) RemoveByExtension(extension []string) {
	p.FilterBy("url", extension, false, false)
}

// RetrieveByExtension - Select only streams information with a certain extension/s.
func (p *M3uParser) RetrieveByExtension(extension []string) {
	p.FilterBy("url", extension, true, false)
}

// RemoveByCategory - Removes streams information with category containing a certain filter word/s.
func (p *M3uParser) RemoveByCategory(extension []string) {
	p.FilterBy("category", extension, false, false)
}

// RetrieveByCategory - Retrieve only streams information that contains a certain filter word/s.
func (p *M3uParser) RetrieveByCategory(extension []string) {
	p.FilterBy("category", extension, true, false)
}

// SortBy - Sort streams information.
func (p *M3uParser) SortBy(key string, asc bool, nestedKey bool) {
	var key0, key1 string
	if nestedKey {
		splittedKey := strings.Split(key, "-")
		key0, key1 = splittedKey[0], splittedKey[1]
	}
	switch nestedKey {
	case true:
		value, ok := p.streamsInfo[0][key0].(map[string]string)
		if ok {
			if _, ok := value[key1]; ok {
				sort.Slice(p.streamsInfo, func(i, j int) bool {
					val1, _ := p.streamsInfo[i][key0].(map[string]string)
					val2, _ := p.streamsInfo[j][key0].(map[string]string)
					if asc {
						return val1[key1] < val2[key1]
					} else {
						return val1[key1] > val2[key1]
					}
				})
			}

		}
	case false:
		if _, ok := p.streamsInfo[0][key]; ok {
			sort.Slice(p.streamsInfo, func(i, j int) bool {
				val1, _ := p.streamsInfo[i][key].(string)
				val2, _ := p.streamsInfo[j][key].(string)
				if asc {
					return val1 < val2
				} else {
					return val1 > val2
				}
			})
		}
	}
}

// GetStreamsSlice - Get the parsed streams information slice.
func (p *M3uParser) GetStreamsSlice() []Channel {
	return p.streamsInfo
}

// GetStreamsJSON - Get the streams information as json.
func (p *M3uParser) GetStreamsJSON() string {
	fmt.Println(p.streamsInfo)
	jsonByte, err := json.Marshal(p.streamsInfo)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return string(jsonByte)
}

// GetRandomStream - Return a random stream information.
func (p *M3uParser) GetRandomStream(shuffle bool) Channel {
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Shuffle(len(p.streamsInfo), func(i, j int) {
		p.streamsInfo[i], p.streamsInfo[j] = p.streamsInfo[j], p.streamsInfo[i]
	})
	return p.streamsInfo[rand.Intn(len(p.streamsInfo))]
}

// SaveJSONToFile -  Save to JSON file.
func (p *M3uParser) SaveJSONToFile(filename string) {
	log.Infof("Saving to file: %s", filename)
	json, err := json.MarshalIndent(p.streamsInfo, "", "  ")
	if err != nil {
		log.Warnln(err)
	}
	if !strings.Contains(filename, "json") {
		filename = filename + ".json"
	}
	err = ioutil.WriteFile(filename, json, 0644)
	if err != nil {
		log.Warnln(err)
	}
}
