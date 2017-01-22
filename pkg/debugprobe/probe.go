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
	"strconv"
	"sync"
	"time"

	"github.com/jbeda/kuard/pkg/apiutils"
	"github.com/jbeda/kuard/pkg/htmlutils"
	"github.com/julienschmidt/httprouter"
)

const maxHistory = 20

type Probe struct {
	mu sync.Mutex

	tg *htmlutils.TemplateGroup

	basePath string
	// If failNext > 0, then fail next probe and decrement.  If failNext < 0, then
	// fail forever.
	failNext int
	history  []*ProbeHistory
}

type ProbeHistory struct {
	When time.Time
	Code int
}

// ProbeContext is appropriate for putting in a template for rendering.
type ProbeContext struct {
	BasePath string
	FailNext int
	History  []ProbeHistory
}

func New(base string, tg *htmlutils.TemplateGroup) *Probe {
	return &Probe{
		basePath: base,
		tg:       tg,
	}
}

func (p *Probe) AddRoutes(r *httprouter.Router) {
	r.GET(p.basePath, p.Handle)
	r.POST(p.basePath+"/config", p.Config)
	r.GET(p.basePath+"/render", p.Render)

	r.GET(p.basePath+"/api", p.APIGet)
	r.PUT(p.basePath+"/api", p.APIPut)
}

func (p *Probe) APIGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p.mu.Lock()
	defer p.mu.Unlock()

	s := &ProbeStatus{
		ProbePath: p.basePath,
		FailNext:  p.failNext,
	}
	l := len(p.history)
	s.History = make([]ProbeStatusHistory, l)
	for i, v := range p.history {
		h := &s.History[l-1-i]
		h.When = htmlutils.FriendlyTime(v.When)
		h.RelWhen = htmlutils.RelativeTime(v.When)
		h.Code = v.Code
	}

	apiutils.ServeJSON(w, s)
}

func (p *Probe) APIPut(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := &ProbeConfig{}

	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.failNext = c.FailNext

	w.WriteHeader(http.StatusOK)
}

func (p *Probe) Handle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := http.StatusOK
	message := "ok"
	if p.failNext > 0 {
		status = http.StatusInternalServerError
		p.failNext--
		message = fmt.Sprintf("fail, %d left", p.failNext)
	} else if p.failNext < 0 {
		status = http.StatusInternalServerError
		message = "fail, permanent"
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(message))
	p.recordRequest(r, status)
}

func (p *Probe) Config(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p.mu.Lock()
	defer p.mu.Unlock()

	rawFail := r.FormValue("fail")
	if len(rawFail) > 0 {
		fail, err := strconv.Atoi(rawFail)
		if err != nil {
			http.Error(w, "Could not parse 'fail' param", 400)
			return
		}
		p.failNext = fail
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (p *Probe) Render(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	p.tg.Render(w, "probe.html", p.GetContext())
}

func (p *Probe) recordRequest(_ *http.Request, code int) {
	entry := &ProbeHistory{
		When: time.Now(),
		Code: code,
	}
	p.history = append(p.history, entry)
	if len(p.history) > maxHistory {
		p.history = p.history[len(p.history)-maxHistory:]
	}
}

func (p *Probe) GetContext() *ProbeContext {
	p.mu.Lock()
	defer p.mu.Unlock()

	c := &ProbeContext{
		BasePath: p.basePath,
		FailNext: p.failNext,
	}
	l := len(p.history)
	c.History = make([]ProbeHistory, l)
	for i, v := range p.history {
		c.History[l-1-i] = *v
	}
	return c
}
