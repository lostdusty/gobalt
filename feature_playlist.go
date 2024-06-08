package gobalt

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.nobrain.org/r4/dischord/extractor"
	exutil "git.nobrain.org/r4/dischord/extractor/util"
	"git.nobrain.org/r4/dischord/util"
)

//Code provided nicely by xypwn.
//Original Source: https://github.com/xypwn/dischord/tree/main/extractor/youtube

var (
	ErrNoSuitableFormat              = errors.New("no suitable audio-only format found")
	ErrGettingUrlFromSignatureCipher = errors.New("error getting URL from signature cipher")
	ErrDecryptFunctionBroken         = errors.New("signature decryptor function is broken (perhaps the extractor is out of date)")
	ErrMalformedJson                 = errors.New("malformed JSON")
	ErrDecryptGettingFunctionName    = errors.New("error getting signature decryption function name")
	ErrDecryptGettingFunction        = errors.New("error getting signature decryption function")
	ErrDecryptGettingOpTable         = errors.New("error getting signature decryption operation table")
	ErrGettingBaseJs                 = errors.New("unable to get base.js")
)

type decryptorOp struct {
	fn  func(a *string, b int)
	arg int
}

type decryptor struct {
	// base.js version ID, used for caching
	versionId string
	// The actual decryption algorithm can be split up into a list of known
	// operations
	ops []decryptorOp
}

func (d *decryptor) decrypt(input string) (string, error) {
	if err := updateDecryptor(d); err != nil {
		return "", err
	}

	s := input
	for _, op := range d.ops {
		op.fn(&s, op.arg)
	}
	return s, nil
}

type configData struct {
	PlayerJsUrl string `json:"PLAYER_JS_URL"`
}

func updateDecryptor(d *decryptor) error {
	prefix := "(function() {window.ytplayer={};\nytcfg.set("
	endStr := ");"
	// Get base.js URL
	var url string
	var funcErr error
	err := exutil.GetHTMLScriptFunc("https://www.youtube.com", false, func(code string) bool {
		if strings.HasPrefix(code, prefix) {
			// Cut out the JSON part
			code = code[len(prefix):]
			end := strings.Index(code, endStr)
			if end == -1 {
				funcErr = ErrGettingBaseJs
				return false
			}

			// Parse config data
			var data configData
			if err := json.Unmarshal([]byte(code[:end]), &data); err != nil {
				funcErr = ErrGettingBaseJs
				return false
			}

			url = "https://www.youtube.com" + data.PlayerJsUrl
			return false
		}
		return true
	})
	if err != nil {
		return err
	}
	if funcErr != nil {
		return err
	}

	// Get base.js version ID
	sp := strings.SplitN(strings.TrimPrefix(url, "/s/player/"), "/", 2)
	if len(sp) != 2 {
		return ErrGettingBaseJs
	}
	verId := sp[0]

	if d.versionId == verId {
		// Decryptor already up-to-date
		return nil
	}

	// Get base.js contents
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ErrGettingBaseJs
	}

	// Copy contents to buffer
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return err
	}

	// Get decryption operations
	ops, err := getDecryptOps(buf.String())
	if err != nil {
		return err
	}

	d.versionId = verId
	d.ops = ops
	return nil
}

var decryptFunctionNameRegexp = regexp.MustCompile(`[a-zA-Z]*&&\([a-zA-Z]*=([a-zA-Z]*)\(decodeURIComponent\([a-zA-Z]*\)\),[a-zA-Z]*\.set\([a-zA-Z]*,encodeURIComponent\([a-zA-Z]*\)\)\)`)

func getDecryptFunction(baseJs string) (string, error) {
	idx := decryptFunctionNameRegexp.FindSubmatchIndex([]byte(baseJs))
	if len(idx) != 4 {
		return "", ErrDecryptGettingFunctionName
	}
	fnName := baseJs[idx[2]:idx[3]]

	startMatch := fnName + `=function(a){a=a.split("");`
	endMatch := `;return a.join("")};`
	start := strings.Index(baseJs, startMatch)
	if start == -1 {
		return "", ErrDecryptGettingFunction
	}
	fn := baseJs[start+len(startMatch):]
	end := strings.Index(fn, endMatch)
	if end == -1 {
		return "", ErrDecryptGettingFunction
	}
	return fn[:end], nil
}

