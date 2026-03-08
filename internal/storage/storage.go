package storage

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChiThang-50Cent/vnstream/internal/config"
)

type WatchedItem struct {
	Timestamp  string
	MovieName  string
	StreamName string
	Link       string
	MovieID    string
}

type Manager struct {
	cfg config.Config
}

func NewManager(cfg config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) EnsureFiles() error {
	if err := os.MkdirAll(m.cfg.DataDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(m.cfg.VLCXDGConfigHome, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(m.cfg.VLCXDGCacheHome, 0o755); err != nil {
		return err
	}

	if !fileExists(m.cfg.SearchHistory) && fileExists(m.cfg.LegacyHistory) {
		if data, err := os.ReadFile(m.cfg.LegacyHistory); err == nil {
			_ = os.WriteFile(m.cfg.SearchHistory, data, 0o644)
		}
	}

	if err := touchFile(m.cfg.SearchHistory); err != nil {
		return err
	}
	if err := touchFile(m.cfg.WatchedHistory); err != nil {
		return err
	}

	return nil
}

func (m *Manager) LoadSearchHistory(limit int) ([]string, error) {
	if err := m.EnsureFiles(); err != nil {
		return nil, err
	}

	lines, err := readLines(m.cfg.SearchHistory)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
		if limit > 0 && len(out) >= limit {
			break
		}
	}

	return out, nil
}

func (m *Manager) SaveHistory(query string) error {
	if err := m.EnsureFiles(); err != nil {
		return err
	}

	query = sanitizeField(query)
	if query == "" {
		return nil
	}

	lines, err := readLines(m.cfg.SearchHistory)
	if err != nil {
		return err
	}

	out := []string{query}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == query {
			continue
		}
		out = append(out, line)
		if len(out) >= 20 {
			break
		}
	}

	return writeLinesAtomic(m.cfg.SearchHistory, out)
}

func (m *Manager) ClearSearchHistory() error {
	if err := m.EnsureFiles(); err != nil {
		return err
	}
	return os.WriteFile(m.cfg.SearchHistory, []byte{}, 0o644)
}

func (m *Manager) LoadWatchedHistory(limit int) ([]WatchedItem, error) {
	if err := m.EnsureFiles(); err != nil {
		return nil, err
	}

	lines, err := readLines(m.cfg.WatchedHistory)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	out := make([]WatchedItem, 0, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}
		item := WatchedItem{
			Timestamp:  parts[0],
			MovieName:  parts[1],
			StreamName: parts[2],
			Link:       parts[3],
		}
		if len(parts) > 4 {
			item.MovieID = parts[4]
		}
		out = append(out, item)
	}

	return out, nil
}

func (m *Manager) SaveWatched(movieName, streamName, link, movieID string) error {
	if err := m.EnsureFiles(); err != nil {
		return err
	}

	movieName = sanitizeField(movieName)
	streamName = sanitizeField(streamName)
	link = sanitizeField(link)
	movieID = sanitizeField(movieID)

	if link == "" {
		return errors.New("missing stream link")
	}

	lines, err := readLines(m.cfg.WatchedHistory)
	if err != nil {
		return err
	}

	filtered := make([]string, 0, len(lines)+1)
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 && parts[3] == link {
			continue
		}
		if strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}

	ts := time.Now().Format("2006-01-02 15:04:05")
	filtered = append(filtered, strings.Join([]string{ts, movieName, streamName, link, movieID}, "\t"))
	if len(filtered) > 200 {
		filtered = filtered[len(filtered)-200:]
	}

	return writeLinesAtomic(m.cfg.WatchedHistory, filtered)
}

func (m *Manager) ClearWatchedHistory() error {
	if err := m.EnsureFiles(); err != nil {
		return err
	}
	return os.WriteFile(m.cfg.WatchedHistory, []byte{}, 0o644)
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer f.Close()

	lines := make([]string, 0, 64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func writeLinesAtomic(path string, lines []string) error {
	tmp := path + ".tmp"
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func touchFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

func sanitizeField(v string) string {
	v = strings.ReplaceAll(v, "\t", " ")
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	return strings.TrimSpace(v)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
