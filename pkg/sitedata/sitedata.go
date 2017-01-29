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
package sitedata

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

var debug bool
var debugRootDir string

func SetConfig(d bool, drd string) {
	debug = d
	debugRootDir = drd
}

func GetStaticHandler(prefix string) httprouter.Handle {
	prefix = strings.TrimPrefix(prefix, "/")
	embedFS := &assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  func(path string) ([]string, error) { return nil, os.ErrNotExist },
		AssetInfo: AssetInfo,
		Prefix:    prefix,
	}
	embedHandler := http.StripPrefix("/"+prefix+"/", http.FileServer(embedFS))

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if debug {
			fs := http.Dir(filepath.Join(debugRootDir, prefix))
			handler := http.StripPrefix("/"+prefix+"/", http.FileServer(fs))
			handler.ServeHTTP(w, r)
		} else {
			embedHandler.ServeHTTP(w, r)
		}
	}
}

func AddRoutes(r *httprouter.Router, prefix string) {
	r.GET(prefix+"/*filepath", GetStaticHandler(prefix))
}

func LoadFilesInDir(dir string) (map[string]string, error) {
	dirData := map[string]string{}
	if debug {
		fullDir := filepath.Join(debugRootDir, dir)
		files, err := ioutil.ReadDir(fullDir)
		if err != nil {
			return dirData, errors.Wrapf(err, "Error reading dir %v", debugRootDir)
		}
		for _, file := range files {
			data, err := ioutil.ReadFile(filepath.Join(fullDir, file.Name()))
			if err != nil {
				return dirData, errors.Wrapf(err, "Error loading %v", file.Name())
			}
			dirData[file.Name()] = string(data)
		}
	} else {
		files, err := AssetDir(dir)
		if err != nil {
			return dirData, errors.Wrapf(err, "Could not load bindata dir %v", dir)
		}
		for _, file := range files {
			fullName := path.Join("templates", file)
			data, err := Asset(fullName)
			if err != nil {
				return dirData, errors.Wrapf(err, "Error loading bindata %v", fullName)
			}
			dirData[file] = string(data)
		}
	}
	return dirData, nil
}