func getDecryptOps(baseJs string) ([]decryptorOp, error) {
	// Extract main decryptor function JS
	decrFn, err := getDecryptFunction(baseJs)
	if err != nil {
		return nil, err
	}

	// Get decyptor operation JS
	var ops string
	{
		sp := strings.SplitN(decrFn, ".", 2)
		if len(sp) != 2 {
			return nil, ErrDecryptGettingOpTable
		}
		opsObjName := sp[0]

		startMatch := `var ` + opsObjName + `={`
		endMatch := `};`
		start := strings.Index(baseJs, startMatch)
		if start == -1 {
			return nil, ErrDecryptGettingOpTable
		}
		ops = baseJs[start+len(startMatch):]
		end := strings.Index(ops, endMatch)
		if end == -1 {
			return nil, ErrDecryptGettingOpTable
		}
		ops = ops[:end]
	}

	// Make a decryptor operation table that associates the operation
	// names with a specific action on an input string
	opTable := make(map[string]func(a *string, b int))
	{
		lns := strings.Split(ops, "\n")
		if len(lns) != 3 {
			return nil, ErrDecryptGettingOpTable
		}
		for _, ln := range lns {
			sp := strings.Split(ln, ":")
			if len(sp) != 2 {
				return nil, ErrDecryptGettingOpTable
			}
			name := sp[0]
			fn := sp[1]
			switch {
			case strings.HasPrefix(fn, `function(a){a.reverse()}`):
				opTable[name] = func(a *string, b int) {
					// Reverse a
					var res string
					for _, c := range *a {
						res = string(c) + res
					}
					*a = res
				}
			case strings.HasPrefix(fn, `function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}`):
				opTable[name] = func(a *string, b int) {
					// Swap a[0] and a[b % len(a)]
					c := []byte(*a)
					c[0], c[b%len(*a)] = c[b%len(*a)], c[0]
					*a = string(c)
				}
			case strings.HasPrefix(fn, `function(a,b){a.splice(0,b)}`):
				opTable[name] = func(a *string, b int) {
					// Slice off all elements of a up to a[b]
					*a = (*a)[b:]
				}
			}
		}
	}

	// Parse all operations in the main decryptor function and return them in
	// order
	var res []decryptorOp
	for _, fn := range strings.Split(decrFn, ";") {
		sp := strings.SplitN(fn, ".", 2)
		if len(sp) != 2 {
			return nil, ErrDecryptGettingOpTable
		}
		sp = strings.SplitN(sp[1], "(", 2)
		if len(sp) != 2 {
			return nil, ErrDecryptGettingOpTable
		}
		name := sp[0]
		argS := strings.TrimSuffix(strings.TrimPrefix(sp[1], "a,"), ")")
		arg, err := strconv.Atoi(argS)
		if err != nil {
			return nil, ErrDecryptGettingOpTable
		}
		callableOp, exists := opTable[name]
		if !exists {
			return nil, ErrDecryptGettingOpTable
		}
		res = append(res, decryptorOp{callableOp, arg})
	}
	return res, nil
}

type playlistVideoData struct {
	Contents struct {
		TwoColumnWatchNextResults struct {
			Playlist struct {
				Playlist struct {
					Title    string `json:"title"`
					Contents []struct {
						PlaylistPanelVideoRenderer struct {
							NavigationEndpoint struct {
								WatchEndpoint struct {
									VideoId string `json:"videoId"`
									Index   int    `json:"index"`
								} `json:"watchEndpoint"`
							} `json:"navigationEndpoint"`
							Title struct {
								SimpleText string `json:"simpleText"`
							} `json:"title"`
							ShortBylineText struct {
								Runs []struct {
									Text string `json:"text"` // uploader name
								} `json:"runs"`
							} `json:"shortBylineText"`
							LengthText struct {
								SimpleText string `json:"simpleText"`
							} `json:"lengthText"`
						} `json:"playlistPanelVideoRenderer"`
					} `json:"contents"`
				} `json:"playlist"`
			} `json:"playlist"`
		} `json:"twoColumnWatchNextResults"`
	} `json:"contents"`
}

