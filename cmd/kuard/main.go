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

package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/kubernetes-up-and-running/kuard/pkg/app"
	"github.com/kubernetes-up-and-running/kuard/pkg/version"
)

func main() {
	app := app.NewApp()

	v := viper.GetViper()

	app.BindConfig(v, pflag.CommandLine)

	pflag.Parse()

	log.Printf("Starting kuard version: %v", version.VERSION)
	log.Println(strings.Repeat("*", 70))
	log.Println("* WARNING: This server may expose sensitive")
	log.Println("* and secret information. Be careful.")
	log.Println(strings.Repeat("*", 70))

	dumpConfig(v)

	app.LoadConfig(v)
	app.Run()
}

func dumpConfig(v *viper.Viper) {
	b, err := json.MarshalIndent(v.AllSettings(), "", "  ")
	if err != nil {
		log.Printf("Could not dump config: %v", err)
		return
	}
	log.Printf("Config: \n%v\n", string(b))
}
