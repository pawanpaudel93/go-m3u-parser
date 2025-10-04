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

```go
 userAgent := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
 timeout := 5 // in seconds
 parser := m3uparser.M3uParser{UserAgent: userAgent, Timeout: timeout}
 // file path can also be used /home/pawan/Downloads/ru.m3u
 parser.ParseM3u("https://drive.google.com/uc?id=1VGv8ZYQrrSYPVQ7GCWLgjMl6w9Ccrs4v&export=download", true, true)
```

>Functions

```go
func (p *M3uParser) ParseM3u(path string, checkLive bool, enforceSchema bool) {

        """Parses the content of local file/URL.
        It downloads the file from the given url or use the local file path to get the content and parses line by line
        to a structured format of streams information.
  
        path: Path can be a url or local filepath
        checkLive: To check if the stream links are working or not
        enforceSchema: If true it keeps all the fields else it removes the key having empty string values.
        """
  
}
 
func (p *M3uParser) FilterBy(key string, filters []string, retrieve bool) {

        """Filter streams infomation.
        It retrieves/removes stream information from streams information list using filter/s on key.

        key: Key can be single or nested. eg. key='name', key='language-name'
        filters: List of filter/s to perform the retrieve or remove operation.
        retrieve: True to retrieve and False for removing based on key.
        """
  
}
  
func (p *M3uParser) ResetOperations() {

        """Reset the stream information list to initial state before various operations."""
  
}
  
func (p *M3uParser) RemoveByExtension(extension []string) {

        """Remove stream information with certain extension/s.
        It removes stream information from streams information list based on extension/s provided.
  
        extension: Name of the extension like mp4, m3u8 etc. It is slice of extension/s.
        """
  
}
  
func (p *M3uParser) RetrieveByExtension(extension []string) {

        """Select only streams information with a certain extension/s.
        It retrieves the stream information based on extension/s provided.
  
        extension: Name of the extension like mp4, m3u8 etc. It is slice of extension/s.
        """
}
  
func (p *M3uParser) RemoveByCategory(category []string) {

        """Removes streams information with category containing a certain filter word/s.
        It removes stream information based on category using filter word/s.
  
        category: It is slice of category/categories.
  
        """
}
  
func (p *M3uParser) RetrieveByCategory(category []string) {

        """Retrieve only streams information that contains a certain filter word/s.
        It retrieves stream information based on category/categories.

        category: It is slice of category/categories.
        """
}
  
func (p *M3uParser) SortBy(key string, asc bool) {

        """Sort streams information.
        It sorts streams information list sorting by key in asc/desc order.

        key: It can be single or nested key.
        asc: Sort by asc or desc order.
        """
}

func (p *M3uParser) GetStreamsJSON() string {

        """Get the streams information as json."""
}
  
func (p *M3uParser) GetStreamsSlice() []Channel {

        """Get the parsed streams information list.
        It returns the streams information slice.
        """
  
func (p *M3uParser) GetRandomStream(shuffle bool) Channel {

        """Return a random stream information
        It returns a random stream information with shuffle if required.

        shuffle: To shuffle the streams information list before returning the random stream information.
        """
}
  
func (p *M3uParser) ToFile(filename string) {

        """Save to json/m3u file.
        It saves streams information as a JSON/M3U file with a given filename.

        filename: Name of the file to save streams information.
        """
}

```

## Other Implementations

- `Rust`: [rs-m3u-parser](https://github.com/pawanpaudel93/rs-m3u-parser)
- `Python`: [m3u-parser](https://github.com/pawanpaudel93/m3u-parser)
- `Typescript`: [ts-m3u-parser](https://github.com/pawanpaudel93/ts-m3u-parser)

## Author

👤 **Pawan Paudel**

- Github: [@pawanpaudel93](https://github.com/pawanpaudel93)

## 🤝 Contributing

Contributions, issues and feature requests are welcome! \ Feel free to check [issues page](https://github.com/pawanpaudel93/go-m3u-parser/issues).

## Show your support

Give a ⭐️ if this project helped you!

Copyright © 2020 [Pawan Paudel](https://github.com/pawanpaudel93).
