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

package debugprobe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/kubernetes-up-and-running/kuard/pkg/apiutils"
	"github.com/kubernetes-up-and-running/kuard/pkg/htmlutils"
)

const maxHistory = 20

type Probe struct {
	basePath string
	mu       sync.Mutex

	lastID int

	c       ProbeConfig
	history []*ProbeHistory
}

type ProbeHistory struct {
	ID   int
	When time.Time
	Code int
}

func New() *Probe {
	return &Probe{}
}

func (p *Probe) AddRoutes(r *httprouter.Router, base string) {
	r.GET(base, p.Handle)
	r.GET(base+"/api", p.APIGet)
	r.PUT(base+"/api", p.APIPut)

	if p.basePath != "" {
		p.basePath = base
	}
}

func (p *Probe) APIGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lockedGet(w, r)
}

func (p *Probe) lockedGet(w http.ResponseWriter, r *http.Request) {
	s := &ProbeStatus{
		ProbePath: p.basePath,
		FailNext:  p.c.FailNext,
	}
	l := len(p.history)
	s.History = make([]ProbeStatusHistory, l)
	for i, v := range p.history {
		h := &s.History[l-1-i]
		h.ID = v.ID
		h.When = htmlutils.FriendlyTime(v.When)
		h.RelWhen = htmlutils.RelativeTime(v.When)
		h.Code = v.Code
	}

	apiutils.ServeJSON(w, s)
}

func (p *Probe) APIPut(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	c := ProbeConfig{}

	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.SetConfig(c)

	p.APIGet(w, r, params)
}

func (p *Probe) Handle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := http.StatusOK
	message := "ok"
	if p.c.FailNext > 0 {
		status = http.StatusInternalServerError
		p.c.FailNext--
		message = fmt.Sprintf("fail, %d left", p.c.FailNext)
	} else if p.c.FailNext < 0 {
		status = http.StatusInternalServerError
		message = "fail, permanent"
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(message))
	p.recordRequest(r, status)
}

func (p *Probe) recordRequest(_ *http.Request, code int) {
	p.lastID++
	entry := &ProbeHistory{
		ID:   p.lastID,
		When: time.Now(),
		Code: code,
	}
	p.history = append(p.history, entry)
	if len(p.history) > maxHistory {
		p.history = p.history[len(p.history)-maxHistory:]
	}
}
