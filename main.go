package main

import (
	"fmt"

	m3uparser "github.com/pawanpaudel93/go-m3u-parser/m3uparser"
)

func main() {
	// userAgent and timeout is optional. default timeout is 5 seconds and userAgent is latest chrome version 86.
	userAgent := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"
	timeout := 5 // in seconds
	m3uPlaylist := m3uparser.M3uParser{UserAgent: userAgent, Timeout: timeout}
	// file path can also be used /home/pawan/Downloads/ru.m3u
	m3uPlaylist.ParseM3u("https://drive.google.com/uc?id=1VGv8ZYQrrSYPVQ7GCWLgjMl6w9Ccrs4v&export=download", true)
	m3uPlaylist.FilterBy("status", []string{"GOOD"}, true, false)
	m3uPlaylist.SortBy("category", true, false)
	fmt.Println("Saved stream information: ", len(m3uPlaylist.GetStreamsSlice()))
	m3uPlaylist.SaveJSONToFile("pawan.json")
}
