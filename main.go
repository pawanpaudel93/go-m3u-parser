package main

import (
	"fmt"

	m3uparser "github.com/pawanpaudel93/go-m3u-parser/m3uparser"
)

func main() {
	a := m3uparser.M3uParser{}
	a.ParseM3u("/home/pawan/Downloads/ru.m3u", false)
	fmt.Printf("%v", len(a.StreamsInfo))
	// a.FilterBy("tvg-id", []string{""}, false, true)
	fmt.Printf("%v", len(a.StreamsInfo))
	// fmt.Println("helo", a.GetStreamsJson())
	a.SaveJsonToFile("pawan.json")
}
