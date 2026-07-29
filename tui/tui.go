package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type logMsg string

type statusMsg struct {
	processed int
	total     int
}

type countersMsg struct {
	success      int
	failed       int
	skipped      int
	alreadyNamed int
}

type configMsg struct {
	importFolder    string
	recursive       bool
	dryRun          bool
	fallbackModTime bool
	checksumLength  int
}

type currentFileMsg string

type renameMsg struct {
	from string
	to   string
}

type errorMsg string

type stopMsg struct{}

type model struct {
	logs         []string
	processed    int
	total        int
	success      int
	failed       int
	skipped      int
	alreadyNamed int

	importFolder    string
	recursive       bool
	dryRun          bool
	fallbackModTime bool
	checksumLength  int
	currentFile     string
	lastRename      string
	lastError       string

	startedAt time.Time
	width     int
	height    int
	quitting  bool
}

func newModel() model {
	return model{
		logs:      make([]string, 0),
		startedAt: time.Now(),
		width:     100,
		height:    32,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case logMsg:
		m.logs = append(m.logs, string(message))
		if len(m.logs) > 1000 {
			m.logs = m.logs[len(m.logs)-1000:]
		}
		return m, nil
	case statusMsg:
		m.processed = message.processed
		m.total = message.total
		return m, nil
	case countersMsg:
		m.success = message.success
		m.failed = message.failed
		m.skipped = message.skipped
		m.alreadyNamed = message.alreadyNamed
		return m, nil
	case configMsg:
		m.importFolder = message.importFolder
		m.recursive = message.recursive
		m.dryRun = message.dryRun
		m.fallbackModTime = message.fallbackModTime
		m.checksumLength = message.checksumLength
		return m, nil
	case currentFileMsg:
		m.currentFile = string(message)
		return m, nil
	case renameMsg:
		m.lastRename = fmt.Sprintf("%s -> %s", filepath.Base(message.from), filepath.Base(message.to))
		return m, nil
	case errorMsg:
		m.lastError = string(message)
		return m, nil
	case stopMsg:
		m.quitting = true
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		return m, nil
	case tea.KeyMsg:
		switch message.String() {
		case "q", "ctrl+c":
			cancelProcessing()
			m.quitting = true
			return m, tea.Quit
		case "c":
			m.logs = m.logs[:0]
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	palette := struct {
		bg     lipgloss.Color
		panel  lipgloss.Color
		title  lipgloss.Color
		text   lipgloss.Color
		accent lipgloss.Color
	}{
		bg:     lipgloss.Color("#1A103D"),
		panel:  lipgloss.Color("#24124D"),
		title:  lipgloss.Color("#FF5EF1"),
		text:   lipgloss.Color("#E6DCFF"),
		accent: lipgloss.Color("#7DF9FF"),
	}

	viewportWidth := maxInt(m.width, 1)
	viewportHeight := maxInt(m.height, 1)
	frame := lipgloss.NewStyle().
		Background(palette.bg).
		Foreground(palette.text).
		Width(viewportWidth).
		Height(viewportHeight)

	titleStyle := lipgloss.NewStyle().Foreground(palette.title).Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(palette.accent)
	panelStyle := lipgloss.NewStyle().Background(palette.panel)

	pct := 0.0
	if m.total > 0 {
		pct = float64(m.processed) / float64(m.total)
	}
	elapsed := time.Since(m.startedAt).Round(time.Second)
	if elapsed < 0 {
		elapsed = 0
	}
	lines := []string{titleStyle.Render(padText("FocalSort | q: Cancel | c: Clear logs", viewportWidth))}
	usedHeight := 1
	if viewportHeight >= 14 {
		lines = append(lines,
			padText("Import: "+valueOrDash(m.importFolder), viewportWidth),
			padText("Current: "+valueOrDash(m.currentFile), viewportWidth),
			padText(activityText(m), viewportWidth),
		)
		usedHeight += 3
	} else if viewportHeight >= 9 {
		lines = append(lines, padText("Current: "+valueOrDash(m.currentFile), viewportWidth))
		usedHeight++
	}

	if viewportHeight >= 8 {
		if viewportWidth >= 70 {
			leftWidth := (viewportWidth - 1) / 2
			rightWidth := viewportWidth - leftWidth - 1
			left := box(leftWidth, []string{
				fmt.Sprintf("Progress: %d/%d (%.1f%%)", m.processed, m.total, pct*100),
				renderProgressBar(pct, maxInt(leftWidth-4, 1)),
				"Elapsed: " + elapsed.String(),
			})
			right := box(rightWidth, []string{
				fmt.Sprintf("Success: %d | Failed: %d", m.success, m.failed),
				fmt.Sprintf("Skipped: %d | Already named: %d", m.skipped, m.alreadyNamed),
				fmt.Sprintf("recursive=%t | dry-run=%t", m.recursive, m.dryRun),
			})
			for i := range left {
				lines = append(lines, borderStyle.Render(left[i])+" "+borderStyle.Render(right[i]))
			}
		} else {
			for _, line := range box(viewportWidth, []string{
				fmt.Sprintf("Progress: %d/%d (%.1f%%)", m.processed, m.total, pct*100),
				fmt.Sprintf("Success: %d | Failed: %d | Skipped: %d", m.success, m.failed, m.skipped),
				renderProgressBar(pct, maxInt(viewportWidth-4, 1)),
			}) {
				lines = append(lines, borderStyle.Render(line))
			}
		}
		usedHeight += 5
	}

	logHeight := viewportHeight - usedHeight
	if logHeight >= 3 {
		innerLogWidth := maxInt(viewportWidth-2, 1)
		logLines := tailLogs(wrapLines(m.logs, innerLogWidth), maxInt(logHeight-2, 1))
		if len(logLines) == 0 {
			logLines = []string{"No log entries yet"}
		}
		for len(logLines) < logHeight-2 {
			logLines = append(logLines, "")
		}
		for _, line := range box(viewportWidth, logLines) {
			lines = append(lines, panelStyle.Render(line))
		}
	}

	return frame.Render(strings.Join(lines, "\n"))
}

func box(width int, content []string) []string {
	width = maxInt(width, 2)
	innerWidth := width - 2
	lines := make([]string, 0, len(content)+2)
	lines = append(lines, "╭"+strings.Repeat("─", innerWidth)+"╮")
	for _, line := range content {
		lines = append(lines, "│"+padText(line, innerWidth)+"│")
	}
	lines = append(lines, "╰"+strings.Repeat("─", innerWidth)+"╯")
	return lines
}

func padText(text string, width int) string {
	text = truncateText(text, width)
	return text + strings.Repeat(" ", maxInt(width-runewidth.StringWidth(text), 0))
}

func activityText(m model) string {
	if m.lastError != "" {
		return "Last error: " + m.lastError
	}
	return "Last rename: " + valueOrDash(m.lastRename)
}

func truncateText(text string, width int) string {
	return runewidth.Truncate(text, maxInt(width, 1), "...")
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func wrapBlock(text string, width int) string {
	return strings.Join(hardWrap(text, width), "\n")
}

func wrapLines(lines []string, width int) []string {
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, hardWrap(line, width)...)
	}
	return wrapped
}

func hardWrap(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	if text == "" {
		return []string{""}
	}

	parts := strings.Split(text, "\n")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			result = append(result, "")
			continue
		}
		current := ""
		for _, r := range part {
			if utf8.RuneCountInString(current) >= width {
				result = append(result, current)
				current = ""
			}
			current += string(r)
		}
		if current != "" {
			result = append(result, current)
		}
	}
	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func tailLogs(logs []string, limit int) []string {
	if limit < 1 {
		limit = 1
	}
	if len(logs) <= limit {
		return logs
	}
	return logs[len(logs)-limit:]
}

