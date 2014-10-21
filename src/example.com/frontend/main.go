package example

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"example.com/auth"
	"example.com/config"
	"example.com/user"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/goauth2/appengine/serviceaccount"
	storage "code.google.com/p/google-api-go-client/storage/v1"
	"github.com/ant0ine/go-json-rest/rest"
)

type AppContext struct {
}

var fileHashes = make(map[string]string)

var templateFuncs = template.FuncMap{
	"static": func(name string) string {
		if v, ok := fileHashes[name]; ok {
			return v
		} else {
			return name
		}
	},
}

var indexTmpl = template.Must(template.New("index.html").Funcs(templateFuncs).ParseFiles("html/index.html"))

type indexArgs struct {
	User   *user.User
	Config config.Config
}

func (app *AppContext) handleIndex(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	data := indexArgs{
		Config: config.Get(ctx),
	}

	// See if we have a user and pass it to the template
	userId, err := auth.VerifyCookie(ctx, req)
	if err == nil {
		userDb := user.NewAppEngineUserDB(ctx)
		u, err := userDb.GetById(userId)
		if err == nil {
			data.User = u
		}
	}
	w.Header().Set("Cache-Control", "private, max-age=0, no-cache")

	if err := indexTmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

var botsRegexp = regexp.MustCompile("(?i)googlebot|yahoo|bing|baidu|jeeves|facebook|twitter|linkedin|archive.org")

func shouldPreRender(req *http.Request) bool {
	var ua = req.Header.Get("User-Agent")

	if strings.Contains(ua, "Prerender") {
		return false
	}

	if _, ok := req.URL.Query()["_force_prerender"]; ok == true {
		return true
	}

	/*	if _, ok := req.URL.Query()["_escaped_fragment_"]; ok == true {
		return true
	}*/

	if botsRegexp.MatchString(ua) {
		return true
	}

	return false
}

func getStorageService(ctx appengine.Context) (*storage.Service, error) {
	client, err := serviceaccount.NewClient(ctx, storage.DevstorageRead_onlyScope)
	if err != nil {
		return nil, err
	}
	service, err := storage.New(client)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func allowPreRender(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if shouldPreRender(req) {
			u := *req.URL
			ctx := appengine.NewContext(req)

			oauthClient, err := serviceaccount.NewClient(ctx, storage.DevstorageRead_onlyScope)
			if err != nil {
				ctx.Errorf("Error oauth client: %s", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Try to fetch the cached version from GoogleStorage
			cacheUrl := "https://storage.googleapis.com/" + config.Get(ctx).PrerenderCacheBucket + u.Path
			resp, err := oauthClient.Get(cacheUrl)
			if err == nil && resp.StatusCode == http.StatusOK {
				ctx.Infof("Found in prerender cache: %s", u.Path)
			} else {
				// Othereise, do it dynamically
				client := &http.Client{
					Transport: &urlfetch.Transport{
						Context:  ctx,
						Deadline: 60 * time.Second,
					},
				}
				ctx.Infof("Dynamic prerendering: %s", u.String())
				proxyReq, err := http.NewRequest("GET", config.Get(ctx).PrerenderServer+u.String(), nil)
				if err != nil {
					ctx.Errorf("Error requesting pre-render: %s", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				proxyReq.Header.Add("X-Prerender-Token", config.Get(ctx).PrerenderToken) // in case we're using prerender.io
				proxyReq.Header.Add("User-Agent", req.Header.Get("User-Agent"))
				resp, err = client.Do(proxyReq)
				if err != nil {
					ctx.Errorf("Error reading pre-render: %s", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			defer resp.Body.Close()
			io.Copy(w, resp.Body)
		} else {
			handler(w, req)
		}
	}
}

func loadFileHashes(fname string) {
	file, err := ioutil.ReadFile(fname)
	if err != nil {
		panic("Failed to load: " + fname)
	}

	var hashes map[string]string
	err = json.Unmarshal(file, &hashes)
	if err != nil {
		panic("Failed to parse: " + fname)
	}

	for k, v := range hashes {
		fileHashes[k] = v
	}
}

func init() {
	app := &AppContext{}

	loadFileHashes("build/appjs-manifest.json")
	loadFileHashes("build/libjs-manifest.json")
	loadFileHashes("build/css-manifest.json")

	restHandler := rest.ResourceHandler{}
	restHandler.SetRoutes(
		&rest.Route{HttpMethod: "POST", PathExp: "/register", Func: app.RegisterUser},
		&rest.Route{HttpMethod: "POST", PathExp: "/login", Func: app.LoginUser},
		&rest.Route{HttpMethod: "POST", PathExp: "/logout", Func: app.LogoutUser},
		&rest.Route{HttpMethod: "POST", PathExp: "/change_password", Func: app.ChangePassword},
		&rest.Route{HttpMethod: "POST", PathExp: "/reset_password", Func: app.ResetPassword},
		&rest.Route{HttpMethod: "POST", PathExp: "/forgot_password", Func: app.SendForgotPassword},

		&rest.Route{HttpMethod: "GET", PathExp: "/user/:id", Func: app.GetUser},
		&rest.Route{HttpMethod: "PUT", PathExp: "/user/:id", Func: app.PutUser},
	)

	http.Handle("/_/api/v1/", http.StripPrefix("/_/api/v1", &restHandler))
	http.HandleFunc("/", allowPreRender(app.handleIndex))
}
