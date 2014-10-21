package config

import (
	"appengine"
	"encoding/json"
	"os"
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

var config_once sync.Once
var config *Config

func Get(ctx appengine.Context) Config {
	config_once.Do(func() {
		fileName := "config/" + appengine.AppID(ctx) + ".json"

		config = &Config{}

		f, err := os.Open(fileName)
		if err != nil {
			ctx.Warningf("Missing config file: %s", fileName)
		} else {
			ctx.Infof("Loading config: %s", fileName)
			jsonParser := json.NewDecoder(f)
			if err = jsonParser.Decode(config); err != nil {
				ctx.Errorf("Failed to parse config file: %s", err.Error())
			}
		}
	})
	return *config
}
