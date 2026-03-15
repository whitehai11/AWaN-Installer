package installer

import (
	"os"
	"path/filepath"
	"runtime"
)

// Paths describes the AWaN installation layout.
type Paths struct {
	Root   string
	Bin    string
	Core   string
	GUI    string
	TUI    string
	Memory string
	Agents string
	Files  string
	Tools  string
}

// NewPaths resolves the install root for the current OS.
func NewPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	root := filepath.Join(home, ".awan")
	if runtime.GOOS == "windows" {
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			root = filepath.Join(profile, ".awan")
		}
	}

	return &Paths{
		Root:   root,
		Bin:    filepath.Join(root, "bin"),
		Core:   filepath.Join(root, "core"),
		GUI:    filepath.Join(root, "gui"),
		TUI:    filepath.Join(root, "tui"),
		Memory: filepath.Join(root, "memory"),
		Agents: filepath.Join(root, "agents"),
		Files:  filepath.Join(root, "files"),
		Tools:  filepath.Join(root, "tools"),
	}, nil
}

// Ensure creates the AWaN directory structure.
func (p *Paths) Ensure() error {
	for _, dir := range []string{p.Root, p.Bin, p.Core, p.GUI, p.TUI, p.Memory, p.Agents, p.Files, p.Tools} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}

	return nil
}
