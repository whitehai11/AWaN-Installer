# AWaN Installer

AWaN Installer is a single Go binary that supports both terminal and graphical installation flows and automatically selects the correct interface for the current environment.

Supported installer modes:

- Terminal installer
- Graphical installer

Supported install targets:

- AWaN GUI
- AWaN TUI

## Structure

```text
cmd/
  main.go

installer/
  install.go
  download.go
  detect_os.go
  environment_detect.go
  paths.go

ui/
  tui/
    tui.go
  gui/
    gui.go

utils/
  logger.go
```

## Install Flow

1. detect operating system
2. create installation directory
3. download AWaN Core
4. download selected interface
5. create `~/.awan` directory structure
6. generate default configuration
7. create CLI launcher in `~/.awan/bin`
8. add the AWaN bin directory to the user PATH
9. verify `awan --version`

## Install Location

Linux:

```text
~/.awan
```

Windows:

```text
%USERPROFILE%\.awan
```

Directory structure:

```text
.awan/
  bin/
  core/
  gui/
  tui/
  memory/
  agents/
  files/
  tools/
```

## Release Sources

The installer downloads release artifacts from:

- `github.com/whitehai11/AWaN/releases`
- `github.com/whitehai11/AWaN-GUI/releases`
- `github.com/whitehai11/AWaN-TUI/releases`

## Build

Windows:

```bash
GOOS=windows GOARCH=amd64 go build -o awan-installer.exe ./cmd
```

Linux:

```bash
GOOS=linux GOARCH=amd64 go build -o awan-installer ./cmd
```

## Run

```bash
go run ./cmd
```

## CLI Launcher

After installation, the installer creates an `awan` launcher in:

Linux:

```text
~/.awan/bin/awan
```

Windows:

```text
%USERPROFILE%\.awan\bin\awan.bat
```

The installer also updates the user PATH automatically so `awan` can be run without manual shell setup.

## Environment Detection

On startup the installer automatically detects the current environment:

- `DESKTOP`: starts the graphical installer
- `SERVER`: starts the terminal installer
- `UNKNOWN`: defaults to the terminal installer

Server detection signals:

- `SSH_CONNECTION`
- `SSH_CLIENT`
- missing `DISPLAY` on Linux
- missing `WAYLAND_DISPLAY`

Desktop detection signals:

- `DISPLAY` on Linux
- `WAYLAND_DISPLAY`
- Windows desktop session

If GUI startup fails on a detected desktop system, the installer falls back to the TUI automatically.
