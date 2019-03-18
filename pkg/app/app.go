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
package app

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kubernetes-up-and-running/kuard/pkg/debugprobe"
	"github.com/kubernetes-up-and-running/kuard/pkg/dnsapi"
	"github.com/kubernetes-up-and-running/kuard/pkg/env"
	"github.com/kubernetes-up-and-running/kuard/pkg/htmlutils"
	"github.com/kubernetes-up-and-running/kuard/pkg/keygen"
	"github.com/kubernetes-up-and-running/kuard/pkg/memory"
	memqserver "github.com/kubernetes-up-and-running/kuard/pkg/memq/server"
	"github.com/kubernetes-up-and-running/kuard/pkg/sitedata"
	"github.com/kubernetes-up-and-running/kuard/pkg/version"

	"github.com/felixge/httpsnoop"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(requestDuration)
}

var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "request_duration_seconds",
	Help:    "Time serving HTTP request",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "route", "status_code"})

func promMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(h, w, r)
		requestDuration.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(m.Code)).Observe(m.Duration.Seconds())
	})
}

func loggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

type pageContext struct {
	URLBase      string       `json:"urlBase"`
	Hostname     string       `json:"hostname"`
	Addrs        []string     `json:"addrs"`
	Version      string       `json:"version"`
	VersionColor template.CSS `json:"versionColor"`
	RequestDump  string       `json:"requestDump"`
	RequestProto string       `json:"requestProto"`
	RequestAddr  string       `json:"requestAddr"`
}

type App struct {
	c  Config
	tg *htmlutils.TemplateGroup

	m     *memory.MemoryAPI
	live  *debugprobe.Probe
	ready *debugprobe.Probe
	env   *env.Env
	dns   *dnsapi.DNSAPI
	kg    *keygen.KeyGen
	mq    *memqserver.Server

	r *httprouter.Router
}

func (k *App) getPageContext(r *http.Request, urlBase string) *pageContext {
	c := &pageContext{}
	c.URLBase = urlBase
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

func (k *App) getRootHandler(urlBase string) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		k.tg.Render(w, "index.html", k.getPageContext(r, urlBase))
	})
}

// Exists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (k *App) Run() {
	r := promMiddleware(loggingMiddleware(k.r))

	// Look to see if we can find TLS certs
	certFile := filepath.Join(k.c.TLSDir, "kuard.crt")
	keyFile := filepath.Join(k.c.TLSDir, "kuard.key")
	if fileExists(certFile) && fileExists(keyFile) {
		go func() {
			log.Printf("Serving HTTPS on %v", k.c.TLSAddr)
			log.Fatal(http.ListenAndServeTLS(k.c.TLSAddr, certFile, keyFile, r))
		}()
	} else {
		log.Printf("Could not find certificates to serve TLS")
	}

	log.Printf("Serving on HTTP on %v", k.c.ServeAddr)
	log.Fatal(http.ListenAndServe(k.c.ServeAddr, r))
}

func NewApp() *App {
	k := &App{
		tg: &htmlutils.TemplateGroup{},
		r:  httprouter.New(),
	}

	// Init all of the subcomponents

	router := k.r
	k.m = memory.New()
	k.live = debugprobe.New()
	k.ready = debugprobe.New()
	k.env = env.New()
	k.dns = dnsapi.New()
	k.kg = keygen.New()
	k.mq = memqserver.NewServer()

	// Add handlers
	for _, prefix := range []string{"", "/a", "/b", "/c"} {
		rootHandler := k.getRootHandler(prefix)
		router.GET(prefix+"/", rootHandler)
		router.GET(prefix+"/-/*path", rootHandler)

		router.Handler("GET", prefix+"/metrics", prometheus.Handler())

		// Add the static files
		sitedata.AddRoutes(router, prefix+"/built")
		sitedata.AddRoutes(router, prefix+"/static")

		router.Handler("GET", prefix+"/fs/*filepath", http.StripPrefix(prefix+"/fs", http.FileServer(http.Dir("/"))))

		k.m.AddRoutes(router, prefix+"/mem")
		k.live.AddRoutes(router, prefix+"/healthy")
		k.ready.AddRoutes(router, prefix+"/ready")
		k.env.AddRoutes(router, prefix+"/env")
		k.dns.AddRoutes(router, prefix+"/dns")
		k.kg.AddRoutes(router, prefix+"/keygen")
		k.mq.AddRoutes(router, prefix+"/memq/server")
	}

	return k
}
