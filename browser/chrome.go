package browser

import (
	"github.com/Noooste/azuretls-client"
)

// TODO: update default browser headers
var (
	CHROME = azuretls.OrderedHeaders{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		{"accept-encoding", "gzip, deflate, br"},
		{"accept-language", "zh-HK,zh-TW;q=0.9,zh;q=0.8"},
		{"cache-control", "max-age=0"},
		{"sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"Windows"`},
		{"sec-fetch-site", "none"},
		{"sec-fetch-mode", "navigate"},
		{"sec-fetch-user", "?1"},
		{"sec-fetch-dest", "document"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		{"upgrade-insecure-requests", "1"},
	}
)
