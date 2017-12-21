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
	"github.com/kubernetes-up-and-running/kuard/pkg/memq/server"
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
	kg    *keygen.KeyGen
	live  *debugprobe.Probe
	ready *debugprobe.Probe

	r *httprouter.Router
}

func (k *App) getPageContext(r *http.Request) *pageContext {
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

func (k *App) rootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	k.tg.Render(w, "index.html", k.getPageContext(r))
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

	router := k.r

	// Add the root handler
	router.GET("/", k.rootHandler)
	router.GET("/-/*path", k.rootHandler)

	router.Handler("GET", "/metrics", prometheus.Handler())

	// Add the static files
	sitedata.AddRoutes(router, "/built")
	sitedata.AddRoutes(router, "/static")

	router.Handler("GET", "/fs/*filepath", http.StripPrefix("/fs", http.FileServer(http.Dir("/"))))

	k.m = memory.New("/mem")
	k.m.AddRoutes(router)
	k.live = debugprobe.New("/healthy")
	k.live.AddRoutes(router)
	k.ready = debugprobe.New("/ready")
	k.ready.AddRoutes(router)
	env.New("/env").AddRoutes(router)
	dnsapi.New("/dns").AddRoutes(router)

	k.kg = keygen.New("/keygen")
	k.kg.AddRoutes(router)

	memqserver.NewServer("/memq/server").AddRoutes(router)

	return k
}
