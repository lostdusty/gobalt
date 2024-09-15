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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mcuadros/go-version"
)

var (
	CobaltApi = "https://cobalt-backend.canine.tools"  //Override this value to use your own cobalt instance. See https://instances.hyper.lol/ for alternatives from the main instance.
	Client    = http.Client{Timeout: 10 * time.Second} //This allows you to modify the HTTP Client used in requests. This Client will be re-used.
	useragent = fmt.Sprintf("gobalt/2.0.1 (+https://github.com/lostdusty/gobalt/v2; go/%v; %v/%v)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
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
	//Parse url before testing, sanity check
	parseApiUrl, err := url.Parse(api)
	if err != nil {
		return nil, fmt.Errorf("net/url failed to parse provided url, check it and try again (details: %v)", err)
	}

	if parseApiUrl.Scheme == "" {
		parseApiUrl.Scheme = "https"
	}

	//Check if the server is reachable
	req, err := http.NewRequest(http.MethodGet, parseApiUrl.String(), nil)
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

/* Download settings structs and types */

// Struct Settings contains changable options that you can change before download. An URL MUST be set before calling gobalt.Run(Settings).
type Settings struct {
	Url                   string       `json:"url"`                   //Any URL from bilibili.com, instagram, pinterest, reddit, rutube, soundcloud, streamable, tiktok, tumblr, twitch clips, twitter/x, vimeo, vine archive, vk or youtube (as long it's configured on the instance).
	Mode                  downloadMode `json:"downloadMode"`          //Mode to download the videos, either Auto, Audio or Mute. Default: Auto
	Proxy                 bool         `json:"alwaysProxy"`           //Tunnel downloaded file thru cobalt, bypassing potential restrictions and protecting your identity and privacy. Default: false
	AudioBitrate          int          `json:"audioBitrate,string"`   //Audio Bitrate settings. Values: 320Kbps, 256Kbps, 128Kbps, 96Kbps, 64Kbps or 8Kbps. Default: 128
	AudioFormat           audioCodec   `json:"audioFormat"`           //"Best", .mp3, .opus, .ogg or .wav. If not specified will default to "Best".
	FilenameStyle         pattern      `json:"filenameStyle"`         //"Classic", "Basic", "Pretty" or "Nerdy". Default is "Basic".
	DisableMetadata       bool         `json:"disableMetadata"`       //Don't include file metadata. Default: false
	TikTokH265            bool         `json:"tiktokH265"`            //Allows downloading TikTok videos in 1080p at cost of compatibility. Default: false
	TikTokFullAudio       bool         `json:"tiktokFullAudio"`       //Enables download of original sound used in a TikTok video. Default: false
	TwitterConvertGif     bool         `json:"twitterGif"`            //Changes whether twitter gifs should be converted to .gif (Twitter gifs are usually looping .mp4s). Default: true
	VideoQuality          int          `json:"videoQuality,string"`   //144p to 2160p (4K), if not specified will default to 1080p.
	YoutubeDubbedAudio    bool         `json:"youtubeDubBrowserLang"` //Downloads the YouTube dubbed audio according to the value set in YoutubeDubbedLanguage (and if present). Default is English (US). Follows the ISO 639-1 standard.
	YoutubeDubbedLanguage string       `json:"youtubeDubLang"`        //Language code to download the dubbed audio, Default is "en".
	YoutubeVideoFormat    videoCodecs  `json:"youtubeVideoCodec"`     //Which video format to download from YouTube, see videoCodecs type for details.
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

// This function creates the Settings struct with these default values:
//
//   - Url: "" (empty)
//   - YoutubeVideoFormat: `H264`
//   - VideoQuality: `1080`
//   - AudioFormat: `Best`
//   - AudioBitrate: `128`
//   - FilenameStyle: `Basic`
//   - TwitterConvertGif: `true`
//   - Mode: `Auto`
//
// You MUST set an url before calling Run().
func CreateDefaultSettings() Settings {
	options := Settings{
		Url:                   "",
		YoutubeVideoFormat:    H264,
		VideoQuality:          1080,
		AudioFormat:           Best,
		AudioBitrate:          128,
		FilenameStyle:         Basic,
		TwitterConvertGif:     true,
		Mode:                  Auto,
		YoutubeDubbedLanguage: "en",
	}
	return options
}

// Cobalt response to your request
type CobaltResponse struct {
	Status string      `json:"status"` //4 possible status. Error = Something went wrong, see CobaltResponse.Error.Code | Tunnel or Redirect = Everything is right. | Picker = Multiple media, see CobaltResponse.Picker.
	Picker *[]struct { //This is an array of items, each containing the media type, url to download and thumbnail.
		Type  string `json:"type"`  //Type of the media, either photo, video or gif
		URL   string `json:"url"`   //Url to download.
		Thumb string `json:"thumb"` //Media preview url, optional.
	} `json:"picker"`
	URL      string `json:"url"`      //Returns the download link. If the status is picker this field will be empty. Direct link to a file or a link to cobalt's live render.
	Filename string `json:"filename"` //Various text, mostly used for errors.
	Error    *Error `json:"error"`    //Error information, may be <NIL> if theres no error.
}

type Error struct {
	Code    string  `json:"code"`    // Machine-readable error code explaining the failure reason.
	Context Context `json:"context"` //(optional) container for providing more context.
}

type Context struct {
	Service string `json:"service"`         //What service failed.
	Limit   int    `json:"limit,omitempty"` //Number providing the ratelimit maximum number of requests, or maximum downloadable video duration
}

// Run(gobalt.Settings) sends the request to the provided cobalt api and returns the server response (gobalt.CobaltResponse) and error, use this to download something AFTER setting your desired configuration.
func Run(options Settings) (*CobaltResponse, error) {
	//Check if an url is set.
	if options.Url == "" {
		return nil, errors.New("no url was provided in Settings.Url")
	}

	//Do a basic check to see if the server is online and handling requests
	_, err := CobaltServerInfo(CobaltApi)
	if err != nil {
		return nil, fmt.Errorf("hello to cobalt instance %v failed, reason: %v", CobaltApi, err)
	}

	jsonBody, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json body due of the following error: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, CobaltApi, strings.NewReader(string(jsonBody)))
	req.Header.Add("User-Agent", useragent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send your request, %v", err)
	}
	defer res.Body.Close()

	jsonbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var media CobaltResponse
	err = json.Unmarshal(jsonbody, &media)
	if err != nil {
		return nil, err
	}

	if media.Status == "error" {
		return nil, fmt.Errorf("cobalt rejected our request: %v", media.Error.Code)
	}

	return &media, nil
}

/* End of: Download settings structs and types */

//Cobalt response end

type CobaltInstance struct {
	Trust     string   `json:"trust"`
	APIOnline bool     `json:"api_online"`
	Cors      int      `json:"cors"`
	Commit    string   `json:"commit"`
	Services  Services `json:"services,omitempty"`
	Version   string   `json:"version"`
	Branch    string   `json:"branch"`
	Score     float64  `json:"score"`
	Protocol  string   `json:"protocol"`
	Name      string   `json:"name"`
	StartTime int64    `json:"startTime"`
	API       string   `json:"api"`
	FrontEnd  string   `json:"frontEnd"`
}

type Services struct {
	Youtube       bool `json:"youtube"`
	Facebook      bool `json:"facebook"`
	Rutube        bool `json:"rutube"`
	Tumblr        bool `json:"tumblr"`
	Bilibili      bool `json:"bilibili"`
	Pinterest     bool `json:"pinterest"`
	Instagram     bool `json:"instagram"`
	Soundcloud    bool `json:"soundcloud"`
	YoutubeMusic  bool `json:"youtube_music"`
	Odnoklassniki bool `json:"odnoklassniki"`
	Dailymotion   bool `json:"dailymotion"`
	Snapchat      bool `json:"snapchat"`
	Twitter       bool `json:"twitter"`
	Loom          bool `json:"loom"`
	Vimeo         bool `json:"vimeo"`
	Streamable    bool `json:"streamable"`
	Vk            bool `json:"vk"`
	Tiktok        bool `json:"tiktok"`
	Reddit        bool `json:"reddit"`
	TwitchClips   bool `json:"twitch_clips"`
	YoutubeShorts bool `json:"youtube_shorts"`
	Vine          bool `json:"vine"`
}

// GetCobaltInstances makes a request to instances.hyper.lol and returns a list of all online cobalt instances.
func GetCobaltInstances() ([]CobaltInstance, error) {
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

	var listOfCobaltInstances []CobaltInstance
	err = json.Unmarshal(jsonbody, &listOfCobaltInstances)
	if err != nil {
		return nil, fmt.Errorf("json err? %v", err)
	}

	parseModernInstances := make([]CobaltInstance, 0)
	for _, v := range listOfCobaltInstances {
		if version.Compare(v.Version, "10.0.0", ">=") {
			parseModernInstances = append(parseModernInstances, v)
		}

	}

	return parseModernInstances, nil
}

type MediaInfo struct {
	Size uint   //Media size in bytes.
	Name string //Media name.
	Type string //Mime type of the media.
}

// ProcessMedia(url) attempts to fetch the file size, mime type and name.
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
