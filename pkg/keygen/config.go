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
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config is the input parameters to the keygen workload.
type Config struct {
	Enable bool `json:"enable"`

	// This limits the amount of work to do.  The workload will stop when either
	// of these is complete.  Zero is interpreted as "infinity".  TimeToRun is in
	// seconds.
	NumToGen  int `json:"numToGen" mapstructure:"num-to-gen"`
	TimeToRun int `json:"timeToRun" mapstructure:"time-to-run"`

	// If both of these variables are set, then the keygen worker will pull work
	// items off of the MemQ.  If there is an error it will keep retrying with a
	// small pause.  If the queue is empty, and exitOnComplete is set, then the
	// process will exit.  If MemQ is used, then NumToGen and TimeToRun are
	// ignored.
	MemQServer string `json:"memQServer" mapstructure:"memq-server"`
	MemQQueue  string `json:"memQQueue" mapstructure:"memq-queue"`

	// What should happen when the workload is complete?
	ExitOnComplete bool `json:"exitOnComplete" mapstructure:"exit-on-complete"`
	ExitCode       int  `json:"exitCode" mapstructure:"exit-code"`
}

func (kg *KeyGen) BindConfig(v *viper.Viper, fs *pflag.FlagSet) {
	v.Set("keygen", map[string]interface{}{})
	fs.Bool("keygen-enable", false, "Enable KeyGen workload")
	fs.Int("keygen-num-to-gen", 0, "The number of keys to generate. Set to 0 for infinite")
	fs.Int("keygen-time-to-run", 0, "The target run time in seconds. Set to 0 for infinite")
	fs.String("keygen-memq-server", "", "The MemQ server to draw work items from.  If MemQ is used, other limits are ignored.")
	fs.String("keygen-memq-queue", "", "The MemQ server queue to use. If MemQ is used, other limits are ignored.")
	fs.Bool("keygen-exit-on-complete", false, "Exit after workload is complete")
	fs.Int("keygen-exit-code", 0, "Exit code when workload complete")

	// Iterate through all flags and register with the passed in viper.  Only
	// apply to those flags with our prefix but strip it out.
	fs.VisitAll(func(f *pflag.Flag) {
		name := strings.TrimPrefix(f.Name, "keygen-")
		if name != f.Name {
			v.BindPFlag("keygen."+name, f)
		}
	})
}

func (kg *KeyGen) LoadConfig(c Config) {
	kg.config = c

	kg.Restart()
}
