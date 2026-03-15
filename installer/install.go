package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/whitehai11/AWaN-Installer/utils"
)

const (
	TargetGUI = "gui"
	TargetTUI = "tui"
)

// Step is a single installer step.
type Step struct {
	Title string
}

// Result summarizes an installation.
type Result struct {
	RootPath   string `json:"rootPath"`
	Target     string `json:"target"`
	Platform   string `json:"platform"`
	FinishedAt string `json:"finishedAt"`
	Launcher   string `json:"launcher"`
	Verified   bool   `json:"verified"`
}

// Installer coordinates detection, downloads, directory creation, and config generation.
type Installer struct {
	paths      *Paths
	platform   Platform
	downloader *Downloader
	logger     *utils.Logger
}

// New creates a shared installer engine.
func New(logger *utils.Logger) (*Installer, error) {
	paths, err := NewPaths()
	if err != nil {
		return nil, err
	}

	return &Installer{
		paths:      paths,
		platform:   DetectOS(),
		downloader: NewDownloader(logger),
		logger:     logger,
	}, nil
}

// Paths returns the resolved install layout.
func (i *Installer) Paths() *Paths {
	return i.paths
}

// Platform returns the detected OS details.
func (i *Installer) Platform() Platform {
	return i.platform
}

// Steps returns the shared install flow.
func (i *Installer) Steps(target string) []Step {
	return []Step{
		{Title: "Detect operating system"},
		{Title: "Create installation directory"},
		{Title: "Download AWaN Core"},
		{Title: "Download selected interface"},
		{Title: "Create ~/.awan directory structure"},
		{Title: "Generate default configuration"},
		{Title: "Create CLI launcher"},
		{Title: "Update user PATH"},
		{Title: "Verify installation"},
	}
}

// Run performs the selected installation flow.
func (i *Installer) Run(target string, progress func(step Step, index, total int)) (*Result, error) {
	if target != TargetGUI && target != TargetTUI {
		return nil, fmt.Errorf("unknown installation target %q", target)
	}

	var launcherPath string
	verified := false

	steps := i.Steps(target)
	for index, step := range steps {
		progress(step, index, len(steps))

		switch step.Title {
		case "Detect operating system":
			i.logger.Log("INSTALL", "Detected platform "+i.platform.OS+"/"+i.platform.Architecture)
		case "Create installation directory":
			if err := os.MkdirAll(i.paths.Root, 0o700); err != nil {
				return nil, err
			}
		case "Download AWaN Core":
			archivePath, err := i.downloader.Download(componentCore, i.platform, i.paths.Core)
			if err != nil {
				return nil, err
			}
			if err := i.extractArchive(archivePath, i.paths.Core); err != nil {
				return nil, err
			}
		case "Download selected interface":
			component := componentGUI
			destination := i.paths.GUI
			if target == TargetTUI {
				component = componentTUI
				destination = i.paths.TUI
			}

			archivePath, err := i.downloader.Download(component, i.platform, destination)
			if err != nil {
				return nil, err
			}
			if err := i.extractArchive(archivePath, destination); err != nil {
				return nil, err
			}
		case "Create ~/.awan directory structure":
			if err := i.paths.Ensure(); err != nil {
				return nil, err
			}
		case "Generate default configuration":
			if err := i.writeDefaultConfig(target); err != nil {
				return nil, err
			}
		case "Create CLI launcher":
			var err error
			launcherPath, err = i.createLauncher()
			if err != nil {
				return nil, err
			}
		case "Update user PATH":
			if err := i.ensureUserPath(); err != nil {
				return nil, err
			}
		case "Verify installation":
			var err error
			verified, err = i.verifyLauncher()
			if err != nil {
				return nil, err
			}
		}
	}

	return &Result{
		RootPath:   i.paths.Root,
		Target:     target,
		Platform:   i.platform.OS + "/" + i.platform.Architecture,
		FinishedAt: time.Now().UTC().Format(time.RFC3339),
		Launcher:   launcherPath,
		Verified:   verified,
	}, nil
}

