//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type wailsConfig struct {
	Info struct {
		ProductVersion string `json:"productVersion"`
	} `json:"Info"`
}

type packageJSON struct {
	Version string `json:"version"`
}

func main() {
	if len(os.Args) != 2 {
		fatalf("用法: go run scripts/check-version.go vX.Y.Z")
	}
	want := strings.TrimPrefix(strings.TrimSpace(os.Args[1]), "v")
	if want == "" {
		fatalf("版本号不能为空")
	}

	canonical := strings.TrimSpace(readFile("internal/appversion/version.txt"))
	var wails wailsConfig
	readJSON("wails.json", &wails)
	var frontend packageJSON
	readJSON("frontend/package.json", &frontend)
	var lock packageJSON
	readJSON("frontend/package-lock.json", &lock)

	versions := map[string]string{
		"Release tag":                want,
		"internal/appversion":        canonical,
		"wails.json":                 wails.Info.ProductVersion,
		"frontend/package.json":      frontend.Version,
		"frontend/package-lock.json": lock.Version,
	}
	for name, version := range versions {
		if version != want {
			fatalf("版本不一致: %s 为 %q，期望 %q", name, version, want)
		}
	}
	fmt.Printf("CatScope 版本一致性检查通过: v%s\n", want)
}

func readFile(path string) string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		fatalf("读取 %s 失败: %v", path, err)
	}
	return string(data)
}

func readJSON(path string, target any) {
	if err := json.Unmarshal([]byte(readFile(path)), target); err != nil {
		fatalf("解析 %s 失败: %v", path, err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
