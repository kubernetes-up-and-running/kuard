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

package memqclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/kubernetes-up-and-running/kuard/pkg/memq"
	"github.com/pkg/errors"
)

type Client struct {
	BaseServerURL string
}

func errorFromResponse(resp *http.Response) error {
	if resp.StatusCode >= 300 {
		return errors.Errorf("HTTP Error: %v", resp.Status)
	}
	return nil
}

func (c *Client) queueURL(queue string, s ...string) string {
	s = append([]string{"queues", queue}, s...)
	tail := path.Join(s...)
	return fmt.Sprintf("%s/%s", c.BaseServerURL, tail)
}

func (c *Client) CreateQueue(queue string) error {
	req, err := http.NewRequest("PUT", c.queueURL(queue), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return errorFromResponse(resp)
}

func (c *Client) DeleteQueue(queue string) error {
	req, err := http.NewRequest("DELETE", c.queueURL(queue), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return errorFromResponse(resp)
}

func (c *Client) DrainQueue(queue string) error {
	req, err := http.NewRequest("POST", c.queueURL(queue, "drain"), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return errorFromResponse(resp)
}

func (c *Client) Enqueue(queue, data string) (*memq.Message, error) {
	req, err := http.NewRequest(
		"POST", c.queueURL(queue, "enqueue"),
		bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	err = errorFromResponse(resp)
	if err != nil {
		return nil, err
	}

	m := &memq.Message{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Dequeue takes an item off of queue from the server.  If a nil message is
// returned with no error then the queue is empty.
func (c *Client) Dequeue(queue string) (*memq.Message, error) {
	req, err := http.NewRequest("POST", c.queueURL(queue, "dequeue"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	err = errorFromResponse(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	m := &memq.Message{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *Client) Stats() (*memq.Stats, error) {
	req, err := http.NewRequest("GET", c.BaseServerURL+"/stats", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	err = errorFromResponse(resp)
	if err != nil {
		return nil, err
	}

	s := &memq.Stats{}
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
