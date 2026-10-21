package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	wailsupdater "github.com/wailsapp/wails/v3/pkg/updater"
)

type GiteeProvider struct {
	Owner         string
	Repo          string
	HTTPClient    *http.Client
	AssetsMatcher GiteeAssetsMatcher
	AccessToken   string
}

type GiteeAssetsMatcher func(req wailsupdater.CheckRequest, assets []GiteeAsset) int

type giteeRelease struct {
	ID        int64  `json:"id"`
	TagName   string `json:"tag_name"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type GiteeAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func defaultAssetMatcher(req wailsupdater.CheckRequest, assets []GiteeAsset) int {
	// First pass: skip obvious sidecars.
	for i, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, ".exe") {
			return i
		}
	}
	return -1
}

func (p *GiteeProvider) Name() string {
	return "gitee"
}

func (p *GiteeProvider) apiURL(path string) string {
	pStr := ""
	if p.AccessToken != "" {
		pStr = fmt.Sprintf("?access_token=%s", p.AccessToken)
	}
	return fmt.Sprintf(
		"https://gitee.com/api/v5/repos/%s/%s/%s%s",
		p.Owner, p.Repo, path, pStr,
	)
}

func (p *GiteeProvider) client() *http.Client {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return http.DefaultClient
}

func (p *GiteeProvider) getJSON(
	ctx context.Context,
	url string,
	dst any,
) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"Gitee API returned %s: %s",
			resp.Status, strings.TrimSpace(string(body)),
		)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (p *GiteeProvider) Check(
	ctx context.Context,
	req wailsupdater.CheckRequest,
) (*wailsupdater.Release, error) {
	if req.Platform != "windows" {
		return nil, fmt.Errorf(
			"unsupported platform: %s", req.Platform,
		)
	}

	var rel giteeRelease
	if err := p.getJSON(
		ctx,
		p.apiURL("releases/latest"),
		&rel,
	); err != nil {
		return nil, err
	}

	version := strings.TrimPrefix(rel.TagName, "v")
	if version == "" {
		return nil, fmt.Errorf("release tag_name is empty")
	}

	newer, err := isNewerVersion(version, req.CurrentVersion)
	if err != nil {
		return nil, err
	}
	if !newer {
		return nil, nil
	}

	var assets []GiteeAsset
	if err := p.getJSON(
		ctx,
		p.apiURL(fmt.Sprintf(
			"releases/%d/attach_files", rel.ID,
		)),
		&assets,
	); err != nil {
		return nil, err
	}

	// Require a predictable asset name, e.g. mo-out-windows-amd64.exe.
	if p.AssetsMatcher == nil {
		p.AssetsMatcher = defaultAssetMatcher
	}
	idx := p.AssetsMatcher(req, assets)
	if idx > len(assets)-1 {
		return nil, errors.New("No asset in Gitee release")
	}

	var selected *GiteeAsset = &assets[idx]
	publishedAt, _ := time.Parse(time.RFC3339, rel.CreatedAt)

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
			"release_id": rel.ID,
			"asset_id":   selected.ID,
		},
	}, nil
}

func (p *GiteeProvider) Download(
	ctx context.Context,
	rel *wailsupdater.Release,
	dst io.Writer,
	onProgress func(written, total int64),
) error {
	releaseID, err := metadataInt64(rel, "release_id")
	if err != nil {
		return err
	}
	assetID, err := metadataInt64(rel, "asset_id")
	if err != nil {
		return err
	}

	url := p.apiURL(fmt.Sprintf(
		"releases/%d/attach_files/%d/download",
		releaseID, assetID,
	))

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)
	if err != nil {
		return err
	}

	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"Gitee download returned %s: %s",
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

func metadataInt64(
	rel *wailsupdater.Release,
	key string,
) (int64, error) {
	value, ok := rel.Metadata[key]
	if !ok {
		return 0, fmt.Errorf("missing release metadata: %s", key)
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf(
			"invalid metadata %s: %T", key, value,
		)
	}
}

// Compares numeric versions such as 1.2.3 and 1.10.0.
// This intentionally does not implement prerelease semver ordering.
func isNewerVersion(latest, current string) (bool, error) {
	parse := func(s string) ([]uint64, error) {
		s = strings.TrimPrefix(s, "v")
		parts := strings.Split(s, ".")
		if len(parts) == 0 {
			return nil, fmt.Errorf("invalid version %q", s)
		}

		values := make([]uint64, len(parts))
		for i, part := range parts {
			if part == "" {
				return nil, fmt.Errorf("invalid version %q", s)
			}
			v, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid version %q", s)
			}
			values[i] = v
		}
		return values, nil
	}

	a, err := parse(latest)
	if err != nil {
		return false, err
	}
	b, err := parse(current)
	if err != nil {
		return false, err
	}

	n := len(a)
	if len(b) > n {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		var av, bv uint64
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if av != bv {
			return av > bv, nil
		}
	}

	return false, nil
}
