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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/julienschmidt/httprouter"
)

const maxHistory = 20

type KeyGen struct {
	mu             sync.Mutex
	path           string
	config         Config
	history        []History
	nextHistoryID  int
	nextWorkloadID int
	cancelFunc     context.CancelFunc
}

func New(path string) *KeyGen {
	kg := &KeyGen{
		path:    path,
		history: []History{},
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

		if len(kg.config.MemQQueue) > 0 && len(kg.config.MemQServer) > 0 {
			w := newMemQWorker(ctx, kg.config, kg.WorkloadOutput)
			go w.startWork()
		} else {
			w := workload{
				id:  kg.nextWorkloadID,
				c:   kg.config,
				ctx: ctx,
				out: kg.WorkloadOutput,
			}
			kg.nextWorkloadID++
			go w.startWork()
		}
	}
}

func (kg *KeyGen) WorkloadOutput(s string) {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	log.Print(s)

	kg.history = append(kg.history, History{ID: kg.nextHistoryID, Data: s})
	if len(kg.history) > maxHistory {
		kg.history = kg.history[len(kg.history)-maxHistory:]
	}

	kg.nextHistoryID++
}

func generateKey() string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Sprintf("Error generating key: %v", err)
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Sprintf("Error generating ssh key: %v", err)
	}

	return ssh.FingerprintSHA256(pub)
}
