package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyanite/syntax/internal/editor"
	"github.com/kyanite/syntax/internal/scene"
	"github.com/kyanite/syntax/internal/story"
	"github.com/kyanite/syntax/internal/theme"
)

// Screen represents different app screens
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenProjectList
	ScreenEditor
	ScreenCharacters
	ScreenScenes
	ScreenLocations
	ScreenTextEditor
	ScreenHelp
	ScreenStats
	ScreenExport
)

// Model is the root Bubble Tea model
type Model struct {
	CurrentScreen  Screen
	Width          int
	Height         int
	ThemeManager   *theme.Manager
	CurrentTheme   theme.Theme
	Styles         theme.Styles
	CurrentProject *story.Project
	Message        string
	Error          error

	// Sub-models
	Projects      []*story.Project
	SelectedIndex int
	InputMode     bool
	InputValue    string

	// Editor state
	CurrentScene   *scene.Scene
	Buffer         *editor.Buffer
	EditorMode     EditorMode
	PreviousScreen Screen
}

// NewModel creates a new root model
func NewModel() Model {
	themeManager := theme.NewManager("monochrome")
	currentTheme := themeManager.GetCurrent()

	return Model{
		CurrentScreen: ScreenWelcome,
		ThemeManager:  themeManager,
		CurrentTheme:  currentTheme,
		Styles:        currentTheme.ApplyTheme(),
		SelectedIndex: 0,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global shortcuts
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit

	case "ctrl+shift+t":
		// Cycle theme
		m.CurrentTheme = m.ThemeManager.NextTheme()
		m.Styles = m.CurrentTheme.ApplyTheme()
		m.Message = fmt.Sprintf("Theme: %s", m.CurrentTheme.Name)
		return m, nil

	case "?", "h":
		// Show help
		m.PreviousScreen = m.CurrentScreen
		m.CurrentScreen = ScreenHelp
		return m, nil
	}

	// Screen-specific shortcuts
	switch m.CurrentScreen {
	case ScreenWelcome:
		return m.handleWelcomeKeys(msg)
	case ScreenProjectList:
		return m.handleProjectListKeys(msg)
	case ScreenEditor:
		return m.handleEditorKeys(msg)
	case ScreenCharacters:
		return m.handleCharactersKeys(msg)
	case ScreenScenes:
		return m.handleScenesKeys(msg)
	case ScreenLocations:
		return m.handleLocationsKeys(msg)
	case ScreenTextEditor:
		return m.handleTextEditorKeys(msg)
	case ScreenHelp:
		return m.handleHelpKeys(msg)
	case ScreenStats:
		return m.handleStatsKeys(msg)
	case ScreenExport:
		return m.handleExportKeys(msg)
	}

	return m, nil
}

// View renders the current screen
func (m Model) View() string {
	switch m.CurrentScreen {
	case ScreenWelcome:
		return m.viewWelcome()
	case ScreenProjectList:
		return m.viewProjectList()
	case ScreenEditor:
		return m.viewEditor()
	case ScreenCharacters:
		return m.viewCharacters()
	case ScreenScenes:
		return m.viewScenes()
	case ScreenLocations:
		return m.viewLocations()
	case ScreenTextEditor:
		return m.viewTextEditor()
	case ScreenHelp:
		return m.viewHelp()
	case ScreenStats:
		return m.viewStats()
	case ScreenExport:
		return m.viewExport()
	default:
		return "Unknown screen"
	}
}
