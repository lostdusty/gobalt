// Package Gobalt provides a go way to communicate with https://cobalt.tools servers.

package gobalt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	CobaltApi    = "https://beta.cobalt.canine.tools"     //Override this value to use your own cobalt instance. See https://instances.hyper.lol/ for alternatives from the main instance.
	UserLanguage = "en"                                   //Replace this following the ISO 639-1 standard. This downloads dubbed YouTube audio according to the language set here. Only takes effect if DubbedYoutubeAudio is set to true.
	Client       = http.Client{Timeout: 10 * time.Second} //This allows you to modify the HTTP Client used in requests. This Client will be re-used.
	useragent    = fmt.Sprintf("Mozilla/5.0 (%v; %v); gobalt/v2.0.0-Alpha (%v; %v); +(https://github.com/lostdusty/gobalt)", runtime.GOOS, runtime.GOARCH, runtime.Compiler, runtime.Version())
)

// ServerInfo is the struct used in the function CobaltServerInfo(). It contains two sub-structs: Cobalt and Git
type ServerInfo struct {
	Cobalt CobaltServerInformation `json:"cobalt"`
	Git    CobaltGitInformation    `json:"git"`
}

// This is ServerInfo.Cobalt struct, it contains information about the cobalt backend running on the server.
type CobaltServerInformation struct {
	Version       string   `json:"version"`       //Cobalt version running.
	URL           string   `json:"url"`           //Backend URL of the cobalt server.
	StartTime     string   `json:"startTime"`     //Time when the server started in Unix miliseconds.
	DurationLimit int      `json:"durationLimit"` //Maximum media lenght you can download in seconds. 10800 seconds = 3 hours.
	Services      []string `json:"services"`      //List of configured/enabled services on the instance.
}

// This is ServerInfo.Git struct, it contains informtions about the git commit (from cobalt) the server is using.
type CobaltGitInformation struct {
	Branch string `json:"branch"` //Git branch the cobalt instance is using.
	Commit string `json:"commit"` //Git commit the cobalt instance is using.
	Remote string `json:"remote"` //Git repository name used by the cobalt instance.
}

// CobaltServerInfo(api) gets the server information and returns ServerInfo struct.
//
// This function is called before Run() to check if the cobalt server used is reachable.
// If you can't contact the main server, try using another instance using GetCobaltinstances().
func CobaltServerInfo(api string) (*ServerInfo, error) {
	//Check if the server is reachable
	req, err := http.NewRequest(http.MethodGet, api+"/serverInfo", nil)
	req.Header.Add("User-Agent", useragent)
	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	jsonbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var serverResponse ServerInfo
	err = json.Unmarshal(jsonbody, &serverResponse)
	if err != nil {
		return nil, err
	}

	return &serverResponse, nil
}

//Server info end

// Cobalt request
type Settings struct {
	Url                string       `json:"url"`                   //Any URL from bilibili.com, instagram, pinterest, reddit, rutube, soundcloud, streamable, tiktok, tumblr, twitch clips, twitter/x, vimeo, vine archive, vk or youtube.
	Mode               downloadMode `json:"downloadMode"`          //Mode to download the videos, either Auto, Audio or Mute. Default: Auto
	VideoFormat        videoCodecs  `json:"youtubeVideoCodec"`     //H264, AV1 or VP9, defaults to H264.
	VideoQuality       int          `json:"videoQuality,string"`   //144p to 2160p (4K), if not specified will default to 1080p.
	AudioFormat        audioCodec   `json:"audioFormat"`           //MP3, Opus, Ogg or Wav. If not specified will default to best.
	AudioBitrate       int          `json:"audioBitrate"`          //Audio Bitrate settings. Values: 320Kbps, 256Kbps, 128Kbps, 96Kbps, 64Kbps or 8Kbps.
	FilenameStyle      pattern      `json:"filenameStyle"`         //Classic, Basic, Pretty or Nerdy. Defaults to Pretty
	TikTokH265         bool         `json:"tiktokH265"`            //Changes whether 1080p h265 [tiktok] videos are preferred or not. Default: false
	FullTikTokAudio    bool         `json:"tiktokFullAudio"`       //Enables download of original sound used in a tiktok video. Default: false
	DubbedYoutubeAudio bool         `json:"youtubeDubBrowserLang"` //Pass the User-Language HTTP header to use the dubbed audio of the respective language, must change according to user's preference, default is English (US). Uses ISO 639-1 standard.
	DisableMetadata    bool         `json:"disableMetadata"`       //Removes file metadata. Default: false
	ConvertTwitterGifs bool         `json:"twitterGif"`            //Changes whether twitter gifs are converted to .gif (Twitter gifs are usually stored in .mp4 format). Default: true
}

type downloadMode string

const (
	Audio downloadMode = "audio" //Download only the audio.
	Auto  downloadMode = "auto"  //Auto mode, audio + video (if video is present).
	Mute  downloadMode = "mute"  //Downloads only the video, no audio.
)

type videoCodecs string

const (
	H264 videoCodecs = "h264" //Default codec that is supported everywhere. Recommended for social media/phones, but tops up at 1080p.
	AV1  videoCodecs = "av1"  //Recent codec, supports 8K/HDR. Generally less supported by media players, social media, etc.
	VP9  videoCodecs = "vp9"  //Best quality codec with higher bitrate (preserve most detail), goes up to 4K and supports HDR.
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
	Nerdy   pattern = "nerdy"   //Looks like this: Video Title (1080p, h264, youtube, yPYZpwSpKmA).mp4 | audio: Audio Title - Audio Author (soundcloud, 1242868615).mp3
	Pretty  pattern = "pretty"  //Looks like: Video Title (1080p, h264, youtube).mp4 | audio: Audio Title - Audio Author (soundcloud).mp3
)

