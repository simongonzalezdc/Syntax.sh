package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyanite/syntax/internal/storage"
)

// EditorMode represents the editor state
type EditorMode int

const (
	EditorModeNormal EditorMode = iota
	EditorModeInsert
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

	// Status bar
	mode := "NORMAL"
	if m.EditorMode == EditorModeInsert {
		mode = "INSERT"
	}

	line, col := m.Buffer.CursorPosition()
	wordCount := len(strings.Fields(m.Buffer.GetContent()))

	aiStatus := ""
	if m.AIClient != nil && m.AIClient.IsEnabled() {
		aiStatus = " | AI: ON"
	}

	statusBar := m.Styles.StatusBar.Render(fmt.Sprintf(
		" %s | Line %d:%d | Words: %d%s | Ctrl+S: Save | Ctrl+A: AI | Esc: Exit ",
		mode, line+1, col+1, wordCount, aiStatus))
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
			m.CurrentScene = nil
			m.Buffer = nil
			m.CurrentScreen = ScreenScenes
			return m, nil
		case "ctrl+s":
			// Save
			m.CurrentScene.Content = m.Buffer.GetContent()
			err := storage.SaveScene(m.CurrentProject.Directory, m.CurrentScene)
			if err != nil {
				m.Error = err
			} else {
				m.Buffer.SetModified(false)
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
			return m, nil
		case "backspace":
			m.Buffer.DeleteChar()
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
