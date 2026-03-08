package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ChiThang-50Cent/vnstream/internal/config"
)

type Movie struct {
	ID    string
	Name  string
	Year  string
	Label string
	Emoji string
}

type Stream struct {
	URL         string
	Name        string
	Description string
}

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (c *Client) SearchMovies(ctx context.Context, query string) ([]Movie, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Movie{}, nil
	}

	encoded := url.QueryEscape(query)
	seen := make(map[string]struct{})
	results := make([]Movie, 0, 30)

	for i, catalogID := range c.cfg.CatalogIDs {
		label := ""
		emoji := "🎬"
		if i < len(c.cfg.CatalogLabels) {
			label = c.cfg.CatalogLabels[i]
		}
		if i < len(c.cfg.CatalogEmojis) {
			emoji = c.cfg.CatalogEmojis[i]
		}

		endpoint := fmt.Sprintf("%s/catalog/movie/%s/search=%s.json", c.cfg.BaseURL, catalogID, encoded)

		var payload struct {
			Metas []struct {
				ID          string      `json:"id"`
				Name        string      `json:"name"`
				Year        interface{} `json:"year"`
				ReleaseInfo interface{} `json:"releaseInfo"`
			} `json:"metas"`
		}

		if err := c.getJSON(ctx, endpoint, &payload); err != nil {
			continue
		}

		for _, m := range payload.Metas {
			if m.ID == "" {
				continue
			}
			if _, ok := seen[m.ID]; ok {
				continue
			}
			seen[m.ID] = struct{}{}

			year := normalizeYear(m.Year)
			if year == "" {
				year = normalizeYear(m.ReleaseInfo)
			}

			results = append(results, Movie{
				ID:    m.ID,
				Name:  m.Name,
				Year:  year,
				Label: label,
				Emoji: emoji,
			})
		}
	}

	return results, nil
}

func (c *Client) FetchStreams(ctx context.Context, movieID string) ([]Stream, error) {
	movieID = strings.TrimSpace(movieID)
	if movieID == "" {
		return []Stream{}, nil
	}

	endpoint := fmt.Sprintf("%s/stream/movie/%s.json", c.cfg.BaseURL, movieID)

	var payload struct {
		Streams []struct {
			URL         string `json:"url"`
			InfoHash    string `json:"infoHash"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"streams"`
	}

	if err := c.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}

	results := make([]Stream, 0, len(payload.Streams))
	for _, s := range payload.Streams {
		finalURL := strings.TrimSpace(s.URL)
		if finalURL == "" && strings.TrimSpace(s.InfoHash) != "" {
			finalURL = "magnet:?xt=urn:btih:" + strings.TrimSpace(s.InfoHash)
		}
		results = append(results, Stream{
			URL:         finalURL,
			Name:        strings.TrimSpace(s.Name),
			Description: strings.TrimSpace(s.Description),
		})
	}

	return results, nil
}

func (c *Client) ResolveMovieID(ctx context.Context, movieName string) (string, error) {
	wanted := strings.ToLower(strings.TrimSpace(movieName))
	if wanted == "" {
		return "", nil
	}

	movies, err := c.SearchMovies(ctx, movieName)
	if err != nil {
		return "", err
	}
	if len(movies) == 0 {
		return "", nil
	}

	fallback := movies[0].ID
	for _, m := range movies {
		if strings.ToLower(strings.TrimSpace(m.Name)) == wanted {
			return m.ID, nil
		}
	}

	return fallback, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}

	return nil
}

func normalizeYear(v interface{}) string {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" || s == "<nil>" {
		return ""
	}
	return s
}
