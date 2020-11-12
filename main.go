package main

import (
	m3uparser "github.com/pawanpaudel93/go-m3u-parser/m3uparser"
)

func main() {
	a := m3uparser.M3uParser{}
	a.ParseM3u("/home/pawan/Downloads/ru.m3u", false)
	a.SaveJSONToFile("pawan.json")
}
