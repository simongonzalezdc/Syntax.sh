# syntax.sh - Technical Design Document

**Version:** 1.0  
**Date:** November 2025  
**For:** Independent Implementation  
**Status:** READY FOR IMPLEMENTATION

---

## Architecture Overview

```
┌────────────────────────────────────────────┐
│        Bubble Tea Root Model               │
├────────────────────────────────────────────┤
│                                            │
│  ┌──────────────┐  ┌──────────────┐      │
│  │ Editor Pane  │  │ Library Pane │      │
│  │              │  │ (Chars, Locs)│      │
│  └──────┬───────┘  └──────┬───────┘      │
│         │                 │              │
│         └─────────┬───────┘              │
│                   │                      │
│            ┌──────▼────────┐             │
│            │ Navigation    │             │
│            │ Router        │             │
│            └──────┬────────┘             │
├───────────────────┼────────────────────┤
│  Core Packages    │                    │
├───────────────────┼────────────────────┤
│                   │                    │
│  ┌───────┐  ┌────▼─────┐  ┌────────┐ │
│  │story/ │  │character/ │  │scene/  │ │
│  └───────┘  └──────────┘  └────────┘ │
│                                       │
│  ┌────────┐  ┌────────┐  ┌────────┐ │
│  │location│  │outline/│  │export/ │ │
│  └────────┘  └────────┘  └────────┘ │
│                                       │
│  ┌────────┐  ┌────────┐  ┌────────┐ │
│  │storage/│  │editor/ │  │stats/  │ │
│  └────────┘  └────────┘  └────────┘ │
│                                       │
│  ┌────────┐  ┌────────┐              │
│  │ai/     │  │theme/  │              │
│  └────────┘  └────────┘              │
│                                       │
└────────────────────────────────────────┘
```

---

## Tech Stack

### Core

```
Go 1.21+
Bubble Tea (TUI)
Lipgloss (styling)
Glamour (markdown rendering)
```

### Supporting Libraries

```
frontmatter (YAML parsing)
markdown (stats calculation)
yaml (YAML generation)
cobra (CLI)
viper (config)
```

### Storage

```
~/.config/syntax/projects/
└── {project-id}/
    ├── metadata.yaml
    ├── characters/
    ├── locations/
    ├── scenes/
    └── exports/
```

---

## Project Structure

```
syntax/
├── cmd/
│   └── syntax/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── model.go
│   │   ├── nav.go
│   │   └── keys.go
│   ├── ui/
│   │   ├── editor/
│   │   │   ├── editor.go
│   │   │   ├── buffer.go
│   │   │   └── viewport.go
│   │   ├── library/
│   │   │   ├── character_list.go
│   │   │   ├── location_list.go
│   │   │   └── outline_view.go
│   │   ├── components/
│   │   │   ├── textarea.go
│   │   │   ├── modal.go
│   │   │   └── list.go
│   │   └── layout.go
│   ├── story/
│   │   ├── types.go
│   │   └── project.go
│   ├── character/
│   │   ├── types.go
│   │   ├── db.go
│   │   └── render.go
│   ├── location/
│   │   ├── types.go
│   │   ├── db.go
│   │   └── render.go
│   ├── scene/
│   │   ├── types.go
│   │   ├── db.go
│   │   ├── compile.go
│   │   └── stats.go
│   ├── outline/
│   │   ├── types.go
│   │   └── manager.go
│   ├── editor/
│   │   ├── buffer.go
│   │   ├── cursor.go
│   │   └── state.go
│   ├── storage/
│   │   ├── project.go
│   │   ├── character.go
│   │   ├── scene.go
│   │   ├── location.go
│   │   ├── outline.go
│   │   └── config.go
│   ├── export/
│   │   ├── markdown.go
│   │   ├── pdf.go
│   │   ├── docx.go
│   │   ├── html.go
│   │   └── stats_report.go
│   ├── ai/
│   │   ├── assistant.go
│   │   ├── modes.go
│   │   └── prompts.go
│   ├── stats/
│   │   ├── calculator.go
│   │   ├── tracker.go
│   │   └── goals.go
│   ├── theme/
│   │   ├── registry.go
│   │   ├── manager.go
│   │   └── types.go
│   └── config/
│       └── types.go
├── tests/
│   ├── editor_test.go
│   ├── story_test.go
│   ├── character_test.go
│   └── export_test.go
├── go.mod
├── go.sum
├── README.md
├── ARCHITECTURE.md
├── ROADMAP.md
└── Makefile
```

