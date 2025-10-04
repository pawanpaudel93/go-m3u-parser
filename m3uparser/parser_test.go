package m3uparser

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestParseM3uFile(t *testing.T) {
	// Download the M3U file from GitHub
	url := "https://raw.githubusercontent.com/iptv-org/iptv/refs/heads/master/streams/np.m3u"
	fileName := "np_test.m3u"

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to download file, status code: %d", resp.StatusCode)
	}

	// Create local file
	file, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Copy content to local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	file.Close()

	// Clean up the file after test
	defer func() {
		if err := os.Remove(fileName); err != nil {
			t.Logf("Warning: Failed to remove test file: %v", err)
		}
	}()

	// Test parsing the downloaded file
	parser := M3uParser{}
	parser.ParseM3u(fileName, false, false)
	if len(parser.streamsInfo) == 0 {
		t.Error("No streams parsed.")
	}

	// Verify we got some expected content
	fmt.Printf("Parsed %d streams from downloaded file\n", len(parser.streamsInfo))
}

func TestParseM3uURL(t *testing.T) {
	parser := M3uParser{}
	parser.ParseM3u("https://raw.githubusercontent.com/iptv-org/iptv/refs/heads/master/streams/np.m3u", false, false)
	if len(parser.streamsInfo) == 0 {
		t.Error("No streams parsed.")
	}
}

func TestSaveJSONToFile(t *testing.T) {
	// Download the M3U file from GitHub
	url := "https://raw.githubusercontent.com/iptv-org/iptv/refs/heads/master/streams/np.m3u"
	fileName := "np_test_save.m3u"
	jsonFileName := "testFile.json"

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to download file, status code: %d", resp.StatusCode)
	}

	// Create local file
	file, err := os.Create(fileName)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Copy content to local file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	file.Close()

	// Clean up files after test
	defer func() {
		if err := os.Remove(fileName); err != nil {
			t.Logf("Warning: Failed to remove test file: %v", err)
		}
		if err := os.Remove(jsonFileName); err != nil {
			t.Logf("Warning: Failed to remove JSON file: %v", err)
		}
	}()

	// Test parsing and saving to JSON
	parser := M3uParser{}
	parser.ParseM3u(fileName, false, false)
	parser.ToFile(jsonFileName)

	if _, err := os.Stat(jsonFileName); err != nil {
		t.Error("JSON file is not saved.")
	}
}

func TestParseM3uFromRawContent(t *testing.T) {
	// Sample M3U content for testing
	m3uContent := `#EXTM3U
#EXTINF:-1 tvg-id="CapitalTVHD.np",Capital TV (1080p)
https://streaming.tvnepal.com:19360/capitaltv/capitaltv.m3u8
#EXTINF:-1 tvg-id="DivyaDarshanTV.np",Divya Darshan TV (720p)
http://live.divyadarshantv.com/hls/stream.m3u8
#EXTINF:-1 tvg-id="IndigenousTelevision.np",Indigenous Television (720p)
https://np.truestreamz.com/broadcaster/INDIGENOUSmob.stream/playlist.m3u8
#EXTINF:-1 tvg-id="MithilaNepalTV.np",Mithila Nepal TV (1080p)
http://150.107.205.212:1935/live/mithila/playlist.m3u8?DVR=
#EXTINF:-1 tvg-id="ParyawaranTV.np",Paryawaran TV (1080p)
https://webtv-stream.nettv.com.np/broadcaster/Paryawaran.stream/playlist.m3u8
#EXTINF:-1 tvg-id="KantipurTV.np",Kantipur TV
https://ktvhdsg.ekantipur.com:8443/high_quality_85840165/hd/playlist.m3u8
#EXTINF:-1 tvg-id="KantipurMax.np",Kantipur Max
https://ktvhdsg.ekantipur.com:8443/ktvmax2025/match1/playlist.m3u8`

	parser := M3uParser{}
	parser.ParseM3u(m3uContent, false, false)

	if len(parser.streamsInfo) == 0 {
		t.Error("No streams parsed from raw content.")
	}

	if len(parser.streamsInfo) != 7 {
		t.Errorf("Expected 7 streams, got %d", len(parser.streamsInfo))
	}

	// Create maps to check content without worrying about order
	titles := make(map[string]bool)
	categories := make(map[string]bool)
	urls := make(map[string]bool)

	for _, stream := range parser.streamsInfo {
		if title, ok := stream["title"].(string); ok {
			titles[title] = true
		}
		if category, ok := stream["category"].(string); ok {
			categories[category] = true
		}
		if url, ok := stream["url"].(string); ok {
			urls[url] = true
		}
	}

	// Check that some expected titles are present (based on the actual GitHub content)
	if !titles["Capital TV (1080p)"] {
		t.Error("Expected 'Capital TV (1080p)' title not found")
	}
	if !titles["Divya Darshan TV (720p)"] {
		t.Error("Expected 'Divya Darshan TV (720p)' title not found")
	}

	// Check that some expected URLs are present
	if !urls["https://streaming.tvnepal.com:19360/capitaltv/capitaltv.m3u8"] {
		t.Error("Expected Capital TV URL not found")
	}
	if !urls["http://live.divyadarshantv.com/hls/stream.m3u8"] {
		t.Error("Expected Divya Darshan TV URL not found")
	}
}

func TestParseM3uFromRawContentEmpty(t *testing.T) {
	parser := M3uParser{}
	parser.ParseM3u("", false, false)

	if len(parser.streamsInfo) != 0 {
		t.Error("Expected no streams for empty content.")
	}
}

func TestParseM3uFromRawContentInvalid(t *testing.T) {
	parser := M3uParser{}
	parser.ParseM3u("This is not a valid M3U file\nwith multiple lines\nbut no #EXTM3U", false, false)

	if len(parser.streamsInfo) != 0 {
		t.Error("Expected no streams for invalid M3U content.")
	}
}
