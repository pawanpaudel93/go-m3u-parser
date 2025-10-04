package m3uparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
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
	enforceSchema     bool
	lines             []string
	Timeout           int
	UserAgent         string
	CheckLive         bool
	content           string
	regexes           map[string]*regexp.Regexp
	mutex             sync.Mutex
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
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
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
	defer wg.Done()
	_, err := Get(url, p.UserAgent, time.Duration(p.Timeout)*time.Second)
	if err != nil {
		channel["status"] = "BAD"
	} else {
		channel["status"] = "GOOD"
	}
	bar.Increment()
}

// ParseM3u parses the content of local file/URL or raw M3U content.
// It downloads the file from the given URL, reads from a local file path, or parses raw M3U content directly.
// The function parses line by line to extract stream information into a structured format.
//
// Parameters:
//   - source: Can be one of the following:
//   - URL: A valid HTTP/HTTPS URL pointing to an M3U file
//   - File path: Local file path to an M3U file (e.g., "/path/to/file.m3u")
//   - Raw content: M3U content string
//   - checkLive: Boolean flag to check if stream URLs are accessible and working
//   - enforceSchema: If true, keeps all fields even with empty values; if false, removes keys with empty string values
func (p *M3uParser) ParseM3u(source string, checkLive bool, enforceSchema bool) {
	p.enforceSchema = enforceSchema
	p.regexes = make(map[string]*regexp.Regexp)
	p.regexes["file"] = compileRegex(`(?m)^[a-zA-Z]:\\((?:.*?\\)*).*.[\d\w]{3,5}$|^(/[^/]*)+/?.[\d\w]{3,5}$`)
	p.regexes["tvgName"] = compileRegex("tvg-name=\"(.*?)\"")
	p.regexes["tvgID"] = compileRegex("tvg-id=\"(.*?)\"")
	p.regexes["logo"] = compileRegex("tvg-logo=\"(.*?)\"")
	p.regexes["category"] = compileRegex("group-title=\"(.*?)\"")
	p.regexes["title"] = compileRegex(`[,](.*?)$`)
	p.regexes["countryCode"] = compileRegex("tvg-country=\"(.*?)\"")
	p.regexes["language"] = compileRegex("tvg-language=\"(.*?)\"")
	p.regexes["tvgURL"] = compileRegex("tvg-url=\"(.*?)\"")

	if p.Timeout == 0 {
		p.Timeout = 5
	}
	if p.UserAgent == "" {
		p.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
	}
	p.CheckLive = checkLive

	// Check if source starts with "#EXTM3U", is empty, or contains newlines - if so, treat as raw content
	trimmedSource := strings.TrimSpace(source)
	if strings.HasPrefix(trimmedSource, "#EXTM3U") || trimmedSource == "" || strings.Contains(source, "\n") {
		log.Infoln("Started parsing m3u from raw content...")
		p.content = source
	} else if isValidURL(source) {
		log.Infoln("Started parsing m3u URL...")
		resp, err := http.Get(source)
		errorLogger(err)
		body, err := ioutil.ReadAll(resp.Body)
		errorLogger(err)
		p.content = string(body)
	} else {
		log.Infoln("Started parsing m3u file...")
		body, err := ioutil.ReadFile(source)
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
	p.streamsInfoBackup = p.streamsInfo
}

func (p *M3uParser) parseLines() {
	re := compileRegex("#EXTINF")
	var count int
	for lineNumber := range p.lines {
		if re.Match([]byte(p.lines[lineNumber])) {
			count++
		}
	}
	if p.CheckLive {
		bar = pb.StartNew(count)
	}
	wg.Add(count)
	for lineNumber := range p.lines {
		if re.Match([]byte(p.lines[lineNumber])) {
			go p.parseLine(lineNumber)
		}
	}
	wg.Wait()
	if p.CheckLive {
		bar.Finish()
	}
}

func (p *M3uParser) parseLine(lineNumber int) {
	defer wg.Done()
	var isFile bool
	var streamLink string
	channel := make(Channel)
	lineInfo := p.lines[lineNumber]

	for i := range [2]int{1, 2} {
		isUrl := isValidURL(p.lines[lineNumber+i])
		if isUrl {
			streamLink = p.lines[lineNumber+i]
			break
		} else if p.regexes["file"].Match([]byte(p.lines[lineNumber+i])) {
			streamLink = p.lines[lineNumber+i]
			isFile = true
			break
		}
	}

	if lineInfo != "" && streamLink != "" {
		var countryName string

		tvg := make(map[string]string)
		tvg["name"] = getByRegex(p.regexes["tvgName"], lineInfo)
		tvg["id"] = getByRegex(p.regexes["tvgID"], lineInfo)
		tvg["url"] = getByRegex(p.regexes["tvgURL"], lineInfo)
		logo := getByRegex(p.regexes["logo"], lineInfo)
		category := getByRegex(p.regexes["category"], lineInfo)
		title := getByRegex(p.regexes["title"], lineInfo)
		countryCode := getByRegex(p.regexes["countryCode"], lineInfo)
		language := getByRegex(p.regexes["language"], lineInfo)
		country := countryClient.MapByAlpha2(strings.ToUpper(countryCode))
		if country == nil {
			countryName = ""
		} else {
			countryName = country.Name
		}
		if p.CheckLive {
			if isFile {
				bar.Increment()
				channel["status"] = "GOOD"
			} else {
				wg.Add(1)
				go p.isLive(streamLink, channel)
			}
		}
		if title != "" || p.enforceSchema {
			channel["title"] = title
		}
		if logo != "" || p.enforceSchema {
			channel["logo"] = logo
		}
		if category != "" || p.enforceSchema {
			channel["category"] = category
		}
		if language != "" || p.enforceSchema {
			channel["language"] = language
		}
		if tvg["id"] != "" || tvg["name"] != "" || tvg["url"] != "" || p.enforceSchema {
			temp_tvg := make(map[string]string)
			for key, value := range tvg {
				if value != "" || p.enforceSchema {
					temp_tvg[key] = value
				}
			}
			channel["tvg"] = temp_tvg
		}
		if countryCode != "" || p.enforceSchema {
			channel["country"] = map[string]string{"code": countryCode, "name": countryName}
		}
		channel["url"] = streamLink
		p.mutex.Lock()
		p.streamsInfo = append(p.streamsInfo, channel)
		p.mutex.Unlock()
	} else {
		if p.CheckLive {
			bar.Increment()
		}
	}
}

// FilterBy filters stream information.
// It retrieves/removes stream information from streams information slice using filter/s on key.
//
// Parameters:
//   - key: Key can be single or nested. eg. key='name', key='language-name'
//   - filters: Slice of filter/s to perform the retrieve or remove operation.
//   - retrieve: True to retrieve and False for removing based on key.
func (p *M3uParser) FilterBy(key string, filters []string, retrieve bool) {
	if p.isEmpty() {
		log.Infof("No streams info to filter.")
		return
	}

	if len(filters) == 0 {
		log.Warnln("Filter word/s missing!!!")
		return
	}

	var key0, key1 string
	var filteredStreams []Channel
	var nestedKey bool

	splittedKey := strings.Split(key, "-")
	if len(splittedKey) == 2 {
		key0, key1 = splittedKey[0], splittedKey[1]
		nestedKey = true
	} else if len(splittedKey) > 2 {
		log.Warnf("Nested key is seperated by multiple key seperator -")
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

// ResetOperations resets the stream information slice to initial state before various operations.
func (p *M3uParser) ResetOperations() {
	p.streamsInfo = p.streamsInfoBackup
}

// RemoveByExtension removes stream information with certain extension/s.
// It removes stream information from streams information slice based on extension/s provided.
//
// Parameters:
//   - extension: Name of the extension like mp4, m3u8 etc. It is slice of extension/s.
func (p *M3uParser) RemoveByExtension(extension []string) {
	p.FilterBy("url", extension, false)
}

// RetrieveByExtension selects only streams information with a certain extension/s.
// It retrieves the stream information based on extension/s provided.
//
// Parameters:
//   - extension: Name of the extension like mp4, m3u8 etc. It is slice of extension/s.
func (p *M3uParser) RetrieveByExtension(extension []string) {
	p.FilterBy("url", extension, true)
}

// RemoveByCategory removes streams information with category containing a certain filter word/s.
// It removes stream information based on category using filter word/s.
//
// Parameters:
//   - category: It is slice of category/categories.
func (p *M3uParser) RemoveByCategory(category []string) {
	p.FilterBy("category", category, false)
}

// RetrieveByCategory retrieves only streams information that contains a certain filter word/s.
// It retrieves stream information based on category/categories.
//
// Parameters:
//   - category: It is slice of category/categories.
func (p *M3uParser) RetrieveByCategory(category []string) {
	p.FilterBy("category", category, true)
}

// SortBy sorts streams information.
// It sorts streams information slice sorting by key in asc/desc order.
//
// Parameters:
//   - key: It can be single or nested key.
//   - asc: Sort by asc or desc order.
func (p *M3uParser) SortBy(key string, asc bool) {
	if p.isEmpty() {
		log.Infof("No streams info to sort.")
		return
	}

	var key0, key1 string
	var nestedKey bool

	splittedKey := strings.Split(key, "-")
	if len(splittedKey) == 2 {
		key0, key1 = splittedKey[0], splittedKey[1]
		nestedKey = true
	} else if len(splittedKey) > 2 {
		log.Warnf("Nested key is seperated by multiple key seperator -")
		return
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

// GetStreamsSlice gets the parsed streams information slice.
// It returns the streams information slice.
func (p *M3uParser) GetStreamsSlice() []Channel {
	return p.streamsInfo
}

// GetStreamsJSON gets the streams information as json.
func (p *M3uParser) GetStreamsJSON() string {
	jsonByte, err := json.Marshal(p.streamsInfo)
	if err != nil {
		log.Warnln(err)
		return ""
	}
	return string(jsonByte)
}

// GetRandomStream returns a random stream information.
// It returns a random stream information with shuffle if required.
//
// Parameters:
//   - shuffle: To shuffle the streams information slice before returning the random stream information.
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

// ToFile saves streams information to a json/m3u file.
// It saves streams information as a JSON/M3U file with a given filename.
//
// Parameters:
//   - filename: Name of the file to save streams information.
func (p *M3uParser) ToFile(fileName string) {
	var format string
	if p.isEmpty() {
		log.Infoln("No streams info to save.")
		return
	}
	if len(strings.Split(fileName, ".")) > 1 {
		format = strings.ToLower(strings.Split(fileName, ".")[1])
	}
	log.Infof("Saving to file: %s", fileName)
	switch format {
	case "json":
		json, err := json.MarshalIndent(p.streamsInfo, "", "    ")
		json = []byte(strings.ReplaceAll(string(json), `: ""`, ": null"))
		errorLogger(err)
		if !strings.Contains(fileName, "json") {
			fileName = fileName + ".json"
		}
		err = ioutil.WriteFile(fileName, json, 0644)
		errorLogger(err)
	case "m3u":
		content := []string{"#EXTM3U"}
		for _, stream := range p.streamsInfo {
			line := "#EXTINF:-1"
			if tvg, ok := stream["tvg"]; ok {
				tvg := tvg.(map[string]string)
				for key, val := range tvg {
					if val != "" {
						line += fmt.Sprintf(` tvg-%s="%s"`, key, val)
					}
				}
			}
			if logo, ok := stream["logo"]; ok && logo != "" {
				line += fmt.Sprintf(` tvg-logo="%s"`, logo)
			}
			if country, ok := stream["country"]; ok {
				country := country.(map[string]string)
				if code, ok := country["code"]; ok && code != "" {
					line += fmt.Sprintf(` tvg-country="%s"`, code)
				}
			}
			if language, ok := stream["language"]; ok && language != "" {
				line += fmt.Sprintf(` tvg-language="%s"`, language)
			}
			if category, ok := stream["category"]; ok && category != "" {
				line += fmt.Sprintf(` group-title="%s"`, category)
			}
			if title, ok := stream["title"]; ok && title != "" {
				line += fmt.Sprintf(`,%s`, title)
			}
			content = append(content, line)
			content = append(content, stream["url"].(string))
		}
		err := ioutil.WriteFile(fileName, []byte(strings.Join(content, "\n")), 0666)
		errorLogger(err)
	default:
		log.Infoln("File extension not present/supported !!!")
	}
}
