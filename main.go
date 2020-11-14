package main

import (
	"fmt"

	m3uparser "github.com/pawanpaudel93/go-m3u-parser/m3uparser"
)

func main() {
	a := m3uparser.M3uParser{}
	a.ParseM3u("/home/pawan/Downloads/ru.m3u", true)
	a.FilterBy("status", []string{"GOOD"}, true, false)
	a.SortBy("category", true, false)
	fmt.Println("Saved stream information: ", len(a.GetStreamsSlice()))
	a.SaveJSONToFile("pawan.json")
}
