package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyanite/syntax/internal/storage"
)

// EditorMode represents the editor state
type EditorMode int

const (
	EditorModeNormal EditorMode = iota
	EditorModeInsert
	EditorModeSearch
	EditorModeReplace
)

func (m Model) viewTextEditor() string {
	if m.CurrentProject == nil || m.CurrentScene == nil {
		return "No scene loaded"
	}

	var b strings.Builder

	// Title bar
	titleBar := m.Styles.StatusBar.Render(fmt.Sprintf(" %s - Editing: %s | %s ",
		m.CurrentProject.Title, m.CurrentScene.Name, m.CurrentTheme.Name))
	b.WriteString(titleBar)
	b.WriteString("\n")

	// Calculate dimensions for split pane
	contentHeight := m.Height - 4 // Minus title and status bars
	paneWidth := m.Width / 2

	// Editor pane (left)
	editorContent := m.renderEditorPane(paneWidth, contentHeight)

	// Preview pane (right)
	previewContent := m.renderPreviewPane(paneWidth, contentHeight)

	// Combine panes side by side
	editorLines := strings.Split(editorContent, "\n")
	previewLines := strings.Split(previewContent, "\n")

	maxLines := max(len(editorLines), len(previewLines))
	for i := 0; i < maxLines && i < contentHeight; i++ {
		var editorLine, previewLine string
		if i < len(editorLines) {
			editorLine = editorLines[i]
		}
		if i < len(previewLines) {
			previewLine = previewLines[i]
		}

		// Pad to pane width
		editorLine = padRight(editorLine, paneWidth)
		previewLine = padRight(previewLine, paneWidth)

		b.WriteString(editorLine)
		b.WriteString("│")
		b.WriteString(previewLine)
		b.WriteString("\n")
	}

	// Search input box (overlay when in search mode)
	if m.EditorMode == EditorModeSearch && m.InputMode {
		searchBox := m.Styles.Accent.Render(fmt.Sprintf(" Find: %s█ ", m.InputValue))
		b.WriteString(searchBox)
		b.WriteString("\n")
	}

	// Status bar
	mode := "NORMAL"
	switch m.EditorMode {
	case EditorModeInsert:
		mode = "INSERT"
	case EditorModeSearch:
		mode = "SEARCH"
	case EditorModeReplace:
		mode = "REPLACE"
	}

	line, col := m.Buffer.CursorPosition()
	wordCount := len(strings.Fields(m.Buffer.GetContent()))

	aiStatus := ""
	if m.AIClient != nil && m.AIClient.IsEnabled() {
		aiStatus = " | AI: ON"
	}

	// Save status indicator
	saveStatus := ""
	switch m.SaveStatus {
	case SaveStatusSaved:
		elapsed := time.Since(m.LastSaveTime)
		if elapsed < time.Minute {
			saveStatus = fmt.Sprintf(" | Saved %ds ago", int(elapsed.Seconds()))
		} else {
			saveStatus = fmt.Sprintf(" | Saved %dm ago", int(elapsed.Minutes()))
		}
	case SaveStatusSaving:
		saveStatus = " | Saving..."
	case SaveStatusUnsaved:
		if m.Buffer.IsModified() {
			saveStatus = " | Unsaved changes"
		}
	}

	// Search info
	searchInfo := ""
	if m.EditorMode == EditorModeSearch || m.EditorMode == EditorModeReplace {
		searchTerm, current, total := m.Buffer.GetSearchInfo()
		if total > 0 {
			searchInfo = fmt.Sprintf(" | Search: '%s' (%d/%d)", searchTerm, current, total)
		} else if searchTerm != "" {
			searchInfo = fmt.Sprintf(" | Search: '%s' (no matches)", searchTerm)
		} else {
			searchInfo = " | Enter search term"
		}
	}

	statusBar := m.Styles.StatusBar.Render(fmt.Sprintf(
		" %s | Line %d:%d | Words: %d%s%s%s | Ctrl+F: Find | Ctrl+S: Save | Esc: Exit ",
		mode, line+1, col+1, wordCount, aiStatus, saveStatus, searchInfo))
	b.WriteString(statusBar)

	return b.String()
}

