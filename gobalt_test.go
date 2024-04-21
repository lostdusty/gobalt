package gobalt

import (
	"net/url"
	"testing"
)

func TestCobaltDownload(t *testing.T) {
	dlTest := CreateDefaultSettings()
	dlTest.Url = "https://www.youtube.com/watch?v=b3rFbkFjRrA"
	dlTest.AudioCodec = Ogg
	dlTest.VideoCodec = VP9
	dlTest.FilenamePattern = Nerdy
	runDlTest, err := Run(dlTest)
	if err != nil {
		t.Fatalf("Failed to fetch from cobalt, got %v", err)
	}
	//Check if url is co.wuk.sh
	parseDlTest, err := url.Parse(runDlTest.URL)
	if err != nil || parseDlTest.Host == "co.wuk.sh" {
		t.Fatalf("Failed to parse url from cobalt. Expected !nil and got %v", err)
	}
}

func TestCustomInstancesList(t *testing.T) {
	instanceTest, err := GetCobaltInstances()
	if err != nil {
		t.Fatalf("Failed to get the list of cobalt instances: %v", err)
	}
	for n, v := range instanceTest {
		if v.URL == "" {
			t.Fatalf("Failed to lookup %v's cobalt instance #%v, no host url present.", v.Name, n+1)
		}
	}
}

func TestHealthMainInstance(t *testing.T) {
	testHealth, err := CobaltServerInfo(CobaltApi)
	if err != nil || testHealth.StartTime == 0 {
		t.Fatalf("bad health of %v instance. got %v", CobaltApi, err)
	}

}
