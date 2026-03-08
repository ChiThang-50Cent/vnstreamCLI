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

type streamLoadedMsg struct {
	Streams []api.Stream
	Err     error
}

type streamItem struct {
	stream    api.Stream
	movieName string
}

func (s streamItem) Title() string {
	name := s.stream.Name
	if strings.TrimSpace(name) == "" {
		name = "Stream"
	}
	if strings.TrimSpace(s.movieName) != "" {
		return "▶ " + name + " · " + s.movieName
	}
	return "▶ " + name
}

func (s streamItem) Description() string {
	if strings.TrimSpace(s.stream.Description) != "" {
		return s.stream.Description
	}
	if strings.TrimSpace(s.stream.URL) == "" {
		return "No playable link"
	}
	return "Ready to play"
}

func (s streamItem) FilterValue() string {
	return s.stream.Name + " " + s.stream.Description + " " + s.movieName
}

type StreamModel struct {
	apiClient *api.Client
	styles    Styles

	movie       api.Movie
	parentQuery string

	input      textinput.Model
	allStreams []api.Stream
	list       list.Model
	spin       spinner.Model

	loading    bool
	errMsg     string
	nowPlaying string
}

func NewStreamModel(client *api.Client, styles Styles, movie api.Movie, parentQuery string) *StreamModel {
	ti := textinput.New()
	ti.Prompt = "🔍 "
	ti.Placeholder = "Filter streams or search for another movie..."
	ti.CharLimit = 120
	ti.Focus()

	l := list.New([]list.Item{}, NewCompactDelegate(), 0, 0)
	l.Title = "Streams"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	sp := spinner.New()
	sp.Spinner = spinner.Line

	return &StreamModel{
		apiClient:   client,
		styles:      styles,
		movie:       movie,
		parentQuery: parentQuery,
		input:       ti,
		allStreams:  nil,
		list:        l,
		spin:        sp,
		loading:     true,
		errMsg:      "",
		nowPlaying:  "",
	}
}

func (m *StreamModel) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, fetchStreamsCmd(m.apiClient, m.movie.ID))
}

func (m *StreamModel) SetSize(width, height int) {
	listHeight := height - 7
	if listHeight < 6 {
		listHeight = 6
	}
	m.list.SetSize(width-2, listHeight)
}

func (m *StreamModel) applyFilter(raw string) {
	q := strings.ToLower(strings.TrimSpace(raw))
	items := make([]list.Item, 0, len(m.allStreams))
	for _, s := range m.allStreams {
		si := streamItem{stream: s, movieName: m.movie.Name}
		if q == "" || strings.Contains(strings.ToLower(si.FilterValue()), q) {
			items = append(items, si)
		}
	}
	m.list.SetItems(items)
	if len(items) > 0 {
		m.list.Select(0)
	}
}

func (m *StreamModel) shouldUpdateList(msg tea.Msg) bool {
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

func (m *StreamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case streamLoadedMsg:
		m.loading = false
		m.errMsg = ""
		if msg.Err != nil {
			m.errMsg = "Could not load stream list."
			m.allStreams = nil
			m.list.SetItems([]list.Item{})
			return m, nil
		}
		m.allStreams = msg.Streams
		m.applyFilter(m.input.Value())
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
			return m, func() tea.Msg { return backSearchMsg{Query: m.parentQuery} }
		case "alt+enter":
			q := strings.TrimSpace(m.input.Value())
			if q == "" {
				return m, nil
			}
			return m, func() tea.Msg { return openSearchMsg{Query: q} }
		case "enter":
			if m.loading {
				return m, nil
			}
			selected, ok := m.list.SelectedItem().(streamItem)
			if !ok {
				return m, nil
			}
			if strings.TrimSpace(selected.stream.URL) == "" {
				m.errMsg = "This stream has no playable link."
				return m, nil
			}
			m.nowPlaying = selected.stream.Name
			return m, func() tea.Msg {
				return playRequestedMsg{Movie: m.movie, Stream: selected.stream}
			}
		}
	}

	before := m.input.Value()
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	if m.input.Value() != before {
		m.applyFilter(m.input.Value())
	}

	cmds := []tea.Cmd{inputCmd}
	if m.shouldUpdateList(msg) {
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *StreamModel) View() string {
	head := []string{
		m.styles.Title.Render("VNStream Streams"),
		m.styles.Header.Render(fmt.Sprintf("Movie: %s", m.movie.Name)),
	}

	if m.nowPlaying != "" {
		head = append(head, m.styles.Highlight.Render("Now playing: "+m.nowPlaying))
	}

	body := ""
	if m.loading {
		body = m.styles.Busy.Render(m.spin.View() + " Loading streams...")
	} else if m.errMsg != "" {
		body = m.styles.Error.Render("✗ " + m.errMsg)
	} else if len(m.list.Items()) == 0 {
		body = m.styles.Muted.Render("No streams available for this movie.")
	} else {
		body = m.list.View()
	}

	inputSection := m.input.View()
	footer := m.styles.Footer.Render("↑↓: move · Enter: play · Alt+Enter: new search with typed query · ←: back to Search")
	return strings.Join(head, "\n") + "\n\n" + body + "\n\n" + inputSection + "\n" + footer
}

func fetchStreamsCmd(client *api.Client, movieID string) tea.Cmd {
	return func() tea.Msg {
		streams, err := client.FetchStreams(context.Background(), movieID)
		return streamLoadedMsg{Streams: streams, Err: err}
	}
}
