package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ChiThang-50Cent/vnstream/internal/api"
)

type searchLoadedMsg struct {
	Query  string
	Movies []api.Movie
	Err    error
}

type movieItem struct{ movie api.Movie }

func (m movieItem) Title() string {
	if m.movie.Year != "" {
		return fmt.Sprintf("%s %s (%s)", m.movie.Emoji, m.movie.Name, m.movie.Year)
	}
	return fmt.Sprintf("%s %s", m.movie.Emoji, m.movie.Name)
}

func (m movieItem) Description() string {
	if m.movie.Label != "" {
		return m.movie.Label
	}
	return "Movie"
}

func (m movieItem) FilterValue() string {
	return m.movie.Name + " " + m.movie.Year + " " + m.movie.Label
}

type SearchModel struct {
	apiClient *api.Client
	styles    Styles

	input textinput.Model
	list  list.Model
	spin  spinner.Model

	allMovies    []api.Movie
	currentQuery string
	loading      bool
	errMsg       string
}

func NewSearchModel(client *api.Client, styles Styles, query string) *SearchModel {
	ti := textinput.New()
	ti.Prompt = "🔍 "
	ti.Placeholder = "Type a new keyword..."
	ti.SetValue(strings.TrimSpace(query))
	ti.CharLimit = 120
	ti.Focus()

	l := list.New([]list.Item{}, NewCompactDelegate(), 0, 0)
	l.Title = "Search"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &SearchModel{
		apiClient:    client,
		styles:       styles,
		input:        ti,
		list:         l,
		spin:         sp,
		allMovies:    nil,
		currentQuery: strings.TrimSpace(query),
		loading:      strings.TrimSpace(query) != "",
	}
}

func (m *SearchModel) Init() tea.Cmd {
	if strings.TrimSpace(m.currentQuery) == "" {
		return nil
	}
	return tea.Batch(m.spin.Tick, fetchMoviesCmd(m.apiClient, m.currentQuery))
}

func (m *SearchModel) shouldUpdateList(msg tea.Msg) bool {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return true
	}
	switch km.String() {
	case "up", "down", "pgup", "pgdown", "home", "end":
		return true
	default:
		return false
	}
}

func (m *SearchModel) SetSize(width, height int) {
	listHeight := height - 7
	if listHeight < 6 {
		listHeight = 6
	}
	m.list.SetSize(width-2, listHeight)
}

func (m *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case searchLoadedMsg:
		m.loading = false
		m.errMsg = ""
		if msg.Err != nil {
			m.errMsg = "Could not load search results."
			m.allMovies = nil
			m.list.SetItems([]list.Item{})
			return m, nil
		}

		m.allMovies = msg.Movies
		items := make([]list.Item, 0, len(m.allMovies))
		for _, mv := range m.allMovies {
			items = append(items, movieItem{movie: mv})
		}
		m.list.SetItems(items)
		if len(items) > 0 {
			m.list.Select(0)
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			return m, func() tea.Msg { return backHomeMsg{} }
		case "alt+enter":
			q := strings.TrimSpace(m.input.Value())
			if q == "" {
				return m, nil
			}
			m.currentQuery = q
			m.loading = true
			m.errMsg = ""
			m.allMovies = nil
			m.list.SetItems([]list.Item{})
			return m, tea.Batch(m.spin.Tick, fetchMoviesCmd(m.apiClient, q))
		case "enter":
			if m.loading {
				return m, nil
			}

			selected, ok := m.list.SelectedItem().(movieItem)
			if ok {
				return m, func() tea.Msg {
					return openStreamMsg{Movie: selected.movie, ParentQuery: m.currentQuery}
				}
			}

			if len(m.list.Items()) == 0 {
				q := strings.TrimSpace(m.input.Value())
				if q == "" {
					return m, nil
				}
				m.currentQuery = q
				m.loading = true
				m.errMsg = ""
				m.allMovies = nil
				m.list.SetItems([]list.Item{})
				return m, tea.Batch(m.spin.Tick, fetchMoviesCmd(m.apiClient, q))
			}
		}
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	if m.shouldUpdateList(msg) {
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *SearchModel) View() string {
	header := []string{
		m.styles.Title.Render("VNStream Search"),
		m.styles.Header.Render("Query · filter live"),
	}

	body := ""
	if m.loading {
		body = m.styles.Busy.Render(m.spin.View() + " Loading results...")
	} else if m.errMsg != "" {
		body = m.styles.Error.Render("✗ " + m.errMsg)
	} else if len(m.list.Items()) == 0 {
		body = m.styles.Muted.Render("No results found. Try another keyword.")
	} else {
		body = m.list.View()
	}

	inputSection := m.input.View()
	footer := m.styles.Footer.Render("↑↓: move · Enter: select movie/search when empty · Alt+Enter: new search · ←: Home")
	return strings.Join(header, "\n") + "\n\n" + body + "\n\n" + inputSection + "\n" + footer
}

func fetchMoviesCmd(client *api.Client, query string) tea.Cmd {
	return func() tea.Msg {
		movies, err := client.SearchMovies(context.Background(), query)
		return searchLoadedMsg{Query: query, Movies: movies, Err: err}
	}
}
