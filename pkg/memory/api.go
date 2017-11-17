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

package memory

import (
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/kubernetes-up-and-running/kuard/pkg/apiutils"
)

type MemoryAPI struct {
	basePath string
	leaks    [][]byte
}

// MemoryStatus is returned from a GET to this API endpoing
type MemoryStatus struct {
	MemStats runtime.MemStats `json:"memStats"`
}

func New(base string) *MemoryAPI {
	return &MemoryAPI{
		basePath: base,
	}
}

func (e *MemoryAPI) AddRoutes(r *httprouter.Router) {
	r.GET(e.basePath+"/api", e.APIGet)
	r.POST(e.basePath+"/api/alloc", e.APIAlloc)
	r.POST(e.basePath+"/api/clear", e.APIClear)
}

func (e *MemoryAPI) APIGet(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	resp := &MemoryStatus{}

	runtime.ReadMemStats(&resp.MemStats)

	apiutils.ServeJSON(w, resp)
}

func (m *MemoryAPI) APIAlloc(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sSize := r.URL.Query().Get("size")
	if len(sSize) == 0 {
		http.Error(w, "size not specified", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(sSize, 10, 64)
	if err != nil {
		http.Error(w, "bad size param", http.StatusBadRequest)
	}

	leak := make([]byte, i, i)
	for i := 0; i < len(leak); i++ {
		leak[i] = 'x'
	}

	m.leaks = append(m.leaks, leak)
}

func (m *MemoryAPI) APIClear(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	m.leaks = nil
	runtime.GC()
	debug.FreeOSMemory()
}
