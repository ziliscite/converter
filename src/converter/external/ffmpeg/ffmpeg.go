package ffmpeg

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed ffmpeg
var binary []byte

func Open() (string, error) {
	binDir := "/usr/local/bin"
	outputPath := filepath.Join(binDir, "ffmpeg")

	// check if the file already exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// create a directory if it doesn't exist
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return "", err
		}

		// write embedded binary to the file
		if err = os.WriteFile(outputPath, binary, 0755); err != nil {
			return "", err
		}
	}

	return outputPath, nil
}
