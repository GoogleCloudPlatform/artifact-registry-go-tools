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
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/internal/auth"
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

	netrcPath = path.Join(netrcPath, ".netrc")
	data, err := os.ReadFile(netrcPath)
	if os.IsNotExist(err) {
		//  The .netrc file does not exist; create a new one
		return netrcPath, "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("cannot load .netrc file: %v", err)
	}
	return netrcPath, string(data), nil
}

// Save renames existing .netrc file as .netrc-old, and saves netrc as the new contents of the .netrc file.
func Save(netrc, path string) error {
	_, err := os.Stat(path + "-old")
	if err == nil { // delete the old file if Stat didn't fail
		if err := os.Remove(path + "-old"); err != nil {
			return fmt.Errorf("cannot delete %sold: %v", path, err)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Save: %v", err)
	}
	if err := os.Rename(path, path+"-old"); err != nil {
		return fmt.Errorf("rename .netrc to .netrc-old: %v", err)
	}
	if err := os.WriteFile(path, []byte(netrc), 0755); err != nil {
		return fmt.Errorf("write new .netrc: %v", err)
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
		if strings.Contains(netrc, l) {
			log.Printf("Warning: machine %s is already in the .netrc file, skipping\n", h)
		}

		var cfg string
		if jsonKeyPath == "" {
			cfg = arTokenConfigPlaceholder(h)
		} else {
			key, err := auth.EncodeJsonKey(jsonKeyPath)
			if err != nil {
				return "", fmt.Errorf("AddConfigs: %v", err)
			}
			cfg = arJsonKeyConfig(h, key)
		}
		netrc = netrc + "\n" + cfg
	}
	return netrc, nil
}
