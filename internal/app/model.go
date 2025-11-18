package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyanite/syntax/internal/ai"
	_ "github.com/kyanite/syntax/internal/character" // Used in story.Project
	"github.com/kyanite/syntax/internal/editor"
	_ "github.com/kyanite/syntax/internal/location" // Used in story.Project
	"github.com/kyanite/syntax/internal/scene"
	"github.com/kyanite/syntax/internal/storage"
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
	ScreenAISuggestion
	ScreenRelationshipMap
	ScreenSceneValidation
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

	// AI Assistant state
	AIClient      *ai.Client
	AISuggestion  *ai.Suggestion
	AIGenerating  bool

	// Auto-save state
	LastSaveTime   time.Time
	SaveStatus     SaveStatus
	LastEditTime   time.Time
}

// SaveStatus represents the current save state
type SaveStatus int

const (
	SaveStatusSaved SaveStatus = iota
	SaveStatusSaving
	SaveStatusUnsaved
)

// AutoSaveTickMsg is sent periodically to trigger auto-save check
type AutoSaveTickMsg struct{}

// AutoSaveCompleteMsg is sent when auto-save completes
type AutoSaveCompleteMsg struct {
	Err error
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
		SaveStatus:    SaveStatusSaved,
		LastSaveTime:  time.Now(),
	}
}

// autoSaveTick creates a periodic tick command for auto-save
func autoSaveTick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return AutoSaveTickMsg{}
	})
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Start auto-save ticker
	return autoSaveTick()
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

	case AISuggestionMsg:
		m = m.HandleAISuggestionMsg(msg)
		return m, nil

	case AutoSaveTickMsg:
		// Check if auto-save is needed
		cmd := m.checkAutoSave()
		// Schedule next tick
		return m, tea.Batch(cmd, autoSaveTick())

	case AutoSaveCompleteMsg:
		if msg.Err != nil {
			m.SaveStatus = SaveStatusUnsaved
			m.Error = msg.Err
		} else {
			m.SaveStatus = SaveStatusSaved
			m.LastSaveTime = time.Now()
			if m.Buffer != nil {
				m.Buffer.SetModified(false)
			}
		}
		return m, nil
	}

	return m, nil
}

// checkAutoSave performs auto-save if needed
func (m *Model) checkAutoSave() tea.Cmd {
	// Only auto-save in text editor with unsaved changes
	if m.CurrentScreen != ScreenTextEditor || m.Buffer == nil || m.CurrentScene == nil || m.CurrentProject == nil {
		return nil
	}

	// Check if buffer has been modified
	if !m.Buffer.IsModified() {
		return nil
	}

	// Check if enough time has passed since last edit (3 seconds idle)
	if time.Since(m.LastEditTime) < 3*time.Second {
		return nil
	}

	// Perform auto-save
	m.SaveStatus = SaveStatusSaving
	projectDir := m.CurrentProject.Directory
	scene := m.CurrentScene
	content := m.Buffer.GetContent()

	return func() tea.Msg {
		scene.Content = content
		err := storage.SaveScene(projectDir, scene)
		return AutoSaveCompleteMsg{Err: err}
	}
}

// ensureDataLoaded ensures necessary data is loaded for the current screen
func (m *Model) ensureDataLoaded() {
	if m.CurrentProject == nil {
		return
	}

	switch m.CurrentScreen {
	case ScreenScenes, ScreenSceneValidation:
		if m.CurrentProject.Scenes == nil || len(m.CurrentProject.Scenes) == 0 {
			scenes, err := storage.LoadAllScenes(m.CurrentProject.Directory)
			if err == nil {
				m.CurrentProject.Scenes = scenes
			}
		}
	case ScreenCharacters, ScreenRelationshipMap:
		if m.CurrentProject.Characters == nil || len(m.CurrentProject.Characters) == 0 {
			chars, err := storage.LoadAllCharacters(m.CurrentProject.Directory)
			if err == nil {
				m.CurrentProject.Characters = chars
			}
		}
	case ScreenLocations:
		if m.CurrentProject.Locations == nil || len(m.CurrentProject.Locations) == 0 {
			locs, err := storage.LoadAllLocations(m.CurrentProject.Directory)
			if err == nil {
				m.CurrentProject.Locations = locs
			}
		}
	}
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Store previous screen to detect changes
	prevScreen := m.CurrentScreen

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
		m.ensureDataLoaded()
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
		updatedModel, cmd := m.handleCharactersKeys(msg)
		m = updatedModel.(Model)
		if m.CurrentScreen != prevScreen {
			m.ensureDataLoaded()
		}
		return m, cmd
	case ScreenScenes:
		updatedModel, cmd := m.handleScenesKeys(msg)
		m = updatedModel.(Model)
		if m.CurrentScreen != prevScreen {
			m.ensureDataLoaded()
		}
		return m, cmd
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
	case ScreenAISuggestion:
		return m.handleAISuggestionKeys(msg)
	case ScreenRelationshipMap:
		updatedModel, cmd := m.handleRelationshipMapKeys(msg)
		m = updatedModel.(Model)
		if m.CurrentScreen != prevScreen {
			m.ensureDataLoaded()
		}
		return m, cmd
	case ScreenSceneValidation:
		updatedModel, cmd := m.handleSceneValidationKeys(msg)
		m = updatedModel.(Model)
		if m.CurrentScreen != prevScreen {
			m.ensureDataLoaded()
		}
		return m, cmd
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
	case ScreenAISuggestion:
		return m.viewAISuggestion()
	case ScreenRelationshipMap:
		return m.viewRelationshipMap()
	case ScreenSceneValidation:
		return m.viewSceneValidation()
	default:
		return "Unknown screen"
	}
}
