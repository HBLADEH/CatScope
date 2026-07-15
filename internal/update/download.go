package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const maxUpdateSize = 250 << 20

type Download struct {
	Directory string
	Path      string
	SHA256    string
}

func (c *Checker) Download(ctx context.Context, info Info) (result Download, err error) {
	if !info.CanAutoInstall || info.AssetName == "" {
		return result, errors.New("当前版本不能自动安装")
	}
	if !trustedDownloadURL(info.AssetURL) || !trustedDownloadURL(info.ChecksumURL) {
		return result, errors.New("Release 下载地址不可信")
	}

	expected, err := c.fetchChecksum(ctx, info.ChecksumURL, info.AssetName)
	if err != nil {
		return result, err
	}

	dir, err := os.MkdirTemp("", "catscope-update-")
	if err != nil {
		return result, fmt.Errorf("创建升级临时目录失败: %w", err)
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(dir)
		}
	}()

	destination := filepath.Join(dir, "CatScope.exe.new")
	actual, err := c.downloadFile(ctx, info.AssetURL, destination)
	if err != nil {
		return result, err
	}
	if !strings.EqualFold(actual, expected) {
		return result, fmt.Errorf("升级文件 SHA256 校验失败: 期望 %s，实际 %s", expected, actual)
	}
	return Download{Directory: dir, Path: destination, SHA256: actual}, nil
}

func (c *Checker) fetchChecksum(ctx context.Context, rawURL, assetName string) (string, error) {
	body, err := c.fetch(ctx, rawURL, 64<<10)
	if err != nil {
		return "", fmt.Errorf("下载 SHA256 文件失败: %w", err)
	}
	fields := strings.Fields(string(body))
	if len(fields) == 0 || len(fields[0]) != sha256.Size*2 {
		return "", errors.New("SHA256 文件格式无效")
	}
	if _, err := hex.DecodeString(fields[0]); err != nil {
		return "", errors.New("SHA256 文件格式无效")
	}
	if len(fields) > 1 && strings.TrimPrefix(fields[1], "*") != assetName {
		return "", errors.New("SHA256 文件中的文件名与升级资产不一致")
	}
	return strings.ToLower(fields[0]), nil
}

func (c *Checker) downloadFile(ctx context.Context, rawURL, destination string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CatScope/"+normalizeVersion(c.Version))
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("下载升级文件失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载升级文件失败: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxUpdateSize {
		return "", errors.New("升级文件超过允许的大小")
	}

	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return "", fmt.Errorf("创建升级文件失败: %w", err)
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(resp.Body, maxUpdateSize+1))
	closeErr := file.Close()
	if copyErr != nil {
		return "", fmt.Errorf("保存升级文件失败: %w", copyErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("关闭升级文件失败: %w", closeErr)
	}
	if written > maxUpdateSize {
		return "", errors.New("升级文件超过允许的大小")
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (c *Checker) fetch(ctx context.Context, rawURL string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "CatScope/"+normalizeVersion(c.Version))
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("响应内容过大")
	}
	return data, nil
}

func trustedDownloadURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "github.com" || strings.HasSuffix(host, ".github.com") || host == "githubusercontent.com" || strings.HasSuffix(host, ".githubusercontent.com")
}
