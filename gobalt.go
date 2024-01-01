package gobalt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	CobaltApi = "https://co.wuk.sh" //Override this value to use your own cobalt instance. See https://instances.hyper.lol/ for alternatives from the main instance.
	useragent = "Gobalt/1.0"
)

type settings struct {
	url                   string
	videoCodec            string
	videoQuality          string
	audioCodec            string
	filenamePattern       string
	audioOnly             bool
	removeTikTokWatermark bool
	fullTikTokAudio       bool
	removeAudio           bool
	dubbedYoutubeAudio    bool
	disableVideoMetadata  bool
}

type serverInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Branch    string `json:"branch"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Cors      int    `json:"cors"`
	StartTime string `json:"startTime"`
}

type cobaltResponse struct {
	Status string `json:"status"`
	URL    string `json:"url"`
	Text   string `json:"text"`
}

// Define Option type as a function that modifies settings
type Option func(*settings)

/*
	Video Codecs

h264, av1 or vp9. default is h264. applies only to youtube downloads. `h264` is recommended for phones.
*/
func VideoCodec(codec string) Option {
	//Expecting H264, AV1 or VP9
	supportedCodecsAudio := []string{"h264", "av1", "vp9"}
	if !slices.Contains(supportedCodecsAudio, strings.ToLower(codec)) { //Unknown codec, will use h264
		return func(s *settings) { s.videoCodec = "h264" }
	}
	return func(s *settings) { s.videoCodec = codec }
}

/*
	Audio Codecs

MP3, Opus, ogg, wav or best, defaults to `best`.
*/
func AudioCodec(codec string) Option {
	//Expecting MP3, Opus, ogg, wav or best
	supportedCodecsAudio := []string{"mp3", "ogg", "opus", "wav", "best"}
	if !slices.Contains(supportedCodecsAudio, strings.ToLower(codec)) { //Unknown codec, will try the best one
		return func(s *settings) { s.audioCodec = "best" }
	}
	return func(s *settings) { s.audioCodec = codec }
}

// End of audio codecs

func VideoQuality(quality string) Option {
	//Expecting any value between 144p and 4K, or best
	supportedQuality := []string{"144", "240", "360", "480", "720", "1080", "1440", "2160", "best"}
	if !slices.Contains(supportedQuality, quality) { //Unsupported quality, will return best
		return func(s *settings) { s.videoQuality = "best" }
	}
	return func(s *settings) { s.videoQuality = quality }
}

// FilenamePattern types
func FilenamePattern(pattern string) Option {
	//Expected: classic, pretty, nerdy or basic
	supportedPatterns := []string{"classic", "pretty", "nerdy", "basic"}
	/* Pattern examples:
	 * Classic: youtube_yPYZpwSpKmA_1920x1080_h264.mp4            | audio: youtube_yPYZpwSpKmA_audio.mp3
	 * Basic: Video Title (1080p, h264).mp4                       | audio: Audio Title - Audio Author.mp3
	 * Pretty: Video Title (1080p, h264, youtube).mp4             | audio: Audio Title - Audio Author (soundcloud).mp3
	 * Nerdy: Video Title (1080p, h264, youtube, yPYZpwSpKmA).mp4 | audio: Audio Title - Audio Author (soundcloud, 1242868615).mp3
	 */
	if !slices.Contains(supportedPatterns, pattern) { //Unknown pattern will return basic
		return func(s *settings) { s.filenamePattern = "basic" }
	}
	return func(s *settings) { s.filenamePattern = pattern }
}

func AudioOnly(value bool) Option {
	return func(s *settings) { s.audioOnly = value }
}

func RemoveTikTokWatermark(value bool) Option {
	return func(s *settings) { s.removeTikTokWatermark = value }
}

func FullTikTokAudio(value bool) Option {
	return func(s *settings) { s.fullTikTokAudio = value }
}

func RemoveAudio(value bool) Option {
	return func(s *settings) { s.removeAudio = value }
}

func DubbedYoutubeAudio(value bool) Option {
	return func(s *settings) { s.dubbedYoutubeAudio = value }
}

func DisableVideoMetadata(value bool) Option {
	return func(s *settings) { s.disableVideoMetadata = value }
}

// New initializes a settings object with default values and applies any provided options.
func New(url string, options ...Option) *settings {
	if url == "" {
		panic("url cannot be empty")
	}
	defaults := &settings{
		url:                   url,
		videoCodec:            "h264",
		audioCodec:            "mp3",
		videoQuality:          "best",
		filenamePattern:       "basic",
		audioOnly:             false,
		removeTikTokWatermark: false,
		fullTikTokAudio:       false,
		removeAudio:           false,
		dubbedYoutubeAudio:    false,
		disableVideoMetadata:  false,
	}

	for _, option := range options {
		option(defaults)
	}

	return defaults
}

func Run(opts *settings) (*cobaltResponse, error) {
	//I was planning to handle other cobalt instances, but i think this should be left to the user.
	validUrl, _ := regexp.MatchString(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`, opts.url)
	if !validUrl {
		return nil, errors.New("invalid url provided")
	}
	_, err := CobaltServerInfo(CobaltApi)
	if err != nil {
		return nil, fmt.Errorf("could not contact the cobalt server at url %v due of the following error %v", CobaltApi, err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	optionsPayload := map[string]string{"url": url.QueryEscape(opts.url),
		"vCodec":          opts.videoCodec,
		"vQuality":        opts.videoQuality,
		"aFormat":         opts.audioCodec,
		"filenamePattern": opts.filenamePattern,
		"isAudioOnly":     fmt.Sprint(opts.audioOnly),
		"isNoTTWatermark": fmt.Sprint(opts.removeTikTokWatermark),
		"isTTFullAudio":   fmt.Sprint(opts.fullTikTokAudio),
		"isAudioMuted":    fmt.Sprint(opts.removeAudio),
		"dubLang":         fmt.Sprint(opts.dubbedYoutubeAudio),
		"disableMetadata": fmt.Sprint(opts.disableVideoMetadata),
	}

	payload, _ := json.Marshal(optionsPayload)
	//fmt.Println(string(payload))

	req, err := http.NewRequest("POST", CobaltApi+"/api/json", strings.NewReader(string(payload)))
	req.Header.Add("User-Agent", useragent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
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
		return nil, errors.New("not implemented on gobalt yet, please open an issue")
	}

	return &cobaltResponse{
		Status: media.Status,
		URL:    media.URL,
		Text:   "ok", //Cobalt doesn't return any text if its ok
	}, nil
}

func CobaltServerInfo(url string) (*serverInfo, error) {
	//Check if the server is reachable
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	req, err := http.NewRequest("GET", url+"/api/serverInfo", nil)
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
