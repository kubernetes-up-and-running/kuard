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

package keygen

import (
	"encoding/json"
	"net/http"

	"github.com/kubernetes-up-and-running/kuard/pkg/apiutils"
	"github.com/julienschmidt/httprouter"
)

// ProbeStatus is returned from a GET to this API endpoing
type KeyGenStatus struct {
	Config  Config    `json:"config"`
	History []History `json:"history"`
}

type History struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

func (kg *KeyGen) APIPut(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	c := Config{}

	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	kg.LoadConfig(c)

	kg.APIGet(w, r, params)
}

func (kg *KeyGen) APIGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	s := &KeyGenStatus{
		Config:  kg.config,
		History: kg.history,
	}

	apiutils.ServeJSON(w, s)
}
