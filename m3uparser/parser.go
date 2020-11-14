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

	pb "github.com/cheggaaa/pb/v3"
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
	Timeout           int
	UserAgent         string
	CheckLive         bool
	content           string
}

var countryClient *country_mapper.CountryInfoClient
var wg sync.WaitGroup
var bar *pb.ProgressBar

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

func (p *M3uParser) isEmpty() bool {
	return len(p.streamsInfo) == 0
}

func (p *M3uParser) isLive(url string, channel Channel) {
	_, err := Get(url, p.UserAgent, time.Duration(p.Timeout)*time.Second)
	if err != nil {
		channel["status"] = "BAD"
	} else {
		channel["status"] = "GOOD"
	}
	if p.CheckLive {
		bar.Increment()
		wg.Done()
	}
}

// ParseM3u - Parses the content of local file/URL.
func (p *M3uParser) ParseM3u(path string, checkLive bool) {
	if p.Timeout == 0 {
		p.Timeout = 5
	}
	if p.UserAgent == "" {
		p.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
	}
	p.CheckLive = checkLive
	if isValidURL(path) {
		log.Infoln("Started parsing m3u URL...")
		resp, err := http.Get(path)
		errorLogger(err)
		body, err := ioutil.ReadAll(resp.Body)
		p.content = string(body)
	} else {
		log.Infoln("Started parsing m3u file...")
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
	if p.CheckLive {
		wg.Wait()
		bar.Finish()
	}
	p.streamsInfoBackup = p.streamsInfo
}

func (p *M3uParser) parseLines() {
	re := regexp.MustCompile("#EXTINF")
	if p.CheckLive {
		var count int
		for lineNumber := range p.lines {
			if re.Match([]byte(p.lines[lineNumber])) {
				count++
			}
		}
		bar = pb.StartNew(count - 1)
	}
	for lineNumber := range p.lines {
		if re.Match([]byte(p.lines[lineNumber])) {
			go p.parseLine(lineNumber)
		}
	}
}

func (p *M3uParser) parseLine(lineNumber int) {
	channel := make(Channel)
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
			if p.CheckLive {
				wg.Add(1)
				go p.isLive(streamLink, channel)
			} else {
				channel["status"] = "NOT CHECKED"
			}
			channel["name"] = title
			channel["logo"] = logo
			channel["url"] = streamLink
			channel["category"] = category
			channel["language"] = language
			channel["country"] = map[string]string{"code": countryCode, "name": countryName}
			channel["tvg"] = map[string]string{"id": tvgID, "name": tvgName, "url": tvgURL}
			p.streamsInfo = append(p.streamsInfo, channel)
		}
	}
}

// FilterBy - Filter streams infomation.
func (p *M3uParser) FilterBy(key string, filters []string, retrieve bool, nestedKey bool) {
	if p.isEmpty() {
		log.Infof("No streams info to filter.")
		return
	}
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
	if p.isEmpty() {
		log.Infof("No streams info to sort.")
		return
	}
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
	jsonByte, err := json.Marshal(p.streamsInfo)
	if err != nil {
		log.Warnln(err)
		return ""
	}
	return string(jsonByte)
}

// GetRandomStream - Return a random stream information.
func (p *M3uParser) GetRandomStream(shuffle bool) Channel {
	if p.isEmpty() {
		log.Infoln("No streams info for random selection.")
		return Channel{}
	}
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Shuffle(len(p.streamsInfo), func(i, j int) {
		p.streamsInfo[i], p.streamsInfo[j] = p.streamsInfo[j], p.streamsInfo[i]
	})
	return p.streamsInfo[rand.Intn(len(p.streamsInfo))]
}

// SaveJSONToFile -  Save to JSON file.
func (p *M3uParser) SaveJSONToFile(filename string) {
	log.Infof("Saving to file: %s", filename)
	json, err := json.MarshalIndent(p.streamsInfo, "", "    ")
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
