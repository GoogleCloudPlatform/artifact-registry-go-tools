package netrc

import (
	_ "embed"
	"errors"
	"os"
	"path"
	"testing"
)

func TestAddConfigs(t *testing.T) {
	cases := []struct {
		name          string
		existingNetrc string
		locations     []string
		hostPattern   string
		jsonKeyPath   string
		wantNetrc     string
		wantErr       error
	}{
		{
			name:      "add the first location",
			locations: []string{"us-west1"},
			wantNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>
`,
		},
		{
			name:        "add the first location with host pattern",
			locations:   []string{"us-west1"},
			hostPattern: "%s-different-go.pkg.dev",
			wantNetrc: `machine us-west1-different-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>
`,
		},
		{
			name:        "add the first location with json key",
			locations:   []string{"us-west1"},
			jsonKeyPath: "testdata/key.json",
			wantNetrc: `machine us-west1-go.pkg.dev
login _json_key_base64
password ewogICAgInRlc3Qta2V5IjogInRlc3QtdmFsdWUiCn0=
`,
		},
		{
			name:        "json key does not exist",
			locations:   []string{"us-west1"},
			jsonKeyPath: "testdata/not-a-key.json",
			wantErr:     os.ErrNotExist,
		},
		{
			name:      "add two locations",
			locations: []string{"us-west1", "europe-east1"},
			wantNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>

machine europe-east1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>
`,
		},
		{
			name:      "add locations to existing config",
			locations: []string{"us-west1", "europe-east1"},
			existingNetrc: `machine asia-south1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken
`,
			wantNetrc: `machine asia-south1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken

machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>

machine europe-east1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>
`,
		},
		{
			name:      "skip existing configs",
			locations: []string{"us-west1", "europe-east1"},
			existingNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken
`,
			wantNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken

machine europe-east1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hostPattern := tc.hostPattern
			if hostPattern == "" {
				hostPattern = "%s-go.pkg.dev"
			}
			netrc, err := AddConfigs(tc.locations, tc.existingNetrc, hostPattern, tc.jsonKeyPath)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got error %v, want error %v", err, tc.wantErr)
			}
			if err == nil {
				if netrc != tc.wantNetrc {
					t.Errorf("unexpected netrc: got %q, want %q", netrc, tc.wantNetrc)
				}
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	cases := []struct {
		name          string
		exsitingNetrc string
		token         string
		wantNetrc     string
	}{
		{
			name: "replace placeholder",
			exsitingNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>`,
			token: "a-token",
			wantNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password a-token`,
		},
		{
			name: "replace old token",
			exsitingNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password old-token`,
			token: "new-token",
			wantNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password new-token`,
		},
		{
			name: "replace all ending in go.pkg.dev",
			exsitingNetrc: `machine another-env-us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>`,
			token: "a-token",
			wantNetrc: `machine another-env-us-west1-go.pkg.dev
login oauth2accesstoken
password a-token`,
		},
		{
			name: "keep json key entry unchanged",
			exsitingNetrc: `machine us-west1-go.pkg.dev
login _json_key_base64
password ewogICAgInRlc3Qta2V5IjogInRlc3QtdmFsdWUiCn0=`,
			token: "a-token",
			wantNetrc: `machine us-west1-go.pkg.dev
login _json_key_base64
password ewogICAgInRlc3Qta2V5IjogInRlc3QtdmFsdWUiCn0=`,
		},
		{
			name: "keep non go.pkg.dev unchanged",
			exsitingNetrc: `machine example.com
login oauth2accesstoken
password <oauth2accesstoken>`,
			token: "a-token",
			wantNetrc: `machine example.com
login oauth2accesstoken
password <oauth2accesstoken>`,
		},
		{
			name: "put everything together",
			exsitingNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>

machine us-west1-go.pkg.dev
login oauth2accesstoken
password old-token

machine another-env-us-west1-go.pkg.dev
login oauth2accesstoken
password <oauth2accesstoken>

machine us-west1-go.pkg.dev
login _json_key_base64
password ewogICAgInRlc3Qta2V5IjogInRlc3QtdmFsdWUiCn0=

machine example.com
login oauth2accesstoken
password <oauth2accesstoken>`,
			token: "a-token",
			wantNetrc: `machine us-west1-go.pkg.dev
login oauth2accesstoken
password a-token

machine us-west1-go.pkg.dev
login oauth2accesstoken
password a-token

machine another-env-us-west1-go.pkg.dev
login oauth2accesstoken
password a-token

machine us-west1-go.pkg.dev
login _json_key_base64
password ewogICAgInRlc3Qta2V5IjogInRlc3QtdmFsdWUiCn0=

machine example.com
login oauth2accesstoken
password <oauth2accesstoken>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if gotNetrc := Refresh(tc.exsitingNetrc, tc.token); gotNetrc != tc.wantNetrc {
				t.Errorf("unexpected netrc: got %v, want %v", gotNetrc, tc.wantNetrc)
			}
		})
	}
}

//go:embed testdata/load_test/has_netrc_file_dir/.netrc
var hasNetrcFileDirNetrcContent string

func TestLoad(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() = %v", err)
	}
	testDataDir := path.Join(currentDir, "testdata", "load_test")

	cases := []struct {
		name             string
		netrcPath        string
		wantNetrcPath    string
		wantNetrcContent string
		wantErr          error
	}{
		{
			name:             "NETRC env points to existing directory",
			netrcPath:        path.Join(testDataDir, "empty_dir"),
			wantNetrcPath:    path.Join(testDataDir, "empty_dir", ".netrc"),
			wantNetrcContent: "",
			wantErr:          nil,
		},
		{
			name:             "NETRC env points to non-existent directory",
			netrcPath:        path.Join(testDataDir, "non_existent_dir"),
			wantNetrcPath:    "",
			wantNetrcContent: "",
			wantErr:          os.ErrNotExist,
		},
		{
			name:             "NETRC env points to existing file",
			netrcPath:        path.Join(testDataDir, "has_netrc_file_dir", ".netrc"),
			wantNetrcPath:    path.Join(testDataDir, "has_netrc_file_dir", ".netrc"),
			wantNetrcContent: hasNetrcFileDirNetrcContent,
			wantErr:          nil,
		},
		{
			name:             "NETRC env points to non-existent file",
			netrcPath:        path.Join(testDataDir, "empty_dir", ".netrc"),
			wantNetrcPath:    path.Join(testDataDir, "empty_dir", ".netrc"),
			wantNetrcContent: "",
			wantErr:          nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if envErr := os.Setenv("NETRC", tc.netrcPath); envErr != nil {
				t.Fatalf("os.Setenv() = %v", envErr)
			}

			gotNetrcPath, gotNetrcContent, gotErr := Load()

			if gotNetrcPath != tc.wantNetrcPath {
				t.Errorf("unexpected netrc path: got %v, want %v", gotNetrcPath, tc.wantNetrcPath)
			}
			if gotNetrcContent != tc.wantNetrcContent {
				t.Errorf("unexpected netrc content: got %v, want %v", gotNetrcContent, tc.wantNetrcContent)
			}
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("got error %v, want error %v", gotErr, tc.wantErr)
			}
		})
	}
}
