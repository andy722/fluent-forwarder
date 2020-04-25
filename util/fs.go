package util

import (
	"fmt"
	"os"
)

func EnsureDirExists(path string) error {
	stat, err := os.Stat(path)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed creating directory %v: %w", path, err)

		} else {
			return nil
		}

	} else if err != nil {
		return fmt.Errorf("stat failed for %v: %w", path, err)

	} else if !stat.IsDir() {
		return fmt.Errorf("%v should be directory", path)
	}

	return nil
}
