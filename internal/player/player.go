package player

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ChiThang-50Cent/vnstream/internal/config"
)

type Launcher struct {
	cfg config.Config
}

func NewLauncher(cfg config.Config) *Launcher {
	return &Launcher{cfg: cfg}
}

func (l *Launcher) LaunchVLC(link, movieName, streamName string) error {
	link = strings.TrimSpace(link)
	if link == "" {
		return errors.New("empty stream link")
	}

	bin, err := findVLCBinary()
	if err != nil {
		return err
	}

	title := buildTitle(movieName, streamName)
	args := []string{
		"--play-and-exit",
		"--no-fullscreen",
		"--embedded-video",
		"--autoscale",
		"--no-qt-video-autoresize",
		"--zoom=1",
		fmt.Sprintf("--width=%d", l.cfg.VLCWidth),
		fmt.Sprintf("--height=%d", l.cfg.VLCHeight),
		"--avcodec-hw=none",
	}
	if title != "" {
		args = append(args, "--meta-title="+title)
	}
	args = append(args, link)

	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+l.cfg.VLCXDGConfigHome,
		"XDG_CACHE_HOME="+l.cfg.VLCXDGCacheHome,
	)
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	if err := cmd.Start(); err != nil {
		return err
	}

	_ = cmd.Process.Release()
	return nil
}

func findVLCBinary() (string, error) {
	if p, err := exec.LookPath("qvlc"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("vlc"); err == nil {
		return p, nil
	}
	return "", errors.New("vlc not found")
}

func buildTitle(movieName, streamName string) string {
	movieName = sanitize(movieName)
	streamName = sanitize(streamName)
	if movieName != "" && streamName != "" {
		return movieName + " - " + streamName
	}
	if streamName != "" {
		return streamName
	}
	return movieName
}

func sanitize(v string) string {
	v = strings.ReplaceAll(v, "\t", " ")
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	return strings.TrimSpace(v)
}
