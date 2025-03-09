package utils

import (
	"os"
	"path/filepath"

	"github.com/bestruirui/bestsub/utils/log"
)

func GetExecutablePath() string {
	ex, err := os.Executable()
	if err != nil {
		log.Error("get executable path failed: %v", err)
		return "."
	}
	return filepath.Dir(ex)
}
