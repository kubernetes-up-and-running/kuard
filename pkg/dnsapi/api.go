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

package dnsapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kubernetes-up-and-running/kuard/pkg/apiutils"
	"github.com/miekg/dns"
)

type DNSAPI struct {
}

// DNSResponse is returned from a GET to this API endpoing
type DNSResponse struct {
	Results string `json:"result"`
}

type DNSRequest struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func New() *DNSAPI {
	return &DNSAPI{}
}

func (e *DNSAPI) AddRoutes(r *httprouter.Router, base string) {
	r.POST(base+"/api", e.APIGet)
}

func (e *DNSAPI) APIGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	dreq := &DNSRequest{}

	err := json.NewDecoder(r.Body).Decode(dreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := dnsQuery(dreq.Type, dreq.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dresp := &DNSResponse{
		Results: result,
	}

	apiutils.ServeJSON(w, dresp)
}

func dnsQuery(t string, name string) (string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return "", err
	}

	c := new(dns.Client)
	m := new(dns.Msg)

	qtype, ok := dns.StringToType[strings.ToUpper(t)]
	if !ok {
		return "", fmt.Errorf("Unknown DNS type: %v", t)
	}

	if len(name) == 0 {
		name = "."
	}

	names := []string{}
	if dns.IsFqdn(name) {
		names = append(names, name)
	} else {
		// TODO: respect NDOTS
		for _, s := range config.Search {
			names = append(names, name+"."+s)
		}
		names = append(names, name)
	}

	var r *dns.Msg
	for _, name := range names {
		m.SetQuestion(dns.Fqdn(name), qtype)
		m.RecursionDesired = true
		r, _, err = c.Exchange(m, config.Servers[0]+":"+config.Port)
		if err != nil {
			return "", err
		}
		if len(r.Answer) > 0 {
			return r.String(), nil
		}
	}
	return r.String(), nil
}
