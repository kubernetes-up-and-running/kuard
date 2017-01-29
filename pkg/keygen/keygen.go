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
	"context"
	"log"
	"sync"

	"github.com/julienschmidt/httprouter"
)

const maxHistory = 20

type KeyGen struct {
	path    string
	config  Config
	history []string

	nextID int

	cancelFunc context.CancelFunc

	mu sync.Mutex
}

func New(path string) *KeyGen {
	kg := &KeyGen{
		path: path,
	}
	return kg
}

func (kg *KeyGen) AddRoutes(router *httprouter.Router) {
	router.GET(kg.path, kg.APIGet)
	router.PUT(kg.path, kg.APIPut)
}

func (kg *KeyGen) Restart() {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	// Cancel currently running workload
	if kg.cancelFunc != nil {
		kg.cancelFunc()
		kg.cancelFunc = nil
	}

	if kg.config.Enable {
		var ctx context.Context
		ctx, kg.cancelFunc = context.WithCancel(context.Background())

		w := workload{
			id:  kg.nextID,
			c:   kg.config,
			ctx: ctx,
			out: kg.WorkloadOutput,
		}
		kg.nextID++
		go w.startWork()
	}
}

func (kg *KeyGen) WorkloadOutput(s string) {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	log.Print(s)

	kg.history = append(kg.history, s)
	if len(kg.history) > maxHistory {
		kg.history = kg.history[len(kg.history)-maxHistory:]
	}
}
