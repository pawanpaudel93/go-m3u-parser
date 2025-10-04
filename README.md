# go-m3u-parser

![version](https://img.shields.io/badge/version-0.0.1-blue.svg?cacheSeconds=2592000)

A parser for m3u files. It parses the contents of m3u file to a slice of streams information which can be saved as a JSON file.

## Install

```sh
go get github.com/pawanpaudel93/go-m3u-parser
```

## Example

```go
package main

import (
 "fmt"

 m3uparser "github.com/pawanpaudel93/go-m3u-parser/m3uparser"
)

func main() {
 // userAgent and timeout is optional. default timeout is 5 seconds and userAgent is latest chrome version 86.
 userAgent := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
 timeout := 5 // in seconds
 parser := m3uparser.M3uParser{UserAgent: userAgent, Timeout: timeout}
 // file path can also be used /home/pawan/Downloads/ru.m3u
 parser.ParseM3u("https://drive.google.com/uc?id=1VGv8ZYQrrSYPVQ7GCWLgjMl6w9Ccrs4v&export=download", true, true)
 parser.FilterBy("status", []string{"GOOD"}, true)
 parser.SortBy("category", true)
 fmt.Println("Saved stream information: ", len(parser.GetStreamsSlice()))
 parser.ToFile("rowdy.json")
}

```

## Usage

### Basic Usage

```go
 userAgent := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
 timeout := 5 // in seconds
 parser := m3uparser.M3uParser{UserAgent: userAgent, Timeout: timeout}
 // file path can also be used /home/pawan/Downloads/ru.m3u
 parser.ParseM3u("https://drive.google.com/uc?id=1VGv8ZYQrrSYPVQ7GCWLgjMl6w9Ccrs4v&export=download", true, true)
```

### Advanced Usage Examples

```go
// Parse from different sources
parser := m3uparser.M3uParser{}

// 1. Parse from URL
parser.ParseM3u("https://example.com/playlist.m3u", false, false)

// 2. Parse from local file
parser.ParseM3u("/path/to/local/playlist.m3u", false, false)

// 3. Parse raw M3U content directly
rawM3U := `#EXTM3U
#EXTINF:-1 tvg-id="channel1.np",Channel 1 (1080p)
https://example.com/stream1.m3u8
#EXTINF:-1 tvg-id="channel2.np",Channel 2 (720p)
https://example.com/stream2.m3u8`
parser.ParseM3u(rawM3U, false, false)

// 4. Parse with live stream validation
parser.ParseM3u("https://example.com/playlist.m3u", true, false)

// 5. Parse with enforced schema (keeps empty fields)
parser.ParseM3u("https://example.com/playlist.m3u", false, true)
```

### Configuration Options

```go
parser := m3uparser.M3uParser{
    UserAgent: "Custom User Agent",  // Optional: Default is Chrome 86
    Timeout:   10,                   // Optional: Default is 5 seconds
}
```

>Functions

```go
func (p *M3uParser) ParseM3u(source string, checkLive bool, enforceSchema bool) {

        """Parses the content of local file/URL or raw M3U content.
        It downloads the file from the given URL, reads from a local file path, or parses raw M3U content directly.
        The function parses line by line to extract stream information into a structured format.
  
        Parameters:
        - source: Can be one of the following:
          - URL: A valid HTTP/HTTPS URL pointing to an M3U file
          - File path: Local file path to an M3U file (e.g., "/path/to/file.m3u")
          - Raw content: M3U content string
        - checkLive: Boolean flag to check if stream URLs are accessible and working
        - enforceSchema: If true, keeps all fields even with empty values; if false, removes keys with empty string values
        """
}
 
func (p *M3uParser) FilterBy(key string, filters []string, retrieve bool) {

        """Filter streams information.
        It retrieves/removes stream information from streams information slice using filter/s on key.

        Parameters:
        - key: Key can be single or nested. eg. key='name', key='language-name'
        - filters: Slice of filter/s to perform the retrieve or remove operation.
        - retrieve: True to retrieve and False for removing based on key.
        """
  
}
  
func (p *M3uParser) ResetOperations() {

        """Reset the stream information slice to initial state before various operations."""
  
}
  
func (p *M3uParser) RemoveByExtension(extension []string) {

        """Remove stream information with certain extension/s.
        It removes stream information from streams information slice based on extension/s provided.
  
        Parameters:
        - extension: Name of the extension like mp4, m3u8 etc. It is slice of extension/s.
        """
  
}
  
func (p *M3uParser) RetrieveByExtension(extension []string) {

        """Select only streams information with a certain extension/s.
        It retrieves the stream information based on extension/s provided.
  
        Parameters:
        - extension: Name of the extension like mp4, m3u8 etc. It is slice of extension/s.
        """
}
  
func (p *M3uParser) RemoveByCategory(category []string) {

        """Removes streams information with category containing a certain filter word/s.
        It removes stream information based on category using filter word/s.
  
        Parameters:
        - category: It is slice of category/categories.
        """
}
  
func (p *M3uParser) RetrieveByCategory(category []string) {

        """Retrieve only streams information that contains a certain filter word/s.
        It retrieves stream information based on category/categories.

        Parameters:
        - category: It is slice of category/categories.
        """
}
  
func (p *M3uParser) SortBy(key string, asc bool) {

        """Sort streams information.
        It sorts streams information slice sorting by key in asc/desc order.

        Parameters:
        - key: It can be single or nested key.
        - asc: Sort by asc or desc order.
        """
}

func (p *M3uParser) GetStreamsJSON() string {

        """Get the streams information as json."""
}
  
func (p *M3uParser) GetStreamsSlice() []Channel {

        """Get the parsed streams information slice.
        It returns the streams information slice.
        """
  
func (p *M3uParser) GetRandomStream(shuffle bool) Channel {

        """Return a random stream information
        It returns a random stream information with shuffle if required.

        Parameters:
        - shuffle: To shuffle the streams information slice before returning the random stream information.
        """
}
  
func (p *M3uParser) ToFile(filename string) {

        """Save to json/m3u file.
        It saves streams information as a JSON/M3U file with a given filename.

        Parameters:
        - filename: Name of the file to save streams information.
        """
}

```

## Other Implementations

- `Rust`: [rs-m3u-parser](https://github.com/pawanpaudel93/rs-m3u-parser)
- `Python`: [m3u-parser](https://github.com/pawanpaudel93/m3u-parser)
- `Typescript`: [ts-m3u-parser](https://github.com/pawanpaudel93/ts-m3u-parser)
- `Lua`: [lua-m3u-parser](https://github.com/pawanpaudel93/lua-m3u-parser)

## Author

👤 **Pawan Paudel**

- Github: [@pawanpaudel93](https://github.com/pawanpaudel93)

## 🤝 Contributing

Contributions, issues and feature requests are welcome! \ Feel free to check [issues page](https://github.com/pawanpaudel93/go-m3u-parser/issues).

## Show your support

Give a ⭐️ if this project helped you!

Copyright © 2020 [Pawan Paudel](https://github.com/pawanpaudel93).
