package fingerprint

import "time"

type Platform string

const (
	Windows Platform = "windows"
	MacOS   Platform = "macos"
	Linux   Platform = "linux"
	Android Platform = "android"
)

type BrowserProfile struct {
	UserAgent        string
	Platform         Platform
	SecCHUA          string
	SecCHUAMobile    string
	SecCHUAPlatform  string
	AcceptLanguage   string
	Accept           string
	SecFetchSite     string
	SecFetchMode     string
	SecFetchUser     string
	SecFetchDest     string
	Priority         string
	UpgradeInsecure  string
}

var ChromeProfiles = map[Platform]BrowserProfile{
	Windows: {
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
		Platform:        Windows,
		SecCHUA:         `"Google Chrome";v="138", "Chromium";v="138", "Not_A Brand";v="24"`,
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"Windows"`,
		AcceptLanguage:  "en-US,en;q=0.9",
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		SecFetchSite:    "none",
		SecFetchMode:    "navigate",
		SecFetchUser:    "?1",
		SecFetchDest:    "document",
		Priority:        "u=0, i",
		UpgradeInsecure: "1",
	},
	MacOS: {
		UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
		Platform:        MacOS,
		SecCHUA:         `"Google Chrome";v="138", "Chromium";v="138", "Not_A Brand";v="24"`,
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"macOS"`,
		AcceptLanguage:  "en-US,en;q=0.9",
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		SecFetchSite:    "none",
		SecFetchMode:    "navigate",
		SecFetchUser:    "?1",
		SecFetchDest:    "document",
		Priority:        "u=0, i",
		UpgradeInsecure: "1",
	},
	Linux: {
		UserAgent:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
		Platform:        Linux,
		SecCHUA:         `"Google Chrome";v="138", "Chromium";v="138", "Not_A Brand";v="24"`,
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"Linux"`,
		AcceptLanguage:  "en-US,en;q=0.9",
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		SecFetchSite:    "none",
		SecFetchMode:    "navigate",
		SecFetchUser:    "?1",
		SecFetchDest:    "document",
		Priority:        "u=0, i",
		UpgradeInsecure: "1",
	},
	Android: {
		UserAgent:       "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Mobile Safari/537.36",
		Platform:        Android,
		SecCHUA:         `"Google Chrome";v="138", "Chromium";v="138", "Not_A Brand";v="24"`,
		SecCHUAMobile:   "?1",
		SecCHUAPlatform: `"Android"`,
		AcceptLanguage:  "en-US,en;q=0.9",
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		SecFetchSite:    "none",
		SecFetchMode:    "navigate",
		SecFetchUser:    "?1",
		SecFetchDest:    "document",
		Priority:        "u=0, i",
		UpgradeInsecure: "1",
	},
}

func GetProfile(platform Platform) BrowserProfile {
	if p, ok := ChromeProfiles[platform]; ok {
		return p
	}
	return ChromeProfiles[Windows]
}

func GetRandomProfile() BrowserProfile {
	platforms := []Platform{Windows, MacOS, Linux, Android}
	return ChromeProfiles[platforms[time.Now().UnixNano()%int64(len(platforms))]]
}
