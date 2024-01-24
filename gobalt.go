package gobalt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	CobaltApi    = "https://co.wuk.sh" //Override this value to use your own cobalt instance. See https://instances.hyper.lol/ for alternatives from the main instance.
	UserLanguage = "en-US"             //Replace this following the ISO 639-1 standard. This downloads dubbed YouTube audio according to the language set here. Only takes effect if DubbedYoutubeAudio is set to true.
	useragent    = "Gobalt/1.0"
)

type serverInfo struct {
	Version   string `json:"version"`   //cobalt version
	Commit    string `json:"commit"`    //git commit
	Branch    string `json:"branch"`    //git branch
	Name      string `json:"name"`      //name of the server
	URL       string `json:"url"`       //full url of the api
	Cors      int    `json:"cors"`      //cors status, either 0 or 1.
	StartTime string `json:"startTime"` //server start time in linux epoch
}

type cobaltResponse struct {
	Status string     `json:"status"` //Will be error / redirect / stream / success / rate-limit / picker.
	Picker []struct { //array of picker items
		Type  string `json:"type"`
		URL   string `json:"url"`
		Thumb string `json:"thumb"`
	} `json:"picker"`
	URL  string   `json:"url"`  //Returns the download link. If the status is picker this field will be empty. Direct link to a file or a link to cobalt's live render.
	Text string   `json:"text"` //Various text, mostly used for errors.
	URLs []string //If the status is picker all the urls will go here.
}

type Settings struct {
	Url                   string     //Any URL from bilibili.com, instagram, pinterest, reddit, rutube, soundcloud, streamable, tiktok, tumblr, twitch clips, twitter/x, vimeo, vine archive, vk or youtube. Will be url encoded later.
	VideoCodec            codecs     //H264, AV1 or VP9, defaults to H264.
	VideoQuality          int        //144p to 2160p (4K), if not specified will default to 1080p.
	AudioCodec            audioCodec //MP3, Opus, Ogg or Wav. If not specified will default to best.
	FilenamePattern       pattern    //Classic, Basic, Pretty or Nerdy. Defaults to Pretty
	AudioOnly             bool       //Removes the video, downloads audio only. Default: false
	RemoveTikTokWatermark bool       //Removes TikTok watermark from TikTok videos. Default: false
	FullTikTokAudio       bool       //Enables download of original sound used in a tiktok video. Default: false
	VideoOnly             bool       //Downloads only the video, audio is muted/removed. Default: false
	DubbedYoutubeAudio    bool       //Pass the User-Language HTTP header to use the dubbed audio of the respective language, must change according to user's preference, default is English (US). Uses ISO 639-1 standard.
	DisableVideoMetadata  bool       //Removes file metadata. Default: false
	ConvertTwitterGifs    bool       //Changes whether twitter gifs are converted to .gif (Twitter gifs are usually stored in .mp4 format). Default: true
}

type codecs string

const (
	H264 codecs = "h264" //Default codec that is supported everywhere. Recommended for social media/phones, but tops up at 1080p.
	AV1  codecs = "av1"  //Recent codec, supports 8K/HDR. Generally less supported by media players, social media, etc.
	VP9  codecs = "vp9"  //Best quality codec with higher bitrate (preserve most detail), goes up to 4K and supports HDR.
)

type audioCodec string

const (
	Best audioCodec = "best" //When "best" format is selected, you get audio the way it is on service's side. it's not re-encoded.
	Opus audioCodec = "opus" //Re-encodes the audio using Opus codec. It's a lossy codec with low complexity. Works in Android 10+, Windows 10 1809+, MacOS High Sierra/iOS 17+.
	Ogg  audioCodec = "ogg"  //Re-encodes to ogg, an older lossy audio codec. Should work everywhere.
	Wav  audioCodec = "wav"  //Re-encodes to wav, an even older format. Good compatibility for older systems, like Windows 98. Tops up at 4GiB.
	MP3  audioCodec = "mp3"  //Re-encodes to mp3, the format used basically everywhere. Lossy audio, but generally good player/social media support. Can degrade quality as time passes.
)

type pattern string

const (
	Classic pattern = "classic" //Looks like this: youtube_yPYZpwSpKmA_1920x1080_h264.mp4 | audio: youtube_yPYZpwSpKmA_audio.mp3
	Basic   pattern = "basic"   //Looks like: Video Title (1080p, h264).mp4 | audio: Audio Title - Audio Author.mp3
	Nerdy   pattern = "nerdy"   //Looks like this: Video Title (1080p, h264, youtube).mp4 | audio: Audio Title - Audio Author (soundcloud).mp3
	Pretty  pattern = "pretty"  //Looks like: Video Title (1080p, h264, youtube, yPYZpwSpKmA).mp4 | audio: Audio Title - Audio Author (soundcloud, 1242868615).mp3
)

