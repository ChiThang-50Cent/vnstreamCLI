package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ChiThang-50Cent/vnstream/internal/api"
	"github.com/ChiThang-50Cent/vnstream/internal/config"
	"github.com/ChiThang-50Cent/vnstream/internal/player"
	"github.com/ChiThang-50Cent/vnstream/internal/storage"
)

const (
	screenHome   = "home"
	screenSearch = "search"
	screenStream = "stream"
)

type openSearchMsg struct{ Query string }
type openStreamMsg struct {
	Movie       api.Movie
	ParentQuery string
}
type backHomeMsg struct{}
type backSearchMsg struct{ Query string }
type replayWatchedMsg struct{ Item storage.WatchedItem }

type resolveReplayDoneMsg struct {
	Item    storage.WatchedItem
	MovieID string
	Err     error
}

type playRequestedMsg struct {
	Movie  api.Movie
	Stream api.Stream
}

type playDoneMsg struct{ Err error }

type AppModel struct {
	cfg    config.Config
	styles Styles

	apiClient *api.Client
	storage   *storage.Manager
	player    *player.Launcher

	home   *HomeModel
	search *SearchModel
	stream *StreamModel
	screen string

	width  int
	height int

	busyText string
	notice   string
}

func NewAppModel(cfg config.Config, initialQuery string) (*AppModel, error) {
	store := storage.NewManager(cfg)
	if err := store.EnsureFiles(); err != nil {
		return nil, err
	}

	styles := NewStyles()
	home := NewHomeModel(store, styles)

	app := &AppModel{
		cfg:       cfg,
		styles:    styles,
		apiClient: api.NewClient(cfg),
		storage:   store,
		player:    player.NewLauncher(cfg),
		home:      home,
		screen:    screenHome,
	}

	initialQuery = strings.TrimSpace(initialQuery)
	if initialQuery != "" {
		app.screen = screenSearch
		app.search = NewSearchModel(app.apiClient, app.styles, initialQuery)
		_ = app.storage.SaveHistory(initialQuery)
	}

	return app, nil
}

func (a *AppModel) Init() tea.Cmd {
	if a.screen == screenSearch && a.search != nil {
		return a.search.Init()
	}
	return nil
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.home.SetSize(msg.Width, msg.Height)
		if a.search != nil {
			a.search.SetSize(msg.Width, msg.Height)
		}
		if a.stream != nil {
			a.stream.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}

	case openSearchMsg:
		q := strings.TrimSpace(msg.Query)
		a.notice = ""
		if q != "" {
			_ = a.storage.SaveHistory(q)
		}
		// Reuse cached search model if query matches.
		if a.search != nil && a.search.currentQuery == q {
			a.search.SetSize(a.width, a.height)
			a.screen = screenSearch
			return a, nil
		}
		a.search = NewSearchModel(a.apiClient, a.styles, q)
		a.search.SetSize(a.width, a.height)
		a.screen = screenSearch
		return a, a.search.Init()

	case openStreamMsg:
		a.notice = ""
		a.stream = NewStreamModel(a.apiClient, a.styles, msg.Movie, msg.ParentQuery)
		a.stream.SetSize(a.width, a.height)
		a.screen = screenStream
		return a, a.stream.Init()

	case backHomeMsg:
		a.notice = ""
		a.busyText = ""
		a.home.Reload()
		a.screen = screenHome
		return a, nil

	case backSearchMsg:
		a.notice = ""
		q := strings.TrimSpace(msg.Query)
		// Reuse the existing search model if query matches (back from Stream screen).
		if a.search != nil && a.search.currentQuery == q {
			a.search.SetSize(a.width, a.height)
			a.screen = screenSearch
			return a, nil
		}
		a.search = NewSearchModel(a.apiClient, a.styles, q)
		a.search.SetSize(a.width, a.height)
		a.screen = screenSearch
		return a, a.search.Init()

	case replayWatchedMsg:
		if strings.TrimSpace(msg.Item.MovieID) != "" {
			movie := api.Movie{ID: msg.Item.MovieID, Name: msg.Item.MovieName}
			a.stream = NewStreamModel(a.apiClient, a.styles, movie, msg.Item.MovieName)
			a.stream.SetSize(a.width, a.height)
			a.screen = screenStream
			return a, a.stream.Init()
		}
		a.busyText = "Resolving movie ID from watch history..."
		return a, resolveReplayCmd(a.apiClient, msg.Item)

	case resolveReplayDoneMsg:
		a.busyText = ""
		if strings.TrimSpace(msg.MovieID) != "" {
			movie := api.Movie{ID: msg.MovieID, Name: msg.Item.MovieName}
			a.stream = NewStreamModel(a.apiClient, a.styles, movie, msg.Item.MovieName)
			a.stream.SetSize(a.width, a.height)
			a.screen = screenStream
			return a, a.stream.Init()
		}
		a.notice = "Could not resolve source list, replaying previous link."
		return a, launchVLCCmd(a.player, msg.Item.Link, msg.Item.MovieName, msg.Item.StreamName)

	case playRequestedMsg:
		a.busyText = "Launching VLC..."
		if err := a.storage.SaveWatched(msg.Movie.Name, msg.Stream.Name, msg.Stream.URL, msg.Movie.ID); err != nil {
			a.notice = "Could not save watch history."
		}
		return a, launchVLCCmd(a.player, msg.Stream.URL, msg.Movie.Name, msg.Stream.Name)

	case playDoneMsg:
		a.busyText = ""
		if msg.Err != nil {
			a.notice = "Could not launch VLC or playback failed."
		} else {
			a.notice = "Playing video in VLC."
		}
		return a, nil
	}

	var cmd tea.Cmd
	switch a.screen {
	case screenHome:
		var m tea.Model
		m, cmd = a.home.Update(msg)
		a.home = m.(*HomeModel)
	case screenSearch:
		if a.search != nil {
			var m tea.Model
			m, cmd = a.search.Update(msg)
			a.search = m.(*SearchModel)
		}
	case screenStream:
		if a.stream != nil {
			var m tea.Model
			m, cmd = a.stream.Update(msg)
			a.stream = m.(*StreamModel)
		}
	}

	return a, cmd
}

