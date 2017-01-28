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

// ProbeStatus is returned from a GET to this API endpoing
type ProbeStatus struct {
	ProbePath string               `json:"probePath"`
	FailNext  int                  `json:"failNext"`
	History   []ProbeStatusHistory `json:"history"`
}

// ProbeStatusHistory is a record of a probe call
type ProbeStatusHistory struct {
	ID      int    `json:"id"`
	When    string `json:"when"`
	RelWhen string `json:"relWhen"`
	Code    int    `json:"code"`
}
