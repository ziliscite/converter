package domain

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Converter struct {
	ffp string
}

func NewConverter(ffmpegPath string) *Converter {
	return &Converter{ffp: ffmpegPath}
}

func (c *Converter) ConvertMP4ToMP3(filename string, video io.Reader) (string, error) {
	// create a temporary file to store the video (io reader) so that ffmpeg can read it
	inputFile, err := os.CreateTemp("", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer inputFile.Close()

	body, err := io.ReadAll(video)
	if err != nil {
		return "", fmt.Errorf("failed to read video file: %v", err)
	}

	// write the video file to the temporary file
	if _, err = inputFile.Write(body); err != nil {
		return "", fmt.Errorf("failed to write to temporary file: %v", err)
	}

	ifn := inputFile.Name()

	// build ffmpeg base command
	args := []string{
		"-i", ifn,
		"-vn",
		"-y",
		"-ab", "192000",
	}

	// build an output file path
	outputFile := strings.TrimSuffix(ifn, filepath.Ext(ifn))

	switch c.extractCodec(ifn) {
	case "aac":
		args = append(args, "-acodec", "copy", "-f", "adts")
		outputFile += ".aac"
	case "mp3":
		args = append(args, "-acodec", "libmp3lame", "-q:a", "2", "-f", "mp3")
		outputFile += ".mp3"
	case "wav":
		args = append(args, "-acodec", "pcm_s16le", "-f", "wav")
		outputFile += ".wav"
	default:
		return "", fmt.Errorf("unsupported audio codec")
	}

	cmd := exec.Command(c.ffp, append(args, outputFile)...)

	// log errors
	cmd.Stderr = os.Stderr

	// run the ffmpeg command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run ffmpeg: %v", err)
	}

	return outputFile, nil
}

func (c *Converter) extractCodec(path string) string {
	probeCmd := exec.Command(c.ffp, "-i", path)
	probeOutput, _ := probeCmd.CombinedOutput()

	// Extract audio codec from probe output
	codecLine := strings.Split(string(probeOutput), "\n")
	var audioCodec string
	for _, line := range codecLine {
		if strings.Contains(line, "Audio:") {
			parts := strings.Split(line, "Audio: ")
			if len(parts) > 1 {
				audioCodec = strings.Split(parts[1], " ")[0]
			}
			break
		}
	}

	return audioCodec
}
