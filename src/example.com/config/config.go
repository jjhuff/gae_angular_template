package config

import (
	"appengine"
	"sync"
)

type Config struct {
	PrerenderCacheBucket string `json:"prerender_cache_bucket"`
	PrerenderServer      string `json:"prerender_server"`
	PrerenderToken       string `json:"prerender_token"`
	GoogleAnalyticsID    string `json:"google_analytics_id"`
	Minified             bool   `json:"minified"`
	ShowRegisterRandom   bool   `json:"show_register_random"`
}

var configs = map[string]Config{
	"example-prod": Config{
		PrerenderCacheBucket: "prerender-prod.example.com",
		PrerenderServer:      "http://service.prerender.io/",
		PrerenderToken:       "XXXX",
		GoogleAnalyticsID:    "UA-XXXX-1",
		Minified:             true,
		ShowRegisterRandom:   false,
	},
	"example-qa": Config{
		PrerenderCacheBucket: "prerender-qa.example.com",
		PrerenderServer:      "http://service.prerender.io/",
		PrerenderToken:       "XXXX",
		GoogleAnalyticsID:    "UA-XXXX-2",
		Minified:             true,
		ShowRegisterRandom:   true,
	},
	"example-dev": Config{
		PrerenderCacheBucket: "prerender-qa.example.com",
		PrerenderServer:      "http://service.prerender.io/",
		PrerenderToken:       "XXXX",
		GoogleAnalyticsID:    "",
		Minified:             false,
		ShowRegisterRandom:   true,
	},
	"testapp": Config{
		PrerenderCacheBucket: "testbucket",
		PrerenderServer:      "",
		PrerenderToken:       "",
		GoogleAnalyticsID:    "",
		Minified:             false,
		ShowRegisterRandom:   true,
	},
}

var appid_once sync.Once
var appid string

func Get(ctx appengine.Context) Config {
	appid_once.Do(func() {
		appid = appengine.AppID(ctx)

	})
	return configs[appid]
}
