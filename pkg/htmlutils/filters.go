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

package htmlutils

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/dustin/go-humanize"
)

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"friendlytime": FriendlyTime,
		"reltime":      RelativeTime,
		"jsonstring":   JSONString,
	}
}

func FriendlyTime(t time.Time) string {
	return t.Format(time.Stamp)
}

func RelativeTime(t time.Time) string {
	return humanize.RelTime(t, time.Now(), "ago", "from now")
}

func JSONString(v interface{}) (template.JS, error) {
	a, err := json.Marshal(v)
	if err != nil {
		return template.JS(""), err
	}
	return template.JS(a), nil
}
