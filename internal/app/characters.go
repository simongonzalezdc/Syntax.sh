package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyanite/syntax/internal/character"
	"github.com/kyanite/syntax/internal/storage"
)

func (m Model) viewCharacters() string {
	if m.CurrentProject == nil {
		return "No project loaded"
	}

	var b strings.Builder

	// Title bar
	titleBar := m.Styles.StatusBar.Render(fmt.Sprintf(" %s - Characters ", m.CurrentProject.Title))
	b.WriteString(titleBar)
	b.WriteString("\n\n")

	// Load characters if not loaded
	if m.CurrentProject.Characters == nil || len(m.CurrentProject.Characters) == 0 {
		chars, err := storage.LoadAllCharacters(m.CurrentProject.Directory)
		if err == nil {
			m.CurrentProject.Characters = chars
		}
	}

	b.WriteString(m.Styles.Heading.Render("👥 Characters"))
	b.WriteString("\n\n")

	if len(m.CurrentProject.Characters) == 0 {
		b.WriteString(m.Styles.Text.Render("No characters yet. Press 'n' to create one."))
	} else {
		for _, char := range m.CurrentProject.Characters {
			b.WriteString(m.Styles.Accent.Render(fmt.Sprintf("• %s", char.Name)))
			if char.Role != "" {
				b.WriteString(m.Styles.Text.Render(fmt.Sprintf(" (%s)", char.Role)))
			}
			b.WriteString("\n")
			if char.Occupation != "" {
				b.WriteString(m.Styles.Text.Render(fmt.Sprintf("  %s", char.Occupation)))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(m.Styles.Text.Render("n - New Character  Esc - Back"))

	if m.Message != "" {
		b.WriteString("\n\n")
		b.WriteString(m.Styles.Success.Render(m.Message))
	}

	return b.String()
}

func (m Model) handleCharactersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n":
		// Create new character (simplified)
		char := &character.Character{
			Name:       "New Character",
			Role:       "supporting",
			Occupation: "Adventurer",
		}

		err := storage.SaveCharacter(m.CurrentProject.Directory, char)
		if err != nil {
			m.Error = err
			return m, nil
		}

		if m.CurrentProject.Characters == nil {
			m.CurrentProject.Characters = make(map[string]*character.Character)
		}
		m.CurrentProject.Characters[char.ID] = char
		m.Message = fmt.Sprintf("Created character: %s", char.Name)
		return m, nil

	case "esc":
		m.CurrentScreen = ScreenEditor
		m.Message = ""
		return m, nil
	}

	return m, nil
}