---

## Core Data Types

### Project Type

```go
package story

type Project struct {
    ID              string
    Title           string
    Author          string
    Genre           string
    Status          string  // draft, revising, complete
    TargetWordCount int
    CreatedAt       time.Time
    UpdatedAt       time.Time
    
    // Relationships
    Characters map[string]*character.Character
    Locations  map[string]*location.Location
    Outline    *outline.Outline
    Scenes     map[string]*scene.Scene
}

func (p *Project) NewCharacter(name, role string) (*character.Character, error)
func (p *Project) NewScene(chapter, name string) (*scene.Scene, error)
func (p *Project) WordCount() int
func (p *Project) Export(format string) ([]byte, error)
```

---

### Character Type

```go
package character

type Character struct {
    ID            string
    Name          string
    Aliases       []string
    Role          string      // protagonist, antagonist, etc
    Age           int
    Occupation    string
    Appearance    string
    Background    string
    Arc           string      // Character development
    Relationships map[string]Relationship
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Bio           string      // Markdown content
}

type Relationship struct {
    CharacterID string
    Type        string  // love interest, rival, etc
    Tension     string  // Low, Medium, High
    Notes       string
}
```

---

### Scene Type

```go
package scene

type Scene struct {
    ID           string
    Chapter      int
    SceneNumber  int
    Name         string
    POVCharacter string
    Location     string      // location ID
    TimeOfDay    string      // morning, evening, etc
    PlotPoints   []string
    Status       string      // draft, revising, done
    WordCount    int
    Content      string      // Markdown
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func (s *Scene) UpdateContent(text string) error
func (s *Scene) GetWordCount() int
```

---

### Outline Type

```go
package outline

type Outline struct {
    Structure string  // three-act, hero-journey, custom
    Acts      []Act
}

type Act struct {
    Number int
    Name   string
    Goal   string
    Beats  []Beat
}

type Beat struct {
    ID        string
    Number    int
    Name      string
    Status    string  // todo, active, done
    SceneRef  string
}
```

---

### Stats Type

```go
package stats

type ProjectStats struct {
    TotalWords        int
    TotalSessions     int
    TotalTime         time.Duration
    DaysWithWrites    int
    CurrentStreak     int
    WordsByChapter    map[int]int
    WordsByCharacter  map[string]int
    DailyStats        map[time.Time]int
}

func (s *ProjectStats) CalculateProgress(goal int) float64
func (s *ProjectStats) GetStreak() int
```

---

## Module Breakdown

### 1. editor/ - Text Editing

**Key Functions:**

```go
// Buffer
NewBuffer(initialContent string) *Buffer
func (b *Buffer) Insert(pos int, text string)
func (b *Buffer) Delete(start, end int)
func (b *Buffer) GetContent() string
func (b *Buffer) GetLine(lineNum int) string

// Cursor
func (b *Buffer) CursorUp()
func (b *Buffer) CursorDown()
func (b *Buffer) CursorLeft()
func (b *Buffer) CursorRight()
func (b *Buffer) GetCursorPos() (line, col int)

// Selection
func (b *Buffer) SelectAll()
func (b *Buffer) GetSelection() string

// Undo/Redo
func (b *Buffer) Undo()
func (b *Buffer) Redo()
```

**Tests Required:**
- Insert/delete operations
- Cursor movement
- Selection operations
- Undo/redo
- Large documents (10,000+ lines)

---

### 2. character/ - Character Management

**Key Functions:**

```go
func CreateCharacter(name, role string) *Character
func LoadCharacter(projectDir, characterID string) (*Character, error)
func (c *Character) Save(projectDir string) error
func (c *Character) Delete(projectDir string) error
func GetAllCharacters(projectDir string) ([]*Character, error)
func SearchCharacters(projectDir, query string) ([]*Character, error)
```

---

### 3. scene/ - Scene Organization

**Key Functions:**

```go
func NewScene(chapter, name string) *Scene
func LoadScene(projectDir, sceneID string) (*Scene, error)
func (s *Scene) Save(projectDir string) error
func CompileStory(projectDir, format string) ([]byte, error)
func GetAllScenes(projectDir string) ([]*Scene, error)
func GetScenesByCharacter(projectDir, charID string) ([]*Scene, error)
```