func (i *Installer) writeDefaultConfig(target string) error {
	config := map[string]any{
		"defaultModel": "openai",
		"defaultAgent": "default",
		"interface":    target,
		"api": map[string]any{
			"host": "localhost",
			"port": 7452,
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(i.paths.Root, "awan.config.json"), data, 0o600)
}

func (i *Installer) createLauncher() (string, error) {
	coreBinary, err := i.findCoreBinary()
	if err != nil {
		return "", err
	}

	if i.platform.OS == "windows" {
		launcherPath := filepath.Join(i.paths.Bin, "awan.bat")
		content := "@echo off\r\n\"" + coreBinary + "\" %*\r\n"
		return launcherPath, os.WriteFile(launcherPath, []byte(content), 0o600)
	}

	launcherPath := filepath.Join(i.paths.Bin, "awan")
	content := "#!/usr/bin/env bash\n\"" + coreBinary + "\" \"$@\"\n"
	if err := os.WriteFile(launcherPath, []byte(content), 0o700); err != nil {
		return "", err
	}
	return launcherPath, os.Chmod(launcherPath, 0o755)
}

func (i *Installer) ensureUserPath() error {
	if i.platform.OS == "windows" {
		current := os.Getenv("PATH")
		if pathContains(current, i.paths.Bin) {
			return nil
		}

		value := current
		if strings.TrimSpace(value) == "" {
			value = i.paths.Bin
		} else {
			value = value + ";" + i.paths.Bin
		}

		cmd := exec.Command("powershell", "-NoProfile", "-Command", "[Environment]::SetEnvironmentVariable('Path', @'\n"+value+"\n'@, 'User')")
		return cmd.Run()
	}

	line := `export PATH="$HOME/.awan/bin:$PATH"`
	for _, shellFile := range []string{
		filepath.Join(i.homeDir(), ".bashrc"),
		filepath.Join(i.homeDir(), ".zshrc"),
		filepath.Join(i.homeDir(), ".profile"),
	} {
		if err := ensureLine(shellFile, line); err != nil {
			return err
		}
	}

	return nil
}

func (i *Installer) verifyLauncher() (bool, error) {
	launcher := filepath.Join(i.paths.Bin, "awan")
	if i.platform.OS == "windows" {
		launcher = filepath.Join(i.paths.Bin, "awan.bat")
	}

	command := exec.Command(launcher, "--version")
	command.Env = prependPath(os.Environ(), i.paths.Bin, i.platform.OS == "windows")
	if output, err := command.CombinedOutput(); err != nil {
		return false, fmt.Errorf("verification failed: %s", strings.TrimSpace(string(output)))
	}

	return true, nil
}

func (i *Installer) extractArchive(archivePath, destination string) error {
	switch filepath.Ext(archivePath) {
	case ".zip":
		return extractZip(archivePath, destination)
	case ".gz":
		if strings.HasSuffix(archivePath, ".tar.gz") {
			return extractTarGz(archivePath, destination)
		}
	}
	return nil
}

func extractZip(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := safeJoin(destination, file.Name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o700); err != nil {
				return err
			}
			continue
		}

		source, err := file.Open()
		if err != nil {
			return err
		}

		target, err := os.Create(targetPath)
		if err != nil {
			source.Close()
			return err
		}

		if _, err := io.Copy(target, source); err != nil {
			target.Close()
			source.Close()
			return err
		}

		target.Close()
		source.Close()
	}

	return nil
}

func extractTarGz(archivePath, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath, err := safeJoin(destination, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o700); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
				return err
			}

			target, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(target, tarReader); err != nil {
				target.Close()
				return err
			}
			target.Close()
			if runtime.GOOS != "windows" {
				if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func ensureLine(path, line string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(data), line) {
		return nil
	}

	content := strings.TrimRight(string(data), "\n")
	if content != "" {
		content += "\n"
	}
	content += line + "\n"
	return os.WriteFile(path, []byte(content), 0o600)
}

func prependPath(env []string, bin string, windows bool) []string {
	key := "PATH="
	separator := ":"
	if windows {
		separator = ";"
	}

	result := make([]string, 0, len(env)+1)
	replaced := false
	for _, item := range env {
		if strings.HasPrefix(strings.ToUpper(item), "PATH=") {
			value := item[5:]
			result = append(result, key+bin+separator+value)
			replaced = true
			continue
		}
		result = append(result, item)
	}
	if !replaced {
		result = append(result, key+bin)
	}
	return result
}

func pathContains(pathValue, candidate string) bool {
	for _, part := range strings.Split(pathValue, string(os.PathListSeparator)) {
		if strings.EqualFold(strings.TrimSpace(part), candidate) {
			return true
		}
	}
	return false
}

func (i *Installer) homeDir() string {
	if i.platform.OS == "windows" {
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			return profile
		}
	}
	home, _ := os.UserHomeDir()
	return home
}

func (i *Installer) findCoreBinary() (string, error) {
	expected := "awan" + i.platform.ExecutableSuffix
	direct := filepath.Join(i.paths.Core, expected)
	if info, err := os.Stat(direct); err == nil && !info.IsDir() {
		return direct, nil
	}

	var match string
	err := filepath.WalkDir(i.paths.Core, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(entry.Name(), expected) {
			match = path
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", err
	}
	if match == "" {
		return "", fmt.Errorf("could not locate %q inside %s", expected, i.paths.Core)
	}
	return match, nil
}

func safeJoin(root, name string) (string, error) {
	cleanRoot := filepath.Clean(root)
	target := filepath.Join(cleanRoot, filepath.Clean(name))
	relative, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", fmt.Errorf("archive entry %q escapes install directory", name)
	}
	return target, nil
}
