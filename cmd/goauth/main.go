// Copyright 2022 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/artifact-registry-go-tools/pkg/auth"
)

const help = `
Handle Go authentication with Google Cloud Artifact Registry Go Repositories.

Add to your GOAUTH environment variable:

  export GOAUTH="sh -c 'GOPROXY=direct go run github.com/GoogleCloudPlatform/artifact-registry-go-tools/cmd/goauth@latest <location>'"

To support multiple locations, add the command multiple times to the GOAUTH variable (semicolon-separated).

For more details, see https://pkg.go.dev/cmd/go@master#hdr-GOAUTH_environment_variable`

const defaultHostPattern = "%s-go.pkg.dev"

func main() {
	jsonKey := flag.String("json_key", "", "path to the json key of the service account used for this location. Leave empty to use the oauth token instead.")
	hostPattern := flag.String("host_pattern", defaultHostPattern, "Artifact Registry server host pattern, where %s will be replaced by a location string.")

	flag.Parse()

	location := flag.Arg(0)
	if location == "" {
		fmt.Fprintln(os.Stderr, help)
		return
	}
	if strings.HasPrefix(location, "https://") {
		log.Println("Location has to be a Google Cloud region, e.g. 'us-central1'.")
		os.Exit(2)
	}

	// generate the authentication header
	urlLine := locationURL(location, *hostPattern)
	authHeader, err := keyAuthHeader(*jsonKey)
	if err != nil {
		log.Println(err)
		os.Exit(3)
	}

	// send the Go authentication information
	fmt.Printf("%s\n\nAuthorization: %s\n\n", urlLine, authHeader)
}

func locationURL(location string, hostPattern string) string {
	host := fmt.Sprintf(hostPattern, location)

	return fmt.Sprintf("https://%s", host)
}

func keyAuthHeader(jsonKeyPath string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if jsonKeyPath != "" {
		key, err := auth.EncodeJsonKey(jsonKeyPath)
		if err != nil {
			return "", fmt.Errorf("failed to encode JSON key: %w", err)
		}

		return basicAuthHeader("_json_key_base64", key), nil
	} else {
		token, err := auth.Token(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get oauth token: %w", err)
		}

		return basicAuthHeader("oauth2accesstoken", token), nil
	}
}

func basicAuthHeader(username, password string) string {
	a := fmt.Sprintf("%s:%s", username, password)
	b := base64.StdEncoding.EncodeToString([]byte(a))

	return fmt.Sprintf("Basic %s", b)
}