// Cobalt request end

// Cobalt response
type CobaltResponse struct {
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

//Cobalt response end

type CobaltInstances []struct {
	Cors           int    `json:"cors"`             //Cors status: 0 = Enabled; 1 = Disabled; -1 = Instance offline.
	Commit         string `json:"commit,omitempty"` //Commit id. Empty if the instance is offline.
	Name           string `json:"name,omitempty"`   //Name of the server. Empty if the instance is offline.
	StartTime      int64  `json:"startTime"`        //Time when the service started in linux epoch (seconds). -1 Means the instance is offline
	API            string `json:"api"`              //API Url.
	Version        string `json:"version"`          //Version of cobalt running, "-1" if offiline.
	Branch         string `json:"branch,omitempty"` //Branch the server is using, empty if the server is offline
	FrontEnd       string `json:"frontEnd"`         //Front end url.
	ApiOnline      bool   `json:"api_online"`       //Status of the server api. True if online.
	FrontEndOnline bool   `json:"frontend_online"`  //Status of the frontend. Online = true.
}

type MediaInfo struct {
	Size uint //Media size in bytes
	Name string
	Type string
}

// CreateDefaultSettings Function CreateDefaultSettings() creates the Settings struct with default values:
// Url: ""
// VideoCodec:            H264,
// VideoQuality:          1080,
// AudioCodec:            Best,
// FilenamePattern:       Pretty,
// AudioOnly:             false,
// FullTikTokAudio:       false,
// VideoOnly:             false,
// DubbedYoutubeAudio:    false,
// DisableVideoMetadata:  false,
// ConvertTwitterGifs:    false,
// TikTokH265:			  false,
// You MUST set an url before calling Run().
func CreateDefaultSettings() Settings {

	options := Settings{
		Url:                "",
		VideoFormat:        H264,
		VideoQuality:       1080,
		AudioFormat:        Best,
		FilenameStyle:      Pretty,
		ConvertTwitterGifs: true,
	}
	return options
}

// Run Function Run() requests the final url on /api/json and returns error case it fails to do so.
func Run(opts Settings) (*CobaltResponse, error) {
	validUrl, _ := regexp.MatchString(`[-a-zA-Z0-9@:%_+.~#?&/=]{2,256}\.[a-z]{2,4}\b(/[-a-zA-Z0-9@:%_+.~#?&/=]*)?`, opts.Url)
	if opts.Url == "" || !validUrl {
		return nil, errors.New("invalid url provided")
	}

	_, err := CobaltServerInfo(CobaltApi)
	if err != nil {
		return nil, fmt.Errorf("could not contact the cobalt server at url %v due of the following error %v", CobaltApi, err)
	}

	optionsPayload := Settings{
		Url:                url.QueryEscape(opts.Url),
		VideoFormat:        opts.VideoFormat,
		VideoQuality:       opts.VideoQuality,
		AudioFormat:        opts.AudioFormat,
		FilenameStyle:      opts.FilenameStyle,
		TikTokH265:         opts.TikTokH265,
		FullTikTokAudio:    opts.FullTikTokAudio,
		DubbedYoutubeAudio: opts.DubbedYoutubeAudio,
		DisableMetadata:    opts.DisableMetadata,
		ConvertTwitterGifs: opts.ConvertTwitterGifs,
	}
	payload, _ := json.Marshal(optionsPayload)

	req, err := http.NewRequest(http.MethodPost, CobaltApi+"/api/json", strings.NewReader(string(payload)))
	req.Header.Add("User-Agent", useragent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Language", UserLanguage)
	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	jsonbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var media CobaltResponse
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

	return &CobaltResponse{
		Status: media.Status,
		URL:    media.URL,
		Text:   "ok", //Cobalt doesn't return any text if it is ok
		URLs:   media.URLs,
	}, nil
}

// GetCobaltInstances makes a request to instances.hyper.lol and returns a list of all online cobalt instances.
/*func GetCobaltInstances() ([]ServerInfo, error) {
	req, err := http.NewRequest(http.MethodGet, "https://instances.hyper.lol/instances.json", nil)
	req.Header.Add("User-Agent", useragent)
	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	jsonbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cobaltHyperInstances CobaltInstances
	err = json.Unmarshal(jsonbody, &cobaltHyperInstances)
	if err != nil {
		return nil, err
	}

	instancesList := make([]ServerInfo, 0)

	for _, v := range cobaltHyperInstances {

		if v.ApiOnline {
			instancesList = append(instancesList, ServerInfo{
				Version:        v.Version,
				Commit:         v.Commit,
				Branch:         v.Branch,
				Name:           v.Name,
				URL:            v.API,
				Cors:           v.Cors,
				StartTime:      v.StartTime,
				FrontendUrl:    v.FrontEnd,
				ApiOnline:      v.ApiOnline,
				FrontEndOnline: v.FrontEndOnline,
			})
		}
	}
	return instancesList, nil
}*/

func ProcessMedia(url string) (*MediaInfo, error) {
	req, err := http.Head(url)
	if err != nil {
		return nil, err
	}
	_, parsefilename, err := mime.ParseMediaType(req.Header.Get("Content-Disposition"))
	filename := parsefilename["filename"]
	if err != nil {
		filename = path.Base(req.Request.URL.Path)
	}
	size := req.Header.Get("Content-Length")
	if size == "" {
		size = "0"
	}
	parseSize, err := strconv.Atoi(size)
	if err != nil {
		return nil, err
	}

	return &MediaInfo{
		Size: uint(parseSize),
		Name: filename,
		Type: req.Header.Get("Content-Type"),
	}, nil
}
