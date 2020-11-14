package m3uparser

import (
	"os"
	"testing"
)

func TestParseM3uFile(t *testing.T) {
	parser := M3uParser{}
	parser.ParseM3u("/home/pawan/Downloads/ru.m3u", false)
	if len(parser.streamsInfo) == 0 {
		t.Error("No streams parsed.")
	}
}

func TestParseM3uURL(t *testing.T) {
	parser := M3uParser{}
	parser.ParseM3u("https://drive.google.com/uc?id=1VGv8ZYQrrSYPVQ7GCWLgjMl6w9Ccrs4v&export=download", false)
	if len(parser.streamsInfo) == 0 {
		t.Error("No streams parsed.")
	}
}

func TestSaveJSONToFile(t *testing.T) {
	parser := M3uParser{}
	parser.ParseM3u("/home/pawan/Downloads/ru.m3u", false)
	parser.SaveJSONToFile("testFile.json")
	if _, err := os.Stat("testFile.json"); err != nil {
		t.Error("File is not saved.")
	} else if !os.IsNotExist(err) {
		os.Remove("testFile.json")
	}
}
