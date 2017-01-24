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
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/julienschmidt/httprouter"

	"github.com/jbeda/kuard/pkg/config"
	"github.com/jbeda/kuard/pkg/debugprobe"
	"github.com/jbeda/kuard/pkg/debugsitedata"
	"github.com/jbeda/kuard/pkg/dnsapi"
	"github.com/jbeda/kuard/pkg/env"
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
	Hostname     string       `json:"hostname"`
	Addrs        []string     `json:"addrs"`
	Version      string       `json:"version"`
	VersionColor template.CSS `json:"versionColor"`
	RequestDump  string       `json:"requestDump"`
	RequestProto string       `json:"requestProto"`
	RequestAddr  string       `json:"requestAddr"`
}

type kuard struct {
	tg *htmlutils.TemplateGroup
}

func (k *kuard) getPageContext(r *http.Request) *pageContext {
	c := &pageContext{}
	c.Hostname, _ = os.Hostname()

	addrs, _ := net.InterfaceAddrs()
	c.Addrs = []string{}
	for _, addr := range addrs {
		// check the address type and if it is not a loopback
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				c.Addrs = append(c.Addrs, ipnet.IP.String())
			}
		}
	}

	c.Version = version.VERSION
	c.VersionColor = template.CSS(htmlutils.ColorFromString(version.VERSION))
	reqDump, _ := httputil.DumpRequest(r, false)
	c.RequestDump = strings.TrimSpace(string(reqDump))
	c.RequestProto = r.Proto
	c.RequestAddr = r.RemoteAddr

	return c
}

func (k *kuard) rootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	k.tg.Render(w, "index.html", k.getPageContext(r))
}

func fsHandlerForPrefix(prefix string) http.Handler {
	var fs http.FileSystem
	if *config.Debug {
		fs = &assetfs.AssetFS{
			Asset:     debugsitedata.Asset,
			AssetDir:  func(path string) ([]string, error) { return nil, os.ErrNotExist },
			AssetInfo: debugsitedata.AssetInfo,
			Prefix:    prefix,
		}
	} else {
		fs = &assetfs.AssetFS{
			Asset:     sitedata.Asset,
			AssetDir:  func(path string) ([]string, error) { return nil, os.ErrNotExist },
			AssetInfo: sitedata.AssetInfo,
			Prefix:    prefix,
		}
	}
	return http.StripPrefix("/"+prefix+"/", http.FileServer(fs))
}

func NewApp(router *httprouter.Router) *kuard {
	k := &kuard{
		tg: &htmlutils.TemplateGroup{},
	}

	// Add the root handler
	router.GET("/", k.rootHandler)

	// Add the static files
	router.Handler("GET", "/built/*filepath", fsHandlerForPrefix("built"))
	router.Handler("GET", "/static/*filepath", fsHandlerForPrefix("static"))

	router.Handler("GET", "/fs/*filepath", http.StripPrefix("/fs", http.FileServer(http.Dir("/"))))

	debugprobe.New("/healthy").AddRoutes(router)
	debugprobe.New("/ready").AddRoutes(router)
	env.New("/env").AddRoutes(router)
	dnsapi.New("/dns").AddRoutes(router)

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

	router := httprouter.New()
	NewApp(router)

	log.Printf("Serving on %v", *serveAddr)
	log.Fatal(http.ListenAndServe(*serveAddr, loggingMiddleware(router)))
}
