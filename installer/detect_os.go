package installer

import "runtime"

// Platform describes the current target OS.
type Platform struct {
	OS                string
	Architecture      string
	ExecutableSuffix  string
	ArchiveExtension  string
}

// DetectOS resolves the current runtime platform.
func DetectOS() Platform {
	platform := Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	switch runtime.GOOS {
	case "windows":
		platform.ExecutableSuffix = ".exe"
		platform.ArchiveExtension = ".zip"
	default:
		platform.ExecutableSuffix = ""
		platform.ArchiveExtension = ".tar.gz"
	}

	return platform
}
