package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestLocationURL(t *testing.T) {
	url := locationURL("us-central1", defaultHostPattern)

	if url != "https://us-central1-go.pkg.dev" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestAuthHeader_DefaultCredentials(t *testing.T) {
	header, err := keyAuthHeader("")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("header: %s", header)

	if !strings.HasPrefix(header, "Basic ") {
		t.Fatal(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(header[6:])
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("decoded: %s", decoded)

	if !strings.HasPrefix(string(decoded), "oauth2accesstoken:") {
		t.Fatal(err)
	}
}

func TestAuthHeader_JSONKey(t *testing.T) {
	header, err := keyAuthHeader("./test/dummy.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("header: %s", header)

	if !strings.HasPrefix(header, "Basic ") {
		t.Fatal(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(header[6:])
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("decoded: %s", decoded)

	if !strings.HasPrefix(string(decoded), "_json_key_base64:") {
		t.Fatal(err)
	}
}