func (a *AppModel) View() string {
	var body string
	switch a.screen {
	case screenSearch:
		if a.search != nil {
			body = a.search.View()
		}
	case screenStream:
		if a.stream != nil {
			body = a.stream.View()
		}
	default:
		body = a.home.View()
	}

	footer := ""
	if a.busyText != "" {
		footer += a.styles.Busy.Render("⏳ " + a.busyText)
	}
	if a.notice != "" {
		if footer != "" {
			footer += "\n"
		}
		footer += a.styles.Notice.Render("ℹ " + a.notice)
	}

	if footer != "" {
		return body + "\n" + footer
	}
	return body
}

type homeItem struct {
	title string
	desc  string
	kind  string
	data  string
	watch storage.WatchedItem
}

func (i homeItem) Title() string       { return i.title }
func (i homeItem) Description() string { return i.desc }
func (i homeItem) IsSeparator() bool   { return i.kind == "separator" }
func (i homeItem) FilterValue() string {
	if i.kind == "separator" {
		return ""
	}
	return i.title + " " + i.desc
}

type HomeModel struct {
	store  *storage.Manager
	styles Styles

	input textinput.Model
	list  list.Model

	allItems []homeItem

	notice string

	pendingConfirm string
}

func NewHomeModel(store *storage.Manager, styles Styles) *HomeModel {
	ti := textinput.New()
	ti.Placeholder = "Type a movie keyword..."
	ti.Prompt = "🔍 "
	ti.Focus()
	ti.CharLimit = 120

	l := list.New([]list.Item{}, NewCompactDelegate(), 0, 0)
	l.Title = "Home"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	m := &HomeModel{
		store:  store,
		styles: styles,
		input:  ti,
		list:   l,
	}
	m.Reload()
	return m
}

func (m *HomeModel) Init() tea.Cmd { return nil }

func (m *HomeModel) Reload() {
	searches, _ := m.store.LoadSearchHistory(20)
	watched, _ := m.store.LoadWatchedHistory(30)

	items := make([]homeItem, 0, len(searches)+len(watched)+6)

	if len(searches) > 0 {
		items = append(items, homeItem{title: "--- Search History ---", kind: "separator"})
		for _, q := range searches {
			items = append(items, homeItem{title: "🕐 " + q, kind: "search", data: q})
		}
	}

	if len(watched) > 0 {
		items = append(items, homeItem{title: "--- Watch History ---", kind: "separator"})
		for _, w := range watched {
			// Format: timestamp · stream name · movie name.
			title := fmt.Sprintf("🎬 %s · %s · %s", shortTS(w.Timestamp), w.StreamName, w.MovieName)
			items = append(items, homeItem{title: title, kind: "watch", data: w.MovieName, watch: w})
		}
	}

	items = append(items,
		homeItem{title: "--- Actions ---", kind: "separator"},
		homeItem{title: "🗑 Clear watch history", kind: "clear_watched"},
		homeItem{title: "🗑 Clear search history", kind: "clear_search"},
	)

	m.allItems = items
	m.applyFilter(m.input.Value())
}

