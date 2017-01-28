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

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ProbeConfig is used to configure how the probe will respond
type ProbeConfig struct {
	// If failNext > 0, then fail next probe and decrement.  If failNext < 0, then
	// fail forever.
	FailNext int `json:"failNext" mapstructure:"fail-next"`
}

func (p *Probe) SetConfig(c ProbeConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.c = c
}

func (p *Probe) BindConfig(prefix string, v *viper.Viper, fs *pflag.FlagSet) {
	fs.Int(prefix+"-fail-next", 0, "Fail the next N probes. 0 is succeed forever. <0 is fail forever.")
	v.BindPFlag(prefix+".fail-next", fs.Lookup(prefix+"-fail-next"))
}