func renderProgressBar(progress float64, width int) string {
	if width < 8 {
		width = 8
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	programMu sync.Mutex
	program   *tea.Program
	cancelled chan struct{}
	runDone   chan struct{}
)

func StartTUI() {
	programMu.Lock()
	defer programMu.Unlock()

	if program != nil {
		return
	}

	cancelled = make(chan struct{})
	runDone = make(chan struct{})
	m := newModel()
	program = tea.NewProgram(m, tea.WithAltScreen())

	go func(p *tea.Program, done chan struct{}) {
		_, _ = p.Run()
		programMu.Lock()
		if program == p {
			program = nil
		}
		close(done)
		programMu.Unlock()
	}(program, runDone)
}

func StopTUI() {
	programMu.Lock()
	p := program
	done := runDone
	programMu.Unlock()

	if p != nil {
		p.Quit()
	}
	if done != nil {
		<-done
	}
}

// Cancelled reports whether the user exited the TUI and requested processing stop.
func Cancelled() bool {
	programMu.Lock()
	cancel := cancelled
	programMu.Unlock()

	if cancel == nil {
		return false
	}
	select {
	case <-cancel:
		return true
	default:
		return false
	}
}

func cancelProcessing() {
	programMu.Lock()
	defer programMu.Unlock()

	if cancelled == nil {
		return
	}
	select {
	case <-cancelled:
	default:
		close(cancelled)
	}
}

func LogMessage(message string) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(logMsg(message))
	}
}

func SetRunContext(importFolder string, recursive bool, dryRun bool, fallbackToModTime bool, checksumLength int) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(configMsg{
			importFolder:    importFolder,
			recursive:       recursive,
			dryRun:          dryRun,
			fallbackModTime: fallbackToModTime,
			checksumLength:  checksumLength,
		})
	}
}

func SetCurrentFile(path string) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(currentFileMsg(filepath.Base(path)))
	}
}

func SetLastRename(from string, to string) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(renameMsg{from: from, to: to})
	}
}

func SetLastError(err string) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(errorMsg(err))
	}
}

func UpdateCounters(success int, failed int, skipped int, alreadyNamed int) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(countersMsg{success: success, failed: failed, skipped: skipped, alreadyNamed: alreadyNamed})
	}
}

func UpdateStatus(processed int, total int) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(statusMsg{processed: processed, total: total})
	}
}