func (m *HomeModel) applyFilter(raw string) {
	q := strings.ToLower(strings.TrimSpace(raw))

	// When no filter is set, show everything.
	if q == "" {
		all := make([]list.Item, len(m.allItems))
		for i, it := range m.allItems {
			all[i] = it
		}
		m.list.SetItems(all)
		if len(all) > 0 {
			// Select first non-separator.
			for idx, it := range all {
				if !it.(homeItem).IsSeparator() {
					m.list.Select(idx)
					break
				}
			}
		}
		return
	}

	// When filtering, group by sections and drop separators with no matching rows.
	var result []list.Item
	var pendingSep *homeItem
	for _, item := range m.allItems {
		if item.IsSeparator() {
			pendingSep = &homeItem{title: item.title, kind: item.kind}
			continue
		}
		if strings.Contains(strings.ToLower(item.FilterValue()), q) {
			if pendingSep != nil {
				result = append(result, *pendingSep)
				pendingSep = nil
			}
			result = append(result, item)
		}
	}
	m.list.SetItems(result)
	if len(result) > 0 {
		for idx, it := range result {
			if !it.(homeItem).IsSeparator() {
				m.list.Select(idx)
				break
			}
		}
	}
}

func (m *HomeModel) shouldUpdateList(msg tea.Msg) bool {
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

func (m *HomeModel) SetSize(width, height int) {
	listHeight := height - 6
	if listHeight < 6 {
		listHeight = 6
	}
	m.list.SetSize(width-2, listHeight)
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.pendingConfirm != "" {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "y", "Y", "enter":
				switch m.pendingConfirm {
				case "clear_watched":
					_ = m.store.ClearWatchedHistory()
					m.notice = "Watch history cleared."
				case "clear_search":
					_ = m.store.ClearSearchHistory()
					m.notice = "Search history cleared."
				}
				m.pendingConfirm = ""
				m.Reload()
				return m, nil
			case "n", "N", "esc":
				m.pendingConfirm = ""
				return m, nil
			}
		}
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "alt+enter":
			q := strings.TrimSpace(m.input.Value())
			if q == "" {
				return m, nil
			}
			return m, func() tea.Msg { return openSearchMsg{Query: q} }
		case "enter":
			selected, ok := m.list.SelectedItem().(homeItem)
			if ok {
				if selected.IsSeparator() {
					return m, nil
				}
				switch selected.kind {
				case "search":
					return m, func() tea.Msg { return openSearchMsg{Query: selected.data} }
				case "watch":
					return m, func() tea.Msg { return replayWatchedMsg{Item: selected.watch} }
				case "clear_watched", "clear_search":
					m.pendingConfirm = selected.kind
					return m, nil
				}
			}

			if len(m.list.Items()) == 0 {
				q := strings.TrimSpace(m.input.Value())
				if q == "" {
					return m, nil
				}
				return m, func() tea.Msg { return openSearchMsg{Query: q} }
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

func (m *HomeModel) View() string {
	title := m.styles.Title.Render("VNStream") + "  " + m.styles.Help.Render("Type to filter + quick search")

	header := []string{
		title,
		m.styles.Header.Render("Home"),
	}

	body := m.list.View()
	if m.notice != "" {
		body += "\n" + m.styles.Success.Render("✓ "+m.notice)
	}

	if m.pendingConfirm != "" {
		prompt := "Are you sure?"
		if m.pendingConfirm == "clear_watched" {
			prompt = "Clear all watch history?"
		}
		if m.pendingConfirm == "clear_search" {
			prompt = "Clear all search history?"
		}
		body += "\n" + m.styles.Error.Render(prompt+" (y/n)")
	}

	inputSection := m.input.View()
	footer := m.styles.Footer.Render("↑↓: move · Enter: select/search when empty · Alt+Enter: search now · Ctrl+C: exit")
	return strings.Join(header, "\n") + "\n\n" + body + "\n\n" + inputSection + "\n" + footer
}

func resolveReplayCmd(client *api.Client, item storage.WatchedItem) tea.Cmd {
	return func() tea.Msg {
		id, err := client.ResolveMovieID(context.Background(), item.MovieName)
		return resolveReplayDoneMsg{Item: item, MovieID: id, Err: err}
	}
}

func launchVLCCmd(p *player.Launcher, link, movieName, streamName string) tea.Cmd {
	return func() tea.Msg {
		err := p.LaunchVLC(link, movieName, streamName)
		return playDoneMsg{Err: err}
	}
}

func shortTS(ts string) string {
	if len(ts) >= 16 {
		return ts[8:10] + "/" + ts[5:7] + " " + ts[11:16]
	}
	return ts
}
