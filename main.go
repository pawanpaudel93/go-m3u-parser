package main

import (
	"fmt"

	m3uparser "github.com/pawanpaudel93/go-m3u-parser/m3uparser"
)

func main() {
	a := m3uparser.M3uParser{}
	a.ParseM3u("/home/pawan/Downloads/ru.m3u", true)
	fmt.Printf("%v", a.StreamsInfo)
}
