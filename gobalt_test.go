package gobalt

import (
	"net/url"
	"regexp"
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

func TestPlaylistGetter(t *testing.T) {
	v, err := GetPlaylist("https://youtube.com/playlist?list=PLDKxz_KUEUfOJDeQ_KeQxuxG8kRRcXrWs&si=1ZfoNPcDyhum6exn")
	if err != nil {
		t.Fatalf("failed to get playlist: %v", err)
	}
	for _, p := range v {
		t.Logf("Found music \"%v\" by %v (%v)", p.VideoTitle, p.VideoUploader, p.VideoURL)
	}
}

func TestYoutubeDownload(t *testing.T) {
	var e decryptor
	v, err := getVideo(&e, "https://youtube.com/watch?v=lDoXekDxHIU")
	if err != nil {
		t.Fatalf("unable to download due of %v", err)
	}
	t.Logf("stream url: %v", v.StreamUrl)
}

// Benchmarks
func BenchmarkRegexUrlParse(b *testing.B) {
	a, _ := regexp.MatchString(`[-a-zA-Z0-9@:%_+.~#?&/=]{2,256}\.[a-z]{2,4}\b(/[-a-zA-Z0-9@:%_+.~#?&/=]*)?`, "https://www.youtube.com/watch?v=b3rFbkFjRrA")
	if a {
		b.Log("regex pass")
	}
}

func BenchmarkNetUrlParse(b *testing.B) {
	_, err := url.Parse("https://www.youtube.com/watch?v=b3rFbkFjRrA")
	if err != nil {
		b.Fatalf("error parsing url: %v", err)
	}
}