type YoutubePlaylist struct {
	VideoURL      string
	VideoTitle    string
	VideoUploader string
}

// Only gets superficial data, the actual stream URL must be extracted from SourceUrl
func GetPlaylist(pUrl string) ([]YoutubePlaylist, error) {
	u, err := url.Parse(pUrl)
	if err != nil {
		return nil, err
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}
	listId := q.Get("list")
	vidId := ""
	index := 0

	var res []extractor.Data

	// This loop uses the playlist sidebar: each video played in the context
	// of a playlist loads 100 or so of the following videos' infos, which we
	// add to the returned slice; then we take the last retrieved video's infos
	// and use its sidebar and so on
	for {
		vUrl := "https://www.youtube.com/watch?v=" + vidId + "&list=" + listId + "&index=" + strconv.Itoa(index+1)

		// Get JSON string from YouTube
		v, err := getJSVar(vUrl, "ytInitialData")
		if err != nil {
			return nil, err
		}

		// Parse playlist data scraped from YouTube
		var data playlistVideoData
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, err
		}

		added := false
		for _, v := range data.Contents.TwoColumnWatchNextResults.Playlist.Playlist.Contents {
			vidId = v.PlaylistPanelVideoRenderer.NavigationEndpoint.WatchEndpoint.VideoId
			index = v.PlaylistPanelVideoRenderer.NavigationEndpoint.WatchEndpoint.Index

			if index == len(res) {
				srcUrl := "https://www.youtube.com/watch?v=" + vidId

				bylineText := v.PlaylistPanelVideoRenderer.ShortBylineText
				if len(bylineText.Runs) == 0 {
					return nil, ErrMalformedJson
				}
				uploader := bylineText.Runs[0].Text

				length, err := util.ParseDurationSeconds(v.PlaylistPanelVideoRenderer.LengthText.SimpleText)
				if err != nil {
					length = -1
				}

				res = append(res, extractor.Data{
					SourceUrl:     srcUrl,
					Title:         v.PlaylistPanelVideoRenderer.Title.SimpleText,
					PlaylistUrl:   "https://www.youtube.com/playlist?list=" + listId,
					PlaylistTitle: data.Contents.TwoColumnWatchNextResults.Playlist.Playlist.Title,
					Uploader:      uploader,
					Duration:      length,
				})

				added = true
			}
		}

		if !added {
			break
		}
	}

	cleanRes := make([]YoutubePlaylist, 0)
	for _, v := range res {
		cleanRes = append(cleanRes, YoutubePlaylist{
			VideoURL:      v.SourceUrl,
			VideoTitle:    v.Title,
			VideoUploader: v.Uploader,
		})
	}

	return cleanRes, nil
}

// Gets a constant JavaScript variable's value from a URL and a variable name
// (variable format must be: var someVarName = {"somekey": "lol"};)
func getJSVar(url, varName string) (string, error) {
	match := "var " + varName + " = "

	var res string
	err := exutil.GetHTMLScriptFunc(url, true, func(code string) bool {
		if strings.HasPrefix(code, match) {
			res = strings.TrimRight(code[len(match):], ";")
			return false
		}
		return true
	})
	if err != nil {
		return "", err
	}
	return res, nil
}