/*
	Function CreateDefaultSettings() creates the Settings struct with default values:

Url: ""
VideoCodec:            H264,
VideoQuality:          1080,
AudioCodec:            Best,
FilenamePattern:       Pretty,
AudioOnly:             false,
RemoveTikTokWatermark: false,
FullTikTokAudio:       false,
VideoOnly:             false,
DubbedYoutubeAudio:    false,
DisableVideoMetadata:  false,
ConvertTwitterGifs:    false,

You MUST set an url before calling Run().
*/
func CreateDefaultSettings() Settings {

	options := Settings{
		Url:                   "",
		VideoCodec:            H264,
		VideoQuality:          1080,
		AudioCodec:            Best,
		FilenamePattern:       Pretty,
		AudioOnly:             false,
		RemoveTikTokWatermark: false,
		FullTikTokAudio:       false,
		VideoOnly:             false,
		DubbedYoutubeAudio:    false,
		DisableVideoMetadata:  false,
		ConvertTwitterGifs:    true,
	}
	return options
}

// Function Run() requests the final url on /api/json and returns error case it fails to do so.
func Run(opts Settings) (*cobaltResponse, error) {
	validUrl, _ := regexp.MatchString(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`, opts.Url)
	if opts.Url == "" || !validUrl {
		return nil, errors.New("invalid url provided")
	}

	_, err := CobaltServerInfo(CobaltApi)
	if err != nil {
		return nil, fmt.Errorf("could not contact the cobalt server at url %v due of the following error %v", CobaltApi, err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	optionsPayload := map[string]string{"url": url.QueryEscape(opts.Url),
		"vCodec":          string(opts.VideoCodec),
		"vQuality":        fmt.Sprint(opts.VideoQuality),
		"aFormat":         string(opts.AudioCodec),
		"filenamePattern": string(opts.FilenamePattern),
		"isAudioOnly":     fmt.Sprint(opts.AudioOnly),
		"isNoTTWatermark": fmt.Sprint(opts.RemoveTikTokWatermark),
		"isTTFullAudio":   fmt.Sprint(opts.FullTikTokAudio),
		"isAudioMuted":    fmt.Sprint(opts.VideoOnly),
		"dubLang":         fmt.Sprint(opts.DubbedYoutubeAudio),
		"disableMetadata": fmt.Sprint(opts.DisableVideoMetadata),
		"twitterGif":      fmt.Sprint(opts.ConvertTwitterGifs),
	}

	payload, _ := json.Marshal(optionsPayload)

	req, err := http.NewRequest("POST", CobaltApi+"/api/json", strings.NewReader(string(payload)))
	req.Header.Add("User-Agent", useragent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Language", UserLanguage)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	jsonbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var media cobaltResponse
	err = json.Unmarshal(jsonbody, &media)
	if err != nil {
		return nil, err
	}

	if media.Status == "error" || media.Status == "rate-limit" {
		return nil, fmt.Errorf("cobalt error: %v", media.Text)
	}

	if media.Status == "picker" {
		for _, p := range media.Picker {
			media.URLs = append(media.URLs, p.URL)
		}
	} else if media.Status == "stream" {
		media.URLs = append(media.URLs, media.URL)
	}

	return &cobaltResponse{
		Status: media.Status,
		URL:    media.URL,
		Text:   "ok", //Cobalt doesn't return any text if its ok
		URLs:   media.URLs,
	}, nil
}

// This function is called before Run() to check if the cobalt server used is reachable.
// If you can't contact the main server, you should using one of the instances listed in https://instances.hyper.lol/.
func CobaltServerInfo(api string) (*serverInfo, error) {
	//Check if the server is reachable
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	req, err := http.NewRequest("GET", api+"/api/serverInfo", nil)
	req.Header.Add("User-Agent", useragent)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	jsonbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var serverResponse serverInfo
	err = json.Unmarshal(jsonbody, &serverResponse)
	if err != nil {
		return nil, err
	}
	res.Body.Close()
	return &serverInfo{
		Branch:    serverResponse.Branch,
		Commit:    serverResponse.Commit,
		Name:      serverResponse.Name,
		Version:   serverResponse.Version,
		URL:       serverResponse.URL,
		Cors:      serverResponse.Cors,
		StartTime: serverResponse.StartTime,
	}, nil
}
