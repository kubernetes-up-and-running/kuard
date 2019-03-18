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

package env

import (
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kubernetes-up-and-running/kuard/pkg/apiutils"
)

// EnvStatus is returned from a GET to this API endpoing
type EnvStatus struct {
	CommandLine []string          `json:"commandLine"`
	Env         map[string]string `json:"env"`
}

type Env struct {
}

func New() *Env {
	return &Env{}
}

func (e *Env) AddRoutes(r *httprouter.Router, base string) {
	r.GET(base+"/api", e.APIGet)
}

func (e *Env) APIGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := EnvStatus{}

	s.CommandLine = os.Args

	s.Env = map[string]string{}
	for _, e := range os.Environ() {
		splits := strings.SplitN(e, "=", 2)
		k, v := splits[0], splits[1]
		s.Env[k] = v
	}

	apiutils.ServeJSON(w, s)
}