type playerData struct {
	StreamingData struct {
		ExpiresInSeconds string `json:"expiresInSeconds"`
		Formats          []struct {
			Url              string `json:"url"`
			SignatureCipher  string `json:"signatureCipher"`
			MimeType         string `json:"mimeType"`
			Bitrate          int    `json:"bitrate"`
			ApproxDurationMs string `json:"approxDurationMs"`
			AudioSampleRate  string `json:"audioSampleRate"`
			AudioChannels    int    `json:"audioChannels"`
		} `json:"formats"`
		AdaptiveFormats []struct {
			Url              string `json:"url"`
			SignatureCipher  string `json:"signatureCipher"`
			MimeType         string `json:"mimeType"`
			Bitrate          int    `json:"bitrate"`
			ApproxDurationMs string `json:"approxDurationMs"`
			AudioSampleRate  string `json:"audioSampleRate"`
			AudioChannels    int    `json:"audioChannels"`
		} `json:"adaptiveFormats"`
	} `json:"streamingData"`
	VideoDetails struct {
		VideoId          string `json:"videoId"`
		Title            string `json:"title"`
		LengthSeconds    string `json:"lengthSeconds"`
		ShortDescription string `json:"shortDescription"`
		Author           string `json:"author"`
	} `json:"videoDetails"`
}

func getVideo(decryptor *decryptor, vUrl string) (extractor.Data, error) {
	try := func() (extractor.Data, error) {
		// Get JSON string from YouTube
		v, err := getJSVar(vUrl, "ytInitialPlayerResponse")
		if err != nil {
			return extractor.Data{}, err
		}

		// Parse player data scraped from YouTube
		var data playerData
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return extractor.Data{}, err
		}

		// Get audio format with maximum bitrate
		maxBr := -1
		for i, f := range data.StreamingData.AdaptiveFormats {
			if strings.HasPrefix(f.MimeType, "audio/") {
				if maxBr == -1 || f.Bitrate > data.StreamingData.AdaptiveFormats[maxBr].Bitrate {
					maxBr = i
				}
			}
		}
		if maxBr == -1 {
			return extractor.Data{}, ErrNoSuitableFormat
		}

		duration, err := strconv.Atoi(data.VideoDetails.LengthSeconds)
		if err != nil {
			duration = -1
		}
		expires, err := strconv.Atoi(data.StreamingData.ExpiresInSeconds)
		if err != nil {
			return extractor.Data{}, err
		}

		ft := data.StreamingData.AdaptiveFormats[maxBr]
		var resUrl string
		if ft.Url != "" {
			resUrl = ft.Url
		} else {
			// For music, YouTube makes getting the resource URL a bit trickier
			q, err := url.ParseQuery(ft.SignatureCipher)
			if err != nil {
				return extractor.Data{}, ErrGettingUrlFromSignatureCipher
			}
			sig := q.Get("s")
			sigParam := q.Get("sp")
			baseUrl := q.Get("url")
			sigDecrypted, err := decryptor.decrypt(sig)
			if err != nil {
				return extractor.Data{}, err
			}
			resUrl = baseUrl + "&" + sigParam + "=" + sigDecrypted
		}

		return extractor.Data{
			SourceUrl:   vUrl,
			StreamUrl:   resUrl,
			Title:       data.VideoDetails.Title,
			Description: data.VideoDetails.ShortDescription,
			Uploader:    data.VideoDetails.Author,
			Duration:    duration,
			Expires:     time.Now().Add(time.Duration(expires) * time.Second),
		}, nil
	}

	isOk := func(strmUrl string) bool {
		resp, err := http.Get(strmUrl)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}

	// Sometimes we just get an invalid stream URL, and I didn't find anything
	// simple to do about it, so we just try the stream URL we get and repeat
	// if it's invalid
	for tries := 0; tries < 10; tries++ {
		data, err := try()
		if err != nil {
			return extractor.Data{}, err
		}
		if isOk(data.StreamUrl) {
			return data, nil
		}
	}

	return extractor.Data{}, ErrDecryptFunctionBroken
}

