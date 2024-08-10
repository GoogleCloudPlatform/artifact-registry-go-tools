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

// Package netrc provides functions to modify an netrc file.
package netrc

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
)

var arTokenConfigRegexp = regexp.MustCompile("machine (.*go.pkg.dev)\nlogin oauth2accesstoken\npassword (.*)")

func arTokenConfigPlaceholder(host string) string {
	return fmt.Sprintf(`machine %s
login oauth2accesstoken
password <oauth2accesstoken>
`, host)
}

func arJsonKeyConfig(host, base64key string) string {
	return fmt.Sprintf(`machine %s
login _json_key_base64
password %s
`, host, base64key)
}

// Load loads the path and contents of the .netrc file into memory.
func Load() (string, string, error) {
	netrcPath := os.Getenv("NETRC")
	if netrcPath == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("cannot load .netrc file: %v", err)
		}
		netrcPath = h
	}

	if !strings.HasSuffix(netrcPath, ".netrc") {
		netrcPath = path.Join(netrcPath, ".netrc")
	}

	if _, err := os.Stat(path.Dir(netrcPath)); err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf(".netrc directory does not exist: %w", err)
		}
		return "", "", fmt.Errorf("failed to load .netrc directory: %w", err)
	}

	data, err := ioutil.ReadFile(netrcPath)
	if os.IsNotExist(err) {
		//  The .netrc file does not exist; create a new one
		return netrcPath, "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("cannot load .netrc file: %v", err)
	}
	return netrcPath, string(data), nil
}

// Save saves the .netrc file.
func Save(netrc, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("Save: %v", err)
	}

	if _, err := f.WriteString(netrc); err != nil {
		return fmt.Errorf("Save: %v", err)
	}
	return nil
}

// Refresh updates the oauth tokens for all Artifact Registry Go endpoints.
func Refresh(netrc, token string) string {
	return arTokenConfigRegexp.ReplaceAllString(netrc, "machine $1\nlogin oauth2accesstoken\npassword "+token)
}

// AddConfigs adds a config for every new location in locations. If jsonKeyPath
// is an empty string, the config with use an oauth token for login.
func AddConfigs(locations []string, netrc, hostPattern, jsonKeyPath string) (string, error) {
	for _, l := range locations {
		h := fmt.Sprintf(hostPattern, l)
		match, err := regexp.MatchString("machine "+h+"\n", netrc)
		if err != nil {
			return "", fmt.Errorf("Internal: AddConfigs has error: %v", err)
		}
		if match {
			log.Printf("Warning: machine %s is already in the .netrc file, skipping\n", h)
			continue
		}

		var cfg string
		if jsonKeyPath == "" {
			cfg = arTokenConfigPlaceholder(h)
		} else {
			key, err := auth.EncodeJsonKey(jsonKeyPath)
			if err != nil {
				return "", fmt.Errorf("AddConfigs: %w", err)
			}
			cfg = arJsonKeyConfig(h, key)
		}
		if netrc != "" {
			netrc = netrc + "\n"
		}
		netrc = netrc + cfg
	}
	return netrc, nil
}
