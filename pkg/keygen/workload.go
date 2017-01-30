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
	"fmt"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
)

type workload struct {
	id        int
	c         Config
	generated int
	timeout   <-chan time.Time
	endTime   time.Time
	ctx       context.Context
	out       func(string)
}

func (w *workload) startWork() {
	w.logf("(ID %d) Workload starting", w.id)
	if w.c.TimeToRun > 0 {
		dur := time.Duration(w.c.TimeToRun) * time.Second
		w.endTime = time.Now().Add(dur)
		w.timeout = time.After(dur)
	}

	for !w.isDone() {
		w.itemDone(generateKey())
	}
}

func (w *workload) isDone() bool {
	select {
	case <-w.ctx.Done():
		w.done(true)
		return true
	case <-w.timeout:
		w.done(false)
		return true
	default:
	}

	if w.c.NumToGen > 0 && w.generated >= w.c.NumToGen {
		w.done(false)
		return true
	}
	return false
}

func (w *workload) done(canceled bool) {
	w.logf("(ID %d) Workload exiting", w.id)
	if !canceled && w.c.ExitOnComplete {
		os.Exit(w.c.ExitCode)
	}
}

func (w *workload) itemDone(desc string) {
	w.generated = w.generated + 1

	var count string
	if w.c.NumToGen > 0 {
		count = fmt.Sprintf(" %d/%d", w.generated, w.c.NumToGen)
	} else {
		count = fmt.Sprintf(" %d/Inf", w.generated)
	}

	timeleft := ""
	if w.timeout != nil {
		timeleft = " " + humanize.RelTime(time.Now(), w.endTime, "left", "overdue")
	}

	if len(desc) > 0 {
		desc = ": " + desc
	}

	w.logf("(ID %d%s%s) Item done%s", w.id, count, timeleft, desc)
}

func (w *workload) log(s string) {
	w.out(s)
}

func (w *workload) logf(format string, v ...interface{}) {
	w.out(fmt.Sprintf(format, v...))
}
