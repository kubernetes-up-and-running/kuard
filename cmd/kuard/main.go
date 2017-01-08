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
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/elazarl/go-bindata-assetfs"

	"github.com/jbeda/kuard/pkg/sitedata"
	"github.com/jbeda/kuard/pkg/version"
)

const serveAddr = ":8080"

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	data := sitedata.MustAsset("templates/index.html")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	fmt.Fprint(w, string(data))
}

func httpLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	log.Printf("Starting kuard version: %v", version.VERSION)

	http.Handle("/", httpLog(http.HandlerFunc(rootHandler)))
	http.Handle("/static/",
		httpLog(
			http.StripPrefix("/static/",
				http.FileServer(
					&assetfs.AssetFS{
						Asset:     sitedata.Asset,
						AssetDir:  func(path string) ([]string, error) { return nil, os.ErrNotExist },
						AssetInfo: sitedata.AssetInfo,
						Prefix:    "static",
					}))))

	log.Printf("Serving on %v", serveAddr)
	log.Fatal(http.ListenAndServe(serveAddr, nil))
}