type searchData struct {
	Contents struct {
		TwoColumnSearchResultsRenderer struct {
			PrimaryContents struct {
				SectionListRenderer struct {
					Contents []struct {
						ItemSectionRenderer struct {
							Contents []struct {
								PlaylistRenderer struct {
									PlaylistId string `json:"playlistId"`
									Title      struct {
										SimpleText string `json:"simpleText"`
									} `json:"title"`
								} `json:"playlistRenderer"`
								VideoRenderer struct {
									VideoId string `json:"videoId"`
									Title   struct {
										Runs []struct {
											Text string `json:"text"`
										} `json:"runs"`
									} `json:"title"`
									LongBylineText struct {
										Runs []struct {
											Text string `json:"text"` // uploader name
										} `json:"runs"`
									} `json:"longBylineText"`
									LengthText struct {
										SimpleText string `json:"simpleText"`
									} `json:"lengthText"`
									OwnerBadges []struct {
										MetadataBadgeRenderer struct {
											Style string `json:"style"`
										} `json:"metadataBadgeRenderer"`
									} `json:"OwnerBadges"`
								} `json:"videoRenderer"`
							} `json:"contents"`
						} `json:"itemSectionRenderer"`
					} `json:"contents"`
				} `json:"sectionListRenderer"`
			} `json:"primaryContents"`
		} `json:"twoColumnSearchResultsRenderer"`
	} `json:"contents"`
}

// Only gets superficial data, the actual stream URL must be extracted from SourceUrl
func getSearch(query string) ([]extractor.Data, error) {
	// Get JSON string from YouTube
	sanitizedQuery := url.QueryEscape(strings.ReplaceAll(query, " ", "+"))
	queryUrl := "https://www.youtube.com/results?search_query=" + sanitizedQuery
	v, err := getJSVar(queryUrl, "ytInitialData")
	if err != nil {
		return nil, err
	}

	// Parse search data scraped from YouTube
	var data searchData
	if err := json.Unmarshal([]byte(v), &data); err != nil {
		return nil, err
	}

	var res []extractor.Data
	for _, v0 := range data.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents {
		for _, v1 := range v0.ItemSectionRenderer.Contents {
			if v1.VideoRenderer.VideoId != "" {
				titleRuns := v1.VideoRenderer.Title.Runs
				if len(titleRuns) == 0 {
					return nil, ErrMalformedJson
				}
				title := titleRuns[0].Text

				bylineText := v1.VideoRenderer.LongBylineText
				if len(bylineText.Runs) == 0 {
					return nil, ErrMalformedJson
				}
				uploader := bylineText.Runs[0].Text

				length, err := util.ParseDurationSeconds(v1.VideoRenderer.LengthText.SimpleText)
				if err != nil {
					length = -1
				}

				badges := v1.VideoRenderer.OwnerBadges

				res = append(res, extractor.Data{
					SourceUrl:      "https://www.youtube.com/watch?v=" + v1.VideoRenderer.VideoId,
					Title:          title,
					Duration:       length,
					Uploader:       uploader,
					OfficialArtist: len(badges) != 0 && badges[0].MetadataBadgeRenderer.Style == "BADGE_STYLE_TYPE_VERIFIED_ARTIST",
				})
			} else if v1.PlaylistRenderer.PlaylistId != "" {
				res = append(res, extractor.Data{
					PlaylistUrl:   "https://www.youtube.com/playlist?list=" + v1.PlaylistRenderer.PlaylistId,
					PlaylistTitle: v1.PlaylistRenderer.Title.SimpleText,
				})
			}
		}
	}

	return res, nil
}

func getSearchSuggestions(query string) ([]string, error) {
	url := "https://suggestqueries-clients6.youtube.com/complete/search?client=youtube&ds=yt&q=" + url.QueryEscape(query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	raw = []byte(strings.TrimSuffix(strings.TrimPrefix(string(raw), "window.google.ac.h("), ")"))

	var data []any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	if len(data) != 3 {
		return nil, ErrMalformedJson
	}
	rawSuggestions, ok := data[1].([]any)
	if !ok {
		return nil, ErrMalformedJson
	}

	var res []string
	for _, v := range rawSuggestions {
		rawSuggestion, ok := v.([]any)
		if !ok || len(rawSuggestion) != 3 {
			return nil, ErrMalformedJson
		}
		suggestion, ok := rawSuggestion[0].(string)
		if !ok {
			return nil, ErrMalformedJson
		}
		res = append(res, suggestion)
	}
	return res, nil
}
