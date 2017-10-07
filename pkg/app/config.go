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
package app

import (
	"github.com/kubernetes-up-and-running/kuard/pkg/debugprobe"
	"github.com/kubernetes-up-and-running/kuard/pkg/keygen"
	"github.com/kubernetes-up-and-running/kuard/pkg/sitedata"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Debug        bool
	DebugRootDir string `mapstructure:"debug-sitedata-dir"`
	ServeAddr    string `mapstructure:"address"`
	TLSAddr      string `mapstructure:"tls-address"`
	TLSDir       string `mapstructure:"tls-dir"`

	KeyGen keygen.Config

	Liveness  debugprobe.ProbeConfig
	Readiness debugprobe.ProbeConfig
}

func (k *App) BindConfig(v *viper.Viper, fs *pflag.FlagSet) {
	k.kg.BindConfig(v, fs)

	k.live.BindConfig("liveness", v, fs)
	k.ready.BindConfig("readiness", v, fs)

	fs.Bool("debug", false, "Debug/devel mode")
	v.BindPFlag("debug", fs.Lookup("debug"))
	fs.String("debug-sitedata-dir", "./sitedata", "When in debug/dev mode, directory to find the static assets.")
	v.BindPFlag("debug-sitedata-dir", fs.Lookup("debug-sitedata-dir"))
	fs.String("address", ":8080", "The address to serve on")
	v.BindPFlag("address", fs.Lookup("address"))
	fs.String("tls-address", ":8443", "Address to serve TLS on if certs found.")
	v.BindPFlag("tls-address", fs.Lookup("tls-address"))
	fs.String("tls-dir", "/tls", "Directory to look to find TLS certs")
	v.BindPFlag("tls-dir", fs.Lookup("tls-dir"))
}

func (k *App) LoadConfig(v *viper.Viper) {
	err := v.UnmarshalExact(&k.c)
	if err != nil {
		panic(err)
	}

	k.live.SetConfig(k.c.Liveness)
	k.ready.SetConfig(k.c.Readiness)

	k.kg.LoadConfig(k.c.KeyGen)

	k.tg.SetConfig(k.c.Debug)
	sitedata.SetConfig(k.c.Debug, k.c.DebugRootDir)
}
