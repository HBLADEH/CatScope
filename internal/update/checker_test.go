package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckerSelectsReleaseChannelAndAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
          {"tag_name":"v0.7.0-preview","name":"Preview","prerelease":true,"html_url":"https://github.com/HBLADEH/CatScope/releases/tag/v0.7.0-preview","assets":[]},
          {"tag_name":"v0.6.6","name":"Stable","body":"notes","published_at":"2026-07-16T00:00:00Z","html_url":"https://github.com/HBLADEH/CatScope/releases/tag/v0.6.6","assets":[
            {"name":"CatScope-v0.6.6-windows-amd64.exe","browser_download_url":"https://github.com/HBLADEH/CatScope/releases/download/v0.6.6/CatScope.exe"},
            {"name":"CatScope-v0.6.6-windows-amd64.exe.sha256","browser_download_url":"https://github.com/HBLADEH/CatScope/releases/download/v0.6.6/CatScope.exe.sha256"}
          ]}
        ]`))
	}))
	defer server.Close()

	checker := &Checker{Client: server.Client(), APIURL: server.URL, GOOS: "windows", GOARCH: "amd64", Version: "0.6.5"}
	info, err := checker.Check(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.LatestVersion != "0.6.6" || !info.CanAutoInstall {
		t.Fatalf("unexpected stable update: %+v", info)
	}

	info, err = checker.Check(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if info.LatestVersion != "0.7.0-preview" || !info.Prerelease {
		t.Fatalf("unexpected preview update: %+v", info)
	}
}

func TestCheckerDoesNotDowngrade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"tag_name":"v0.6.5","name":"Older","html_url":"https://github.com/HBLADEH/CatScope/releases/tag/v0.6.5"}]`))
	}))
	defer server.Close()

	checker := &Checker{Client: server.Client(), APIURL: server.URL, GOOS: "windows", GOARCH: "amd64", Version: "0.6.6"}
	info, err := checker.Check(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if info.Available {
		t.Fatalf("older release must not be offered: %+v", info)
	}
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		left  string
		right string
		want  int
	}{
		{"v0.6.6", "0.6.5", 1},
		{"0.6.6-preview", "0.6.6", -1},
		{"0.6.6", "0.6.6-preview", 1},
		{"1.0.0", "1.0.0", 0},
	}
	for _, test := range tests {
		t.Run(test.left+"_"+test.right, func(t *testing.T) {
			got := compareVersions(test.left, test.right)
			if got < 0 {
				got = -1
			} else if got > 0 {
				got = 1
			}
			if got != test.want {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
			}
		})
	}
}

func TestDownloadAndChecksumHelpers(t *testing.T) {
	content := []byte("portable exe contents")
	sum := sha256.Sum256(content)
	expected := hex.EncodeToString(sum[:])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			_, _ = fmt.Fprintf(w, "%s  CatScope-v0.6.6-windows-amd64.exe", expected)
			return
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	checker := &Checker{Client: server.Client(), Version: "0.6.6"}
	checksum, err := checker.fetchChecksum(context.Background(), server.URL+"/file.sha256", "CatScope-v0.6.6-windows-amd64.exe")
	if err != nil {
		t.Fatal(err)
	}
	if checksum != expected {
		t.Fatalf("checksum = %q, want %q", checksum, expected)
	}

	destination := filepath.Join(t.TempDir(), "CatScope.exe.new")
	actual, err := checker.downloadFile(context.Background(), server.URL+"/file.exe", destination)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("download hash = %q, want %q", actual, expected)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("downloaded content = %q", got)
	}
}

func TestTrustedDownloadURL(t *testing.T) {
	tests := map[string]bool{
		"https://github.com/HBLADEH/CatScope/releases/download/v0.6.6/file.exe": true,
		"https://objects.githubusercontent.com/file":                            true,
		"http://github.com/file":                                                false,
		"https://github.com.evil.example/file":                                  false,
		"https://example.com/file":                                              false,
	}
	for rawURL, want := range tests {
		if got := trustedDownloadURL(rawURL); got != want {
			t.Errorf("trustedDownloadURL(%q) = %v, want %v", rawURL, got, want)
		}
	}
}
