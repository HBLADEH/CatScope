package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIURL      = "https://api.github.com/repos/HBLADEH/CatScope/releases"
	defaultReleaseURL  = "https://github.com/HBLADEH/CatScope/releases"
	maxReleaseResponse = 4 << 20
)

type Info struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	Available       bool   `json:"available"`
	Prerelease      bool   `json:"prerelease"`
	ReleaseName     string `json:"releaseName"`
	ReleaseNotes    string `json:"releaseNotes"`
	PublishedAt     string `json:"publishedAt"`
	ReleaseURL      string `json:"releaseUrl"`
	AssetURL        string `json:"assetUrl,omitempty"`
	AssetName       string `json:"assetName,omitempty"`
	ChecksumURL     string `json:"checksumUrl,omitempty"`
	CanAutoInstall  bool   `json:"canAutoInstall"`
	AutoInstallHint string `json:"autoInstallHint,omitempty"`
}

type Checker struct {
	Client  *http.Client
	APIURL  string
	GOOS    string
	GOARCH  string
	Version string
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	HTMLURL     string        `json:"html_url"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func NewChecker(version string) *Checker {
	return &Checker{
		Client:  &http.Client{Timeout: 15 * time.Second},
		APIURL:  defaultAPIURL,
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		Version: version,
	}
}

func (c *Checker) Check(ctx context.Context, includePrerelease bool) (Info, error) {
	info := Info{CurrentVersion: normalizeVersion(c.Version), ReleaseURL: defaultReleaseURL}
	if !validVersion(info.CurrentVersion) {
		return info, fmt.Errorf("当前版本号无效: %q", c.Version)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIURL, nil)
	if err != nil {
		return info, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "CatScope/"+info.CurrentVersion)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return info, fmt.Errorf("检查更新失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		return info, fmt.Errorf("检查更新失败: GitHub 返回 HTTP %d", resp.StatusCode)
	}

	var releases []githubRelease
	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxReleaseResponse))
	if err := decoder.Decode(&releases); err != nil {
		return info, fmt.Errorf("解析版本信息失败: %w", err)
	}

	var selected *githubRelease
	for i := range releases {
		release := &releases[i]
		if release.Draft || (!includePrerelease && release.Prerelease) || !validVersion(release.TagName) {
			continue
		}
		if selected == nil || compareVersions(release.TagName, selected.TagName) > 0 {
			selected = release
		}
	}
	if selected == nil {
		return info, errors.New("没有找到可用的 CatScope Release")
	}

	info.LatestVersion = normalizeVersion(selected.TagName)
	info.Available = compareVersions(selected.TagName, info.CurrentVersion) > 0
	info.Prerelease = selected.Prerelease
	info.ReleaseName = strings.TrimSpace(selected.Name)
	info.ReleaseNotes = strings.TrimSpace(selected.Body)
	info.PublishedAt = selected.PublishedAt
	if strings.HasPrefix(selected.HTMLURL, "https://github.com/HBLADEH/CatScope/") {
		info.ReleaseURL = selected.HTMLURL
	}

	assetName := assetNameFor(selected.TagName, c.GOOS, c.GOARCH)
	for _, asset := range selected.Assets {
		switch asset.Name {
		case assetName:
			info.AssetName = asset.Name
			info.AssetURL = asset.BrowserDownloadURL
		case assetName + ".sha256":
			info.ChecksumURL = asset.BrowserDownloadURL
		}
	}

	info.CanAutoInstall = info.Available && c.GOOS == "windows" && info.AssetURL != "" && info.ChecksumURL != ""
	if info.Available && !info.CanAutoInstall {
		switch {
		case c.GOOS != "windows":
			info.AutoInstallHint = "当前平台暂不支持应用内替换，请打开 Release 页面下载安装。"
		case info.AssetURL == "" || info.ChecksumURL == "":
			info.AutoInstallHint = "Release 缺少当前平台的 EXE 或 SHA256 文件，请打开发布页面下载。"
		}
	}
	return info, nil
}

func (c *Checker) httpClient() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func assetNameFor(tag, goos, goarch string) string {
	tag = "v" + normalizeVersion(tag)
	switch goos {
	case "windows":
		return fmt.Sprintf("CatScope-%s-windows-%s.exe", tag, goarch)
	case "darwin":
		return fmt.Sprintf("CatScope-%s-macos-universal.dmg", tag)
	default:
		return fmt.Sprintf("CatScope-%s-%s-%s", tag, goos, goarch)
	}
}

type parsedVersion struct {
	major int
	minor int
	patch int
	pre   string
}

func normalizeVersion(value string) string {
	return strings.TrimPrefix(strings.TrimSpace(value), "v")
}

func validVersion(value string) bool {
	_, ok := parseVersion(value)
	return ok
}

func parseVersion(value string) (parsedVersion, bool) {
	value = normalizeVersion(value)
	core, pre, _ := strings.Cut(value, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return parsedVersion{}, false
	}
	values := make([]int, 3)
	for i, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return parsedVersion{}, false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return parsedVersion{}, false
		}
		values[i] = n
	}
	if strings.ContainsAny(pre, " \t\r\n+") {
		return parsedVersion{}, false
	}
	return parsedVersion{major: values[0], minor: values[1], patch: values[2], pre: pre}, true
}

func compareVersions(left, right string) int {
	a, okA := parseVersion(left)
	b, okB := parseVersion(right)
	if !okA || !okB {
		return strings.Compare(normalizeVersion(left), normalizeVersion(right))
	}
	for _, pair := range [][2]int{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if a.pre == b.pre {
		return 0
	}
	if a.pre == "" {
		return 1
	}
	if b.pre == "" {
		return -1
	}
	return strings.Compare(a.pre, b.pre)
}