func (m Model) renderEditorPane(width, height int) string {
	if m.Buffer == nil {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString(m.Styles.Heading.Render(" EDITOR"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-2))
	b.WriteString("\n")

	line, col := m.Buffer.CursorPosition()
	lines := m.Buffer.GetLines()

	// Simple viewport - show lines around cursor
	startLine := max(0, line-height/2)
	endLine := min(len(lines), startLine+height-3)

	for i := startLine; i < endLine; i++ {
		lineText := lines[i]

		// Show line numbers
		lineNum := fmt.Sprintf("%4d ", i+1)

		// Highlight current line
		if i == line {
			lineNum = m.Styles.Accent.Render(lineNum)

			// Show cursor
			if m.EditorMode == EditorModeInsert && col <= len(lineText) {
				before := lineText[:col]
				cursor := "█"
				after := ""
				if col < len(lineText) {
					after = lineText[col:]
				}
				lineText = before + m.Styles.Accent.Render(cursor) + after
			}
		} else {
			lineNum = m.Styles.Text.Faint(true).Render(lineNum)
		}

		b.WriteString(lineNum)
		b.WriteString(lineText)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderPreviewPane(width, height int) string {
	if m.Buffer == nil {
		return "No preview"
	}

	var b strings.Builder
	b.WriteString(m.Styles.Heading.Render(" PREVIEW"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-2))
	b.WriteString("\n")

	// Simple markdown rendering
	content := m.Buffer.GetContent()
	lines := strings.Split(content, "\n")

	for i := 0; i < min(len(lines), height-3); i++ {
		line := lines[i]
		rendered := m.renderMarkdownLine(line)
		b.WriteString(rendered)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderMarkdownLine(line string) string {
	// Simple markdown rendering
	if strings.HasPrefix(line, "# ") {
		return m.Styles.Heading.Bold(true).Render(strings.TrimPrefix(line, "# "))
	} else if strings.HasPrefix(line, "## ") {
		return m.Styles.Heading.Render(strings.TrimPrefix(line, "## "))
	} else if strings.HasPrefix(line, "### ") {
		return m.Styles.Accent.Render(strings.TrimPrefix(line, "### "))
	} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return m.Styles.Text.Render("  • " + line[2:])
	} else if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
		return m.Styles.Text.Bold(true).Render(strings.Trim(line, "**"))
	} else if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") {
		return m.Styles.Text.Italic(true).Render(strings.Trim(line, "*"))
	}

	return m.Styles.Text.Render(line)
}

func (m Model) handleTextEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search mode
	if m.EditorMode == EditorModeSearch {
		switch msg.String() {
		case "esc":
			m.EditorMode = EditorModeNormal
			m.Buffer.ClearSearch()
			m.InputMode = false
			m.InputValue = ""
			return m, nil
		case "enter":
			// Perform search
			if m.InputValue != "" {
				count := m.Buffer.Find(m.InputValue, false)
				if count == 0 {
					m.Message = "No matches found"
				} else {
					m.Message = fmt.Sprintf("Found %d matches", count)
				}
			}
			m.InputMode = false
			return m, nil
		case "ctrl+n", "n":
			// Find next
			m.Buffer.FindNext()
			return m, nil
		case "ctrl+p", "p":
			// Find previous
			m.Buffer.FindPrevious()
			return m, nil
		case "backspace":
			if m.InputMode && len(m.InputValue) > 0 {
				m.InputValue = m.InputValue[:len(m.InputValue)-1]
			}
			return m, nil
		default:
			// Add character to search input
			if m.InputMode && len(msg.String()) == 1 {
				m.InputValue += msg.String()
			}
			return m, nil
		}
	}

	// Mode switching
	if m.EditorMode == EditorModeNormal {
		switch msg.String() {
		case "i":
			m.EditorMode = EditorModeInsert
			return m, nil
		case "esc":
			// Save and exit
			if m.Buffer.IsModified() {
				m.CurrentScene.Content = m.Buffer.GetContent()
				storage.SaveScene(m.CurrentProject.Directory, m.CurrentScene)
				m.Message = "Scene saved"
			}
			m.Buffer.ClearSearch()
			m.CurrentScene = nil
			m.Buffer = nil
			m.CurrentScreen = ScreenScenes
			return m, nil
		case "ctrl+f":
			// Enter search mode
			m.EditorMode = EditorModeSearch
			m.InputMode = true
			m.InputValue = ""
			return m, nil
		case "n":
			// Find next (if search active)
			m.Buffer.FindNext()
			return m, nil
		case "N":
			// Find previous (if search active)
			m.Buffer.FindPrevious()
			return m, nil
		case "ctrl+s":
			// Manual save
			m.SaveStatus = SaveStatusSaving
			m.CurrentScene.Content = m.Buffer.GetContent()
			err := storage.SaveScene(m.CurrentProject.Directory, m.CurrentScene)
			if err != nil {
				m.Error = err
				m.SaveStatus = SaveStatusUnsaved
			} else {
				m.Buffer.SetModified(false)
				m.SaveStatus = SaveStatusSaved
				m.LastSaveTime = time.Now()
				m.Message = "Saved"
			}
			return m, nil
		case "ctrl+z":
			m.Buffer.Undo()
			return m, nil
		case "ctrl+y":
			m.Buffer.Redo()
			return m, nil
		case "ctrl+a":
			// Show AI suggestion menu
			if m.AIClient != nil && m.AIClient.IsEnabled() {
				m.PreviousScreen = ScreenTextEditor
				m.CurrentScreen = ScreenAISuggestion
				m.SelectedIndex = 0
			} else {
				m.Message = "AI Assistant is disabled. Enable in config."
			}
			return m, nil
		}
	} else if m.EditorMode == EditorModeInsert {
		switch msg.String() {
		case "esc":
			m.EditorMode = EditorModeNormal
			return m, nil
		case "enter":
			m.Buffer.InsertNewline()
			m.LastEditTime = time.Now()
			m.SaveStatus = SaveStatusUnsaved
			return m, nil
		case "backspace":
			m.Buffer.DeleteChar()
			m.LastEditTime = time.Now()
			m.SaveStatus = SaveStatusUnsaved
			return m, nil
		case "up":
			m.Buffer.MoveCursorUp()
			return m, nil
		case "down":
			m.Buffer.MoveCursorDown()
			return m, nil
		case "left":
			m.Buffer.MoveCursorLeft()
			return m, nil
		case "right":
			m.Buffer.MoveCursorRight()
			return m, nil
		default:
			// Regular character input
			if len(msg.String()) == 1 {
				m.Buffer.InsertRune(rune(msg.String()[0]))
				m.LastEditTime = time.Now()
				m.SaveStatus = SaveStatusUnsaved
				return m, nil
			}
		}
	}

	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func padRight(s string, width int) string {
	// Remove ANSI codes for length calculation
	visibleLen := len(stripANSI(s))
	if visibleLen >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-visibleLen)
}

func stripANSI(s string) string {
	// Simple ANSI stripping (for length calculation)
	result := strings.Builder{}
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			result.WriteRune(r)
		}
	}
	return result.String()
}
