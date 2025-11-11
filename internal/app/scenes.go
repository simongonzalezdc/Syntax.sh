package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyanite/syntax/internal/editor"
	"github.com/kyanite/syntax/internal/scene"
	"github.com/kyanite/syntax/internal/storage"
)

func (m Model) viewScenes() string {
	if m.CurrentProject == nil {
		return "No project loaded"
	}

	var b strings.Builder

	// Title bar
	titleBar := m.Styles.StatusBar.Render(fmt.Sprintf(" %s - Scenes ", m.CurrentProject.Title))
	b.WriteString(titleBar)
	b.WriteString("\n\n")

	// Load scenes if not loaded
	if m.CurrentProject.Scenes == nil || len(m.CurrentProject.Scenes) == 0 {
		scenes, err := storage.LoadAllScenes(m.CurrentProject.Directory)
		if err == nil {
			m.CurrentProject.Scenes = scenes
		}
	}

	b.WriteString(m.Styles.Heading.Render("📝 Scenes"))
	b.WriteString("\n\n")

	if len(m.CurrentProject.Scenes) == 0 {
		b.WriteString(m.Styles.Text.Render("No scenes yet. Press 'n' to create one."))
	} else {
		for _, sc := range m.CurrentProject.Scenes {
			b.WriteString(m.Styles.Accent.Render(fmt.Sprintf("• Ch%d Sc%d: %s", sc.Chapter, sc.SceneNumber, sc.Name)))
			b.WriteString(m.Styles.Text.Render(fmt.Sprintf(" [%s]", sc.Status)))
			b.WriteString("\n")
			b.WriteString(m.Styles.Text.Render(fmt.Sprintf("  %d words", sc.WordCount)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.Styles.Text.Render("n - New Scene | v - Validate | Enter - Edit | ↑/↓ - Navigate | Esc - Back"))

	if m.Message != "" {
		b.WriteString("\n\n")
		b.WriteString(m.Styles.Success.Render(m.Message))
	}

	return b.String()
}

func (m Model) handleScenesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedIndex > 0 {
			m.SelectedIndex--
		}
		return m, nil

	case "down", "j":
		sceneCount := len(m.CurrentProject.Scenes)
		if m.SelectedIndex < sceneCount-1 {
			m.SelectedIndex++
		}
		return m, nil

	case "enter":
		// Open scene in editor
		if len(m.CurrentProject.Scenes) > 0 {
			// Get scene at selected index (need to handle map iteration)
			idx := 0
			for _, sc := range m.CurrentProject.Scenes {
				if idx == m.SelectedIndex {
					m.CurrentScene = sc
					m.Buffer = editor.NewBuffer(sc.Content)
					m.EditorMode = EditorModeNormal
					m.CurrentScreen = ScreenTextEditor
					return m, nil
				}
				idx++
			}
		}
		return m, nil

	case "n":
		// Create new scene (simplified)
		sc := &scene.Scene{
			Chapter:     1,
			SceneNumber: len(m.CurrentProject.Scenes) + 1,
			Name:        "New Scene",
			Status:      "draft",
		}

		err := storage.SaveScene(m.CurrentProject.Directory, sc)
		if err != nil {
			m.Error = err
			return m, nil
		}

		if m.CurrentProject.Scenes == nil {
			m.CurrentProject.Scenes = make(map[string]*scene.Scene)
		}
		m.CurrentProject.Scenes[sc.ID] = sc
		m.Message = fmt.Sprintf("Created scene: %s", sc.Name)
		return m, nil

	case "v":
		// Show scene validation
		m.CurrentScreen = ScreenSceneValidation
		return m, nil

	case "esc":
		m.CurrentScreen = ScreenEditor
		m.Message = ""
		return m, nil
	}

	return m, nil
}
