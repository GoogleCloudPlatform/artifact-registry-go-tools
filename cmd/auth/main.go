// Copyright 2022 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/internal/auth"
	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/internal/netrc"
)

const help = `
Update your .netrc file to work with Google Cloud Artifact Registry Go Repositories.

Commands:

* refresh, to refresh oauth tokens for Artifact Registry Go endpoints.
* add-locations, to add new regional Artifact Registry Go endpoints to the netrc file.`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(help)
		return
	}
	switch os.Args[1] {
	case "refresh":
		refresh()
	case "add-locations":
		addLocationFlags := flag.NewFlagSet("add-location", flag.ExitOnError)
		jsonKey := addLocationFlags.String("json_key", "", "path to the json key of the service account used for this location. Leave empty to use the oauth token instead.")
		hostPattern := addLocationFlags.String("host_pattern", "%s-go.pkg.dev", "Artifact Registry server host pattern, where %s will be replaced by a location string.")
		locations := addLocationFlags.String("locations", "", "Required. A list of comma-separated location strings to regional Artifact Registry Go endpoints to the netrc file.")
		addLocationFlags.Parse(os.Args[2:])
		addLocations(*locations, *jsonKey, *hostPattern)
	case "help", "-help", "--help":
		fmt.Println(help)
	default:
		fmt.Printf("unknown command %q. Please rerun the tool with `--help`\n", os.Args[1])
	}
}

func refresh() {
	ctx := context.Background()
	ctx, cf := context.WithTimeout(ctx, 30*time.Second)
	defer cf()

	p, config, err := netrc.Load()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	token, err := auth.Token(ctx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	config = netrc.Refresh(config, token)
	if err := netrc.Save(config, p); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Refresh completed.")
}

func addLocations(locations, jsonKeyPath, hostPattern string) {
	if strings.Count(hostPattern, "%s") != 1 {
		errMsg := "-host_pattern must have one and only one %%s in it."
		log.Println(errMsg)
		return
	}
	if locations == "" {
		log.Println("-locations is required.")
		return
	}
	ll := strings.Split(locations, ",")
	p, config, err := netrc.Load()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	newCfg, err := netrc.AddConfigs(ll, config, hostPattern, jsonKeyPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := netrc.Save(newCfg, p); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Add locations completed.")
}
