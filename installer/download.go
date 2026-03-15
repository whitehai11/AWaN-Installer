package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/awan/awan-installer/utils"
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

// NewDownloader creates a downloader instance.
func NewDownloader(logger *utils.Logger) *Downloader {
	return &Downloader{
		client: &http.Client{Timeout: 60 * time.Second},
		logger: logger,
	}
}

// Download fetches a component artifact into the destination directory.
func (d *Downloader) Download(component string, platform Platform, destinationDir string) (string, error) {
	url := releaseURL(component, platform)
	if url == "" {
		return "", fmt.Errorf("unsupported component %q", component)
	}
	filename := filepath.Base(url)
	targetPath := filepath.Join(destinationDir, filename)

	d.logger.Log("INSTALL", "Downloading "+component+" from "+url)

	if err := os.MkdirAll(destinationDir, 0o700); err != nil {
		return "", err
	}

	response, err := d.client.Get(url)
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

func releaseURL(component string, platform Platform) string {
	name := strings.ToLower(component)
	switch name {
	case componentCore:
		return fmt.Sprintf("https://github.com/awan/core/releases/latest/download/awan-core-%s-%s%s", platform.OS, platform.Architecture, platform.ArchiveExtension)
	case componentGUI:
		return fmt.Sprintf("https://github.com/awan/gui/releases/latest/download/awan-gui-%s-%s%s", platform.OS, platform.Architecture, platform.ArchiveExtension)
	case componentTUI:
		return fmt.Sprintf("https://github.com/awan/tui/releases/latest/download/awan-tui-%s-%s%s", platform.OS, platform.Architecture, platform.ArchiveExtension)
	default:
		return ""
	}
}
