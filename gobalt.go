package gobalt

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

var (
	CobaltPrimaryApi      = "co.wuk.sh"
	CobaltCustomInstances = []string{"cobalt-api.hyper.lol", "coapi.bigbenster702.com", "cobalt.api.timelessnesses.me", "api-dl.cgm.rs", "cobalt.synzr.space", "co-api.mae.wtf", "capi.stilic.net", "capi.oak.li", "api.c0ba.lt", "co.bloba.dev", "wukko.wolfdo.gg", "api.co.rooot.gay", "downloadapi.stuff.solutions", "us-cobalt.reed.wtf", "co-api.blueb.me"}
	useragent             = "Gobalt/1.0"
)

type settings struct {
	URL                   string
	VideoCodec            string
	VideoQuality          string
	AudioCodec            string
	FilenamePattern       string
	AudioOnly             bool
	RemoveTikTokWatermark bool
	FullTikTokAudio       bool
	RemoveAudio           bool
	DubbedYoutubeAudio    bool
	DisableVideoMetadata  bool
}

// Define Option type as a function that modifies settings
type Option func(*settings)

/* Video Codecs */
func VideoCodec(codec string) Option {
	//Expecting H264, AV1 or VP9
	supportedCodecsAudio := []string{"h264", "av1", "vp9"}
	if !slices.Contains(supportedCodecsAudio, strings.ToLower(codec)) { //Unknown codec, will use h264
		return func(s *settings) { s.VideoCodec = "h264" }
	}
	return func(s *settings) { s.VideoCodec = codec }
}

// End of Video Codecs

/* Audio Codecs */
func AudioCodec(codec string) Option {
	//Expecting MP3, Opus, ogg, wav or best
	supportedCodecsAudio := []string{"mp3", "ogg", "opus", "wav", "best"}
	if !slices.Contains(supportedCodecsAudio, strings.ToLower(codec)) { //Unknown codec, will try the best one
		return func(s *settings) { s.AudioCodec = "best" }
	}
	return func(s *settings) { s.AudioCodec = codec }
}

// End of audio codecs

func VideoQuality(quality string) Option {
	//Expecting any value between 144p and 4K, or best
	supportedQuality := []string{"144", "240", "360", "480", "720", "1080", "1440", "2160", "best"}
	if !slices.Contains(supportedQuality, quality) { //Unsupported quality, will return best
		return func(s *settings) { s.VideoQuality = "best" }
	}
	return func(s *settings) { s.VideoQuality = quality }
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
		return func(s *settings) { s.VideoQuality = "basic" }
	}
	return func(s *settings) { s.VideoQuality = pattern }
}

func AudioOnly(value bool) Option {
	return func(s *settings) { s.AudioOnly = value }
}

func RemoveTikTokWatermark(value bool) Option {
	return func(s *settings) { s.RemoveTikTokWatermark = value }
}

func FullTikTokAudio(value bool) Option {
	return func(s *settings) { s.FullTikTokAudio = value }
}

func RemoveAudio(value bool) Option {
	return func(s *settings) { s.RemoveAudio = value }
}

func DubbedYoutubeAudio(value bool) Option {
	return func(s *settings) { s.DubbedYoutubeAudio = value }
}

func DisableVideoMetadata(value bool) Option {
	return func(s *settings) { s.DisableVideoMetadata = value }
}

// New initializes a settings object with default values and applies any provided options.
func New(url string, options ...Option) *settings {
	if url == "" {
		panic("url cannot be empty")
	}
	defaults := &settings{
		URL:                   url,
		VideoCodec:            "h264",
		AudioCodec:            "mp3",
		VideoQuality:          "best",
		FilenamePattern:       "basic",
		AudioOnly:             false,
		RemoveTikTokWatermark: false,
		FullTikTokAudio:       false,
		RemoveAudio:           false,
		DubbedYoutubeAudio:    false,
		DisableVideoMetadata:  false,
	}

	for _, option := range options {
		option(defaults)
	}

	return defaults
}

func GetLink(opts *settings) (string, error) {
	validUrl, _ := regexp.MatchString(`[-a-zA-Z0-9@:%_\+.~#?&//=]{2,256}\.[a-z]{2,4}\b(\/[-a-zA-Z0-9@:%_\+.~#?&//=]*)?`, opts.URL)
	if !validUrl {
		return "", errors.New("invalid url provided.")
	}

}
