/*
Copyright 2017 The KUAR Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/julienschmidt/httprouter"

	"github.com/jbeda/kuard/pkg/config"
	"github.com/jbeda/kuard/pkg/debugprobe"
	"github.com/jbeda/kuard/pkg/debugsitedata"
	"github.com/jbeda/kuard/pkg/htmlutils"
	"github.com/jbeda/kuard/pkg/sitedata"
	"github.com/jbeda/kuard/pkg/version"
)

var serveAddr = flag.String("address", ":8080", "The address to serve on")

func loggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

type pageContext struct {
	Version      string
	VersionColor template.CSS
	RequestDump  string
	RequestProto string
	RequestAddr  string
	Env          map[string]string
}

type kuard struct {
	tg *htmlutils.TemplateGroup

	live  *debugprobe.Probe
	ready *debugprobe.Probe
}

func (k *kuard) getPageContext(r *http.Request) *pageContext {
	c := &pageContext{}
	c.Version = version.VERSION
	c.VersionColor = template.CSS(htmlutils.ColorFromString(version.VERSION))
	reqDump, _ := httputil.DumpRequest(r, false)
	c.RequestDump = strings.TrimSpace(string(reqDump))
	c.RequestProto = r.Proto
	c.RequestAddr = r.RemoteAddr
	c.Env = map[string]string{}
	for _, e := range os.Environ() {
		splits := strings.SplitN(e, "=", 2)
		k, v := splits[0], splits[1]
		c.Env[k] = v
	}

	return c
}

func (k *kuard) rootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	k.tg.Render(w, "index.html", k.getPageContext(r))
}

func (k *kuard) addRoutes(router *httprouter.Router) {
	// Add the root handler
	router.GET("/", k.rootHandler)

	// Add the static files
	var fs http.FileSystem
	if *config.Debug {
		fs = &assetfs.AssetFS{
			Asset:     debugsitedata.Asset,
			AssetDir:  func(path string) ([]string, error) { return nil, os.ErrNotExist },
			AssetInfo: debugsitedata.AssetInfo,
			Prefix:    "static",
		}
	} else {
		fs = &assetfs.AssetFS{
			Asset:     sitedata.Asset,
			AssetDir:  func(path string) ([]string, error) { return nil, os.ErrNotExist },
			AssetInfo: sitedata.AssetInfo,
			Prefix:    "static",
		}
	}
	router.Handler("GET", "/static/*filepath", http.StripPrefix("/static/", http.FileServer(fs)))
	router.Handler("GET", "/fs/*filepath", http.StripPrefix("/fs", http.FileServer(http.Dir("/"))))

	k.live.AddRoutes(router)
	k.ready.AddRoutes(router)
}

func NewApp() *kuard {
	k := &kuard{
		tg: &htmlutils.TemplateGroup{},
	}
	k.live = debugprobe.New("/healthy", k.tg)
	k.ready = debugprobe.New("/ready", k.tg)
	return k
}

func main() {
	flag.Parse()
	debugsitedata.SetRootDir(*config.DebugRootDir)

	log.Printf("Starting kuard version: %v", version.VERSION)
	log.Println(strings.Repeat("*", 70))
	log.Println("* WARNING: This server may expose sensitive")
	log.Println("* and secret information. Be careful.")
	log.Println(strings.Repeat("*", 70))

	app := NewApp()
	router := httprouter.New()
	app.addRoutes(router)

	log.Printf("Serving on %v", *serveAddr)
	log.Fatal(http.ListenAndServe(*serveAddr, loggingMiddleware(router)))
}
