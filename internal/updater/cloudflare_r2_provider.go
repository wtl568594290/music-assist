package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	wailsupdater "github.com/wailsapp/wails/v3/pkg/updater"
)

type CloudflareR2Provider struct {
	Repo          string
	AssetsMatcher cloudflareR2AssetsMatcher
	HTTPClient    *http.Client
	BaseUrl       string
}
type cloudflareR2AssetsMatcher func(req wailsupdater.CheckRequest, assets []CloudflareR2Asset) int

type CloudflareR2Asset struct {
	Platform    string `json:"platform"`
	Name        string `json:"name"`
	DownloadUrl string `json:"download_url"`
	Size        int64  `json:"size"`
}
type CloudflareR2Release struct {
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Body        string              `json:"body"`
	PublishedAt string              `json:"published_at"`
	Assets      []CloudflareR2Asset `json:"assets"`
}

func defaultCloudflareR2AssetMatcher(req wailsupdater.CheckRequest, assets []CloudflareR2Asset) int {
	// First pass: skip obvious sidecars.
	for i, a := range assets {
		name := strings.ToLower(a.Name)
		if a.Platform == "windows" && strings.Contains(name, ".exe") {
			return i
		}
	}
	return -1
}

func (crp *CloudflareR2Provider) Name() string {
	return "cloudflare-r2"
}

func (crp *CloudflareR2Provider) client() *http.Client {
	if crp.HTTPClient == nil {
		crp.HTTPClient = http.DefaultClient
	}
	return crp.HTTPClient
}

func (p *CloudflareR2Provider) apiUrl(path string) string {
	baseUrl := ""
	if p.BaseUrl != "" {
		baseUrl = p.BaseUrl
	}
	return fmt.Sprintf(
		"%s/%s/%s",
		baseUrl, p.Repo, path,
	)
}

func (crp *CloudflareR2Provider) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	resp, err := crp.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"Cloudflare R2 API returned %s: %s",
			resp.Status, strings.TrimSpace(string(body)),
		)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (crp *CloudflareR2Provider) Check(
	ctx context.Context,
	req wailsupdater.CheckRequest,
) (*wailsupdater.Release, error) {
	if crp.BaseUrl == "" {
		return nil, fmt.Errorf("baseUrl is empty")
	}

	if req.Platform != "windows" {
		return nil, fmt.Errorf(
			"unsupported platform: %s", req.Platform,
		)
	}

	var rel CloudflareR2Release
	if err := crp.getJSON(
		ctx,
		crp.apiUrl("info.json"),
		&rel,
	); err != nil {
		return nil, err
	}

	version := strings.TrimPrefix(rel.Version, "v")
	if version == "" {
		return nil, fmt.Errorf("release version is empty")
	}

	newer, err := isNewerVersion(version, req.CurrentVersion)
	if err != nil {
		return nil, err
	}
	if !newer {
		return nil, nil
	}

	assets := rel.Assets
	// Require a predictable asset name, e.g. mo-out-windows-amd64.exe.
	if crp.AssetsMatcher == nil {
		crp.AssetsMatcher = defaultCloudflareR2AssetMatcher
	}
	idx := crp.AssetsMatcher(req, assets)
	if idx > len(assets)-1 {
		return nil, errors.New("No asset in Cloudflare R2 release")
	}

	var selected *CloudflareR2Asset = &assets[idx]
	publishedAt, _ := time.Parse(time.RFC3339, rel.PublishedAt)

	return &wailsupdater.Release{
		Version:     version,
		Name:        rel.Name,
		Notes:       rel.Body,
		PublishedAt: publishedAt,
		Artifact: wailsupdater.Artifact{
			Filename: selected.Name,
			Filetype: "exe",
			Size:     selected.Size,
			Platform: req.Platform,
			Arch:     req.Arch,
		},
		Metadata: map[string]any{
			"download_url": selected.DownloadUrl,
		},
	}, nil
}

func (crp *CloudflareR2Provider) Download(
	ctx context.Context,
	rel *wailsupdater.Release,
	dst io.Writer,
	onProgress func(written, total int64),
) error {
	if crp.BaseUrl == "" {
		return fmt.Errorf("baseUrl is empty")
	}

	downloadUrl := rel.Metadata["download_url"]
	if downloadUrl == nil {
		return fmt.Errorf("download_url is empty")
	}
	url, ok := downloadUrl.(string)
	if !ok {
		return fmt.Errorf("download_url is not a string")
	}

	url = crp.apiUrl(url)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)
	if err != nil {
		return err
	}

	resp, err := crp.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"Cloudflare R2 download returned %s: %s",
			resp.Status, strings.TrimSpace(string(body)),
		)
	}

	total := resp.ContentLength
	if total <= 0 {
		total = rel.Artifact.Size
	}

	buf := make([]byte, 64*1024)
	var written int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			wn, writeErr := dst.Write(buf[:n])
			written += int64(wn)

			if writeErr != nil {
				return writeErr
			}
			if wn != n {
				return io.ErrShortWrite
			}

			onProgress(written, total)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}
