package which

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrNotExecutable = errors.New("not executable")

// Executable returns a full path to first executable program in $PATH.
func Executable(fSys fs.FS, program string) (string, error) {
	for _, path := range getPaths() {
		filename := filepath.Join(path, program)
		executable, err := isExecutable(fSys, filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("check executable for file (%s): %w", filename, err)
		}

		if executable {
			return filename, nil
		}
	}

	return "", ErrNotExecutable
}

func isExecAny(mode os.FileMode) bool {
	return mode&0x0111 != 0
}

func isExecutable(fSys fs.FS, path string) (bool, error) {
	info, err := fs.Stat(fSys, path[1:])
	if err != nil {
		return false, fmt.Errorf("get file stat: %w", err)
	}

	if info.IsDir() {
		return false, nil
	}

	if isExecAny(info.Mode()) {
		return true, nil
	}

	return runtime.GOOS == "windows", nil
}

func getPaths() []string {
	return strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
}
