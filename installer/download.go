package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/whitehai11/AWaN-Installer/utils"
)

const (
	componentCore = "core"
	componentGUI  = "gui"
	componentTUI  = "tui"
)

// Downloader fetches release artifacts from GitHub.
type Downloader struct {
	client *http.Client
	logger *utils.Logger
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type release struct {
	TagName    string         `json:"tag_name"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}

// NewDownloader creates a downloader instance.
func NewDownloader(logger *utils.Logger) *Downloader {
	return &Downloader{
		client: &http.Client{Timeout: 60 * time.Second},
		logger: logger,
	}
}

// Download fetches a component artifact into the destination directory.
func (d *Downloader) Download(component string, platform Platform, destinationDir string) (string, error) {
	asset, rel, err := d.resolveAsset(component, platform)
	if err != nil {
		return "", err
	}
	if asset == nil {
		return "", fmt.Errorf("unsupported component %q", component)
	}
	filename := asset.Name
	targetPath := filepath.Join(destinationDir, filename)

	d.logger.Log("INSTALL", "Downloading "+component+" "+rel.TagName+" from "+asset.URL)

	if err := os.MkdirAll(destinationDir, 0o700); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "awan-installer")

	response, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return "", fmt.Errorf("download failed with status %s", response.Status)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		return "", err
	}

	return targetPath, nil
}

func (d *Downloader) resolveAsset(component string, platform Platform) (*releaseAsset, *release, error) {
	repo := repoForComponent(component)
	if repo == "" {
		return nil, nil, nil
	}

	releases, err := d.fetchReleases(repo)
	if err != nil {
		return nil, nil, err
	}

	for _, rel := range releases {
		if asset := matchingAsset(rel, component, platform); asset != nil {
			copyRelease := rel
			return asset, &copyRelease, nil
		}
	}

	return nil, nil, fmt.Errorf("no matching release asset found for %s on %s/%s", component, platform.OS, platform.Architecture)
}

func (d *Downloader) fetchReleases(repo string) ([]release, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/"+repo+"/releases?per_page=20", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "awan-installer")

	response, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("release lookup failed with status %s", response.Status)
	}

	var releases []release
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func repoForComponent(component string) string {
	switch strings.ToLower(component) {
	case componentCore:
		return "whitehai11/AWaN"
	case componentGUI:
		return "whitehai11/AWaN-GUI"
	case componentTUI:
		return "whitehai11/AWaN-TUI"
	default:
		return ""
	}
}

func matchingAsset(rel release, component string, platform Platform) *releaseAsset {
	baseName := componentBaseName(component)
	if baseName == "" {
		return nil
	}

	osName := strings.ToLower(platform.OS)
	arch := strings.ToLower(platform.Architecture)
	archiveExt := strings.ToLower(platform.ArchiveExtension)
	exeExt := strings.ToLower(platform.ExecutableSuffix)

	candidates := []string{
		baseName + "-" + osName + "-" + arch + archiveExt,
		baseName + "-" + osName + "-" + arch + exeExt,
		baseName + exeExt,
		strings.ReplaceAll(baseName, "awan", "AWaN") + exeExt,
		strings.ToUpper(baseName) + exeExt,
	}

	for _, candidate := range candidates {
		for _, asset := range rel.Assets {
			if strings.EqualFold(asset.Name, candidate) {
				copyAsset := asset
				return &copyAsset
			}
		}
	}

	for _, asset := range rel.Assets {
		name := strings.ToLower(asset.Name)
		if !strings.Contains(name, strings.ToLower(baseName)) {
			continue
		}
		if strings.Contains(name, osName) && strings.Contains(name, arch) {
			copyAsset := asset
			return &copyAsset
		}
		if runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe") {
			copyAsset := asset
			return &copyAsset
		}
		if runtime.GOOS != "windows" && (strings.HasSuffix(name, ".tar.gz") || !strings.Contains(name, ".")) {
			copyAsset := asset
			return &copyAsset
		}
	}

	return nil
}

func componentBaseName(component string) string {
	switch strings.ToLower(component) {
	case componentCore:
		return "awan-core"
	case componentGUI:
		return "awan-gui"
	case componentTUI:
		return "awan-tui"
	default:
		return ""
	}
}
