package gobalt

import (
	"net/url"
	"regexp"
	"testing"
)

func TestCobaltDownload(t *testing.T) {
	dlTest := CreateDefaultSettings()
	//CobaltApi = "https://beta.cobalt.canine.tools/"
	dlTest.Url = "https://www.youtube.com/watch?v=b3rFbkFjRrA"
	dlTest.AudioCodec = Ogg
	dlTest.VideoCodec = VP9
	dlTest.FilenamePattern = Nerdy
	runDlTest, err := Run(dlTest)
	if err != nil {
		t.Fatalf("Failed to fetch from cobalt, got %v", err)
	}

	t.Log(runDlTest.URL)
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

func TestMediaParsing(t *testing.T) {
	v := CreateDefaultSettings()
	v.Url = "https://music.youtube.com/watch?v=JCd4KENZyj4"
	d, err := Run(v)
	if err != nil {
		t.Fatalf("failed getting media because %v", err)
	}
	n, err := ProcessMedia(d.URL)
	if err != nil {
		t.Fatalf("failed processing media because %v", err)
	}
	t.Logf("name %v | size %v bytes | mime %v", n.Name, n.Size, n.Type)

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
