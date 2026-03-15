package installer

import (
	"os"
	"runtime"
	"strings"
)

const (
	EnvironmentDesktop = "DESKTOP"
	EnvironmentServer  = "SERVER"
	EnvironmentUnknown = "UNKNOWN"
)

// DetectEnvironment determines whether the installer is running in a desktop or server context.
func DetectEnvironment() string {
	if hasEnv("SSH_CONNECTION") || hasEnv("SSH_CLIENT") {
		return EnvironmentServer
	}

	switch runtime.GOOS {
	case "windows":
		// A normal Windows user session is treated as a desktop environment.
		return EnvironmentDesktop
	default:
		display := os.Getenv("DISPLAY")
		wayland := os.Getenv("WAYLAND_DISPLAY")

		if display != "" || wayland != "" {
			return EnvironmentDesktop
		}

		if runtime.GOOS == "linux" {
			return EnvironmentServer
		}
	}

	return EnvironmentUnknown
}

func hasEnv(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}
