package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/brawaru/marct/network"
	"github.com/brawaru/marct/validfile"
)

// linux (x64) => linux
// linux (x86) => linux-i386
// mac os => mac-os
// windows (x64) => windows-x64
// windows (x86) => windows-x86

// GetJRESelector returns a selector, representing operating system and architecture to select the most appropriate
// JRE version in JavaRuntimesMap. It is used internally by GetMatching function.
func GetJRESelector() (selector string) {
	selector = runtime.GOOS
	arch := runtime.GOARCH

	switch selector {
	case "darwin":
		if arch == "amd64" {
			selector = "mac-os"
		} else {
			// do not download 64 JRE on M1
			selector = "mac-os-" + arch
		}
	case "linux":
		switch arch {
		case "amd64":
			selector = "linux"
		case "386":
			selector = "linux-i386"
		default:
			selector = "linux-" + arch
		}
	case "windows":
		switch arch {
		case "386":
			selector += "-x86"
		case "amd64":
			// selector = "linux"
			selector += "-x64"
		default:
			selector += "-" + arch
		}
	default:
		if arch == "386" {
			selector += "-" + "i386"
		} else {
			selector += "-" + arch
		}
	}

	return
}

// ReadJREs tries to read existing Java Runtimes manifest file
func (w *Instance) ReadJREs() (runtimes *JavaRuntimesMap, err error) {
	var manifestBytes []byte

	manifestBytes, err = os.ReadFile(filepath.Join(w.jreRuntimesPath(), javaRuntimesManifestName))

	if err == nil {
		err = json.Unmarshal(manifestBytes, &runtimes)
	}

	return
}

// FetchJREs checks whether existing Java Runtimes manifest file is not too old, then, if it is old, fetches anew, or
// otherwise, re-uses existing file, unless force argument is set to true.
func (w *Instance) FetchJREs(force bool) (runtimes *JavaRuntimesMap, err error) {
	name := filepath.Join(w.jreRuntimesPath(), javaRuntimesManifestName)

	expired := force || validfile.NotExpired(name, javaRuntimesManifestTTL) != nil

	if expired {
		_, err = network.Download(javaRuntimesURL, name)
	}

	if err == nil {
		return w.ReadJREs()
	}

	return
}

func (w *Instance) jreRuntimesPath() string {
	return filepath.Join(w.Path, filepath.FromSlash(runtimesPath))
}

func (w *Instance) JREPath(version string, selector string) string {
	return filepath.Join(w.jreRuntimesPath(), version, selector)
}

func (w *Instance) InstallJRE(runtimes JavaRuntimesMap, version string) error {
	matching, selector := runtimes.GetMatching()

	if matching == nil {
		return &JavaUnavailableError{
			System:  selector,
			Version: version,
			Errno:   ErrSystemUnsupported,
		}
	}

	desc := matching[version].MostRecent()

	if desc == nil {
		return &JavaUnavailableError{
			System:  selector,
			Version: version,
			Errno:   ErrVersionUnavailable,
		}
	}

	installation := NewInstallation(version, selector, desc, w.JREPath(version, selector))

	if installErr := installation.Install(); installErr != nil {
		return fmt.Errorf("cannot install JRE %s (%s): %w", version, selector, installErr)
	}

	return nil
}
