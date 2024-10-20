package gobalt

import (
	"math/rand/v2"
	"testing"
)

func TestCobaltDownload(t *testing.T) {
	dlTest := CreateDefaultSettings()
	dlTest.Url = "https://www.youtube.com/watch?v=bV68_Vy0Uis&list=RD-Sr668sSEIA&index=19"
	dlTest.AudioFormat = Ogg
	dlTest.YoutubeVideoFormat = VP9
	CobaltApi = "https://cobalt-api.kwiatekmiki.com"
	runDlTest, err := Run(dlTest)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Log(runDlTest.URL)
}

func TestCustomInstancesList(t *testing.T) {
	instanceTest, err := GetCobaltInstances()
	if err != nil || len(instanceTest) == 0 {
		if len(instanceTest) == 0 {
			t.Log("Looks like no v10.0.0 instance was found, this test will be skipped.")
			t.SkipNow() //Skips this test
		}
		t.Fatalf("Failed to get the list of cobalt instances. Either theres no instances found, or something else went wrong.\nErr: %v, instances found: %v", err, len(instanceTest))
	}
	t.Logf("Found %v instances!\n", len(instanceTest))
	randomInstanceToTest := rand.IntN(len(instanceTest))
	t.Logf("Will test instance #%v", randomInstanceToTest)
	testHealthRandomInstance, err := CobaltServerInfo(instanceTest[randomInstanceToTest].API)
	if err != nil {
		t.Logf("unable to test api selected due of %v", err)
	}
	t.Logf("Sucessfully accessed the instance %v, running cobalt %v", testHealthRandomInstance.Cobalt.URL, testHealthRandomInstance.Cobalt.Version)
}

func TestHealthMainInstance(t *testing.T) {
	testHealth, err := CobaltServerInfo(CobaltApi)
	if err != nil {
		t.Fatalf("bad health of %v instance. got %v", CobaltApi, err)
	}
	t.Log(testHealth.Cobalt.URL)

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
	t.Logf("name %v | size %v bytes | mime %v", d.Filename, n.Size, n.Type)

}

func TestPlaylistGet(t *testing.T) {
	a, err := GetYoutubePlaylist("https://youtube.com/playlist?list=PLDKxz_KUEUfMDTqDgv4eHuZq1u_SQtRiu&si=a-f1kK5lSGFRJO8z")
	if err != nil {
		t.Fatalf("failed to get playlist: %v", err)
	}
	if a[0] != "https://youtu.be/gYygotHLyjo" {
		t.Fatalf("got unexpected link: %v, instead of https://youtu.be/gYygotHLyjo", a[0])
	}
	t.Log("Urls of the playlist:")
	for _, v := range a {
		t.Log(v)
	}
}
