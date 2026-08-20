package fingerprint

import (
	"fmt"
	"net/http"
)

var HeaderOrder = []string{
	"sec-ch-ua",
	"sec-ch-ua-mobile",
	"sec-ch-ua-platform",
	"upgrade-insecure-requests",
	"user-agent",
	"accept",
	"sec-fetch-site",
	"sec-fetch-mode",
	"sec-fetch-user",
	"sec-fetch-dest",
	"accept-encoding",
	"accept-language",
	priority",
}

var PseudoHeaderOrder = []string{
	":method",
	":authority",
	":scheme",
	":path",
}

func BuildHeaders(profile BrowserProfile) http.Header {
	h := http.Header{}
	h.Set("sec-ch-ua", profile.SecCHUA)
	h.Set("sec-ch-ua-mobile", profile.SecCHUAMobile)
	h.Set("sec-ch-ua-platform", profile.SecCHUAPlatform)
	h.Set("upgrade-insecure-requests", profile.UpgradeInsecure)
	h.Set("user-agent", profile.UserAgent)
	h.Set("accept", profile.Accept)
	h.Set("sec-fetch-site", profile.SecFetchSite)
	h.Set("sec-fetch-mode", profile.SecFetchMode)
	h.Set("sec-fetch-user", profile.SecFetchUser)
	h.Set("sec-fetch-dest", profile.SecFetchDest)
	h.Set("accept-encoding", "gzip, deflate, br, zstd")
	h.Set("accept-language", profile.AcceptLanguage)
	h.Set("priority", profile.Priority)
	return h
}

func BuildAPIHeaders(profile BrowserProfile, origin string) http.Header {
	h := http.Header{}
	h.Set("accept", "*/*")
	h.Set("accept-language", profile.AcceptLanguage)
	h.Set("content-type", "text/plain;charset=UTF-8")
	h.Set("origin", origin)
	h.Set("sec-ch-ua", profile.SecCHUA)
	h.Set("sec-ch-ua-mobile", profile.SecCHUAMobile)
	h.Set("sec-ch-ua-platform", profile.SecCHUAPlatform)
	h.Set("sec-fetch-dest", "empty")
	h.Set("sec-fetch-mode", "cors")
	h.Set("sec-fetch-site", "cross-site")
	h.Set("user-agent", profile.UserAgent)
	h.Set("accept-encoding", "gzip, deflate, br, zstd")
	return h
}

// BuildChallengeHeaders builds headers for making requests after solving a challenge.
// These mimic a real browser that just passed the Cloudflare challenge.
func BuildChallengeHeaders(profile BrowserProfile, referer string) http.Header {
	h := http.Header{}
	h.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	h.Set("accept-encoding", "gzip, deflate, br, zstd")
	h.Set("accept-language", profile.AcceptLanguage)
	h.Set("sec-ch-ua", profile.SecCHUA)
	h.Set("sec-ch-ua-mobile", profile.SecCHUAMobile)
	h.Set("sec-ch-ua-platform", profile.SecCHUAPlatform)
	h.Set("sec-fetch-dest", "document")
	h.Set("sec-fetch-mode", "navigate")
	h.Set("sec-fetch-site", "same-origin")
	h.Set("sec-fetch-user", "?1")
	h.Set("upgrade-insecure-requests", "1")
	h.Set("user-agent", profile.UserAgent)
	if referer != "" {
		h.Set("referer", referer)
	}
	h.Set("priority", profile.Priority)
	return h
}

func HeaderOrderString() string {
	s := ""
	for i, name := range HeaderOrder {
		if i > 0 {
			s += ","
		}
		s += name
	}
	return s
}

func FormatAkamaiFingerprint(settings H2Settings) string {
	return fmt.Sprintf("1:%d;2:%d;4:%d;6:%d|%d|0|%s",
		settings.HeaderTableSize,
		settings.EnablePush,
		settings.InitialWindowSize,
		settings.MaxHeaderListSize,
		settings.WindowUpdateSize,
		"m,a,s,p",
	)
}
