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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/kubernetes-up-and-running/kuard/pkg/memq"
)

var ErrEmptyQueue = errors.New("empty queue")
var ErrNotExist = errors.New("does not exist")
var ErrAlreadyExist = errors.New("already exists")
var ErrEmptyName = errors.New("empty name")

type Queue struct {
	Depth    int64
	Enqueued int64
	Dequeued int64
	Drained  int64
	Messages []*memq.Message
	mu       *sync.RWMutex
}

type Broker struct {
	Queues map[string]*Queue
	mu     *sync.RWMutex
}

func newStats() *memq.Stats {
	return &memq.Stats{
		Kind:   "stats",
		Queues: make([]memq.Stat, 0),
	}
}

func newMessage(body string) (*memq.Message, error) {
	id, err := uuid()
	if err != nil {
		return nil, err
	}
	m := &memq.Message{
		Kind:    "message",
		ID:      id,
		Body:    body,
		Created: time.Now()}
	return m, nil
}

func newQueue(name string) *Queue {
	return &Queue{
		Depth:    0,
		Messages: make([]*memq.Message, 0),
		mu:       &sync.RWMutex{},
	}
}

func NewBroker() *Broker {
	return &Broker{
		Queues: make(map[string]*Queue),
		mu:     &sync.RWMutex{},
	}
}

func (b *Broker) CreateQueue(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.Queues[name]; ok {
		return ErrAlreadyExist
	}

	b.Queues[name] = newQueue(name)

	return nil
}

func (b *Broker) DeleteQueue(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.Queues[name]; !ok {
		return ErrNotExist
	}
	delete(b.Queues, name)
	return nil
}

func (b *Broker) DrainQueue(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	q, ok := b.Queues[name]
	if !ok {
		return ErrNotExist
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.Messages = make([]*memq.Message, 0)
	q.Drained += q.Depth
	q.Depth = 0

	return nil
}

// getQueue safely gets a queue.  There is no guarantee that the queue won't be
// thown away (via DrainQueue or DeleteQueue) before it can be used.
func (b *Broker) getQueue(queue string) (*Queue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	q, ok := b.Queues[queue]
	if !ok {
		return nil, ErrNotExist
	}
	return q, nil
}

func (b *Broker) PutMessage(queue, body string) (*memq.Message, error) {
	q, err := b.getQueue(queue)
	if err != nil {
		return nil, err
	}

	message, err := newMessage(body)
	if err != nil {
		return nil, err
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	q.Messages = append(q.Messages, message)
	q.Depth++
	q.Enqueued++
	return message, nil
}

func (b *Broker) GetMessage(queue string) (*memq.Message, error) {
	q, err := b.getQueue(queue)
	if err != nil {
		return nil, err
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.Messages) < 1 {
		return nil, ErrEmptyQueue
	}
	var m *memq.Message
	m, q.Messages = q.Messages[0], q.Messages[1:]
	q.Depth--
	q.Dequeued++
	return m, nil
}

func (b *Broker) Stats() *memq.Stats {
	s := newStats()

	b.mu.RLock()
	defer b.mu.RUnlock()

	for name, q := range b.Queues {
		q.mu.RLock()
		stat := memq.Stat{
			Name:     name,
			Depth:    q.Depth,
			Enqueued: q.Enqueued,
			Dequeued: q.Dequeued,
			Drained:  q.Drained,
		}
		s.Queues = append(s.Queues, stat)
		q.mu.RUnlock()
	}
	return s
}

func uuid() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