---

### 4. storage/ - File Persistence

**Key Functions:**

```go
func LoadProject(projectID string) (*Project, error)
func SaveProject(p *Project) error
func CreateProject(name, genre string) (*Project, error)
func LoadAllCharacters(projectDir string) (map[string]*Character, error)
func LoadAllScenes(projectDir string) (map[string]*Scene, error)
```

---

### 5. export/ - Output Formats

**Key Functions:**

```go
func ExportMarkdown(projectDir) ([]byte, error)
func ExportPDF(projectDir) ([]byte, error)
func ExportDOCX(projectDir) ([]byte, error)
func ExportHTML(projectDir) ([]byte, error)
func ExportStatsReport(projectDir) (string, error)
```

---

### 6. ai/ - Story Assistant

**Key Functions:**

```go
func (a *Assistant) CheckCharacter(scene, character string) string
func (a *Assistant) PolishDialogue(dialogueText string) string
func (a *Assistant) EvaluateSceneEnergy(sceneText string) string
func (a *Assistant) CheckPlotConsistency(scene *Scene) string
```

---

### 7. stats/ - Progress Tracking

**Key Functions:**

```go
func GetProjectWordCount(projectDir string) int
func GetChapterWordCount(projectDir, chapterNum int) int
func StartSession()
func EndSession() SessionStats
func GetGoalProgress() (current, target int, percent float64)
```

---

## Implementation Phases

### Phase 1: Foundation (Days 1-2)

- [ ] Project structure
- [ ] Theme system
- [ ] Editor buffer and cursor
- [ ] File I/O layer

**Deliverable:** Project compiles and runs

---

### Phase 2: Features (Days 3-5)

- [ ] Split-pane editor UI
- [ ] Character database
- [ ] Scene organization
- [ ] Location database
- [ ] Outline editor
- [ ] Storage layer

**Deliverable:** All features functional

---

### Phase 3: Polish (Days 6-8)

- [ ] AI assistant
- [ ] Stats calculation
- [ ] Export formats
- [ ] Help system
- [ ] Testing
- [ ] Documentation

**Deliverable:** v1.0 release ready

---

## Complete Code Skeleton

### main.go

```go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/kyanite/syntax/internal/app"
)

func main() {
    m := app.NewRootModel()
    p := tea.NewProgram(
        m,
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    )
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

### internal/app/model.go

```go
package app

import tea "github.com/charmbracelet/bubbletea"

type Screen int

const (
    ScreenWelcome Screen = iota
    ScreenEditor
    ScreenLibrary
)

type RootModel struct {
    CurrentScreen Screen
    Width         int
    Height        int
    CurrentProject *story.Project
}

func NewRootModel() RootModel {
    return RootModel{
        CurrentScreen: ScreenWelcome,
    }
}

func (m RootModel) Init() tea.Cmd {
    return tea.EnterAltScreen
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, nil
}

func (m RootModel) View() string {
    switch m.CurrentScreen {
    case ScreenEditor:
        return "Editor"
    default:
        return "Welcome"
    }
}
```

---

## Testing Strategy

### Test Files

```
tests/
├── editor_test.go
├── story_test.go
├── character_test.go
└── export_test.go
```

### Example Test

```go
func TestBufferInsert(t *testing.T) {
    buf := editor.NewBuffer("Hello")
    buf.Insert(5, " World")
    
    if buf.GetContent() != "Hello World" {
        t.Errorf("got %q, want %q", buf.GetContent(), "Hello World")
    }
}
```

---

## Performance Targets

| Operation | Target | Must-Have |
|-----------|--------|-----------|
| Startup | <1s | ✅ |
| Auto-save | <100ms | ✅ |
| Scene switch | <200ms | ✅ |
| Search | <100ms | ✅ |
| Memory idle | <50MB | ✅ |

---

## Validation Checklist

### Before Release

- [ ] All 8 features implemented
- [ ] All acceptance criteria met
- [ ] 0 critical bugs
- [ ] 10 themes working
- [ ] Universal shortcuts implemented
- [ ] Performance targets met
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Works on 80x24 terminal
- [ ] No panics

---

**This is a completely independent, standalone tool. Everything needed is documented here.**

**Next step:** Review README for usage
