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

package memqserver

import (
	"io/ioutil"
	"net/http"

	"github.com/kubernetes-up-and-running/kuard/pkg/apiutils"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	path   string
	broker *Broker
}

func NewServer(path string) *Server {
	return &Server{
		path:   path,
		broker: NewBroker(),
	}
}

func (s *Server) AddRoutes(router *httprouter.Router) {
	router.GET(s.path+"/stats", s.GetStats)
	router.PUT(s.path+"/queues/:queue", s.CreateQueue)
	router.DELETE(s.path+"/queues/:queue", s.DeleteQueue)
	router.POST(s.path+"/queues/:queue/drain", s.DrainQueue)
	router.POST(s.path+"/queues/:queue/dequeue", s.Dequeue)
	router.POST(s.path+"/queues/:queue/enqueue", s.Enqueue)
}

func (s *Server) CreateQueue(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	qName := p.ByName("queue")
	if len(qName) == 0 {
		http.Error(w, ErrEmptyName.Error(), http.StatusBadRequest)
		return
	}
	err := s.broker.CreateQueue(qName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *Server) DeleteQueue(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	qName := p.ByName("queue")
	if len(qName) == 0 {
		http.Error(w, ErrEmptyName.Error(), http.StatusBadRequest)
		return
	}
	err := s.broker.DeleteQueue(qName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *Server) DrainQueue(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	qName := p.ByName("queue")
	if len(qName) == 0 {
		http.Error(w, ErrEmptyName.Error(), http.StatusBadRequest)
		return
	}
	err := s.broker.DrainQueue(qName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s *Server) Enqueue(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	qName := p.ByName("queue")
	if len(qName) == 0 {
		http.Error(w, ErrEmptyName.Error(), http.StatusBadRequest)
		return
	}

	msg, err := s.broker.PutMessage(qName, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiutils.ServeJSON(w, msg)
}

func (s *Server) Dequeue(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	qName := p.ByName("queue")
	if len(qName) == 0 {
		http.Error(w, ErrEmptyName.Error(), http.StatusBadRequest)
		return
	}

	m, err := s.broker.GetMessage(qName)
	if err == ErrEmptyQueue {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiutils.ServeJSON(w, &m)
}

func (s *Server) GetStats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	stats := s.broker.Stats()
	apiutils.ServeJSON(w, &stats)
}
