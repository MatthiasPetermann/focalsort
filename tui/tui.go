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
)

type logMsg string

type statusMsg struct {
	processed int
	total     int
}

type countersMsg struct {
	success int
	failed  int
	skipped int
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
	logs      []string
	processed int
	total     int
	success   int
	failed    int
	skipped   int

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
		bg      lipgloss.Color
		panel   lipgloss.Color
		title   lipgloss.Color
		text    lipgloss.Color
		muted   lipgloss.Color
		success lipgloss.Color
		error   lipgloss.Color
		accent  lipgloss.Color
	}{
		bg:      lipgloss.Color("#1A103D"),
		panel:   lipgloss.Color("#24124D"),
		title:   lipgloss.Color("#FF5EF1"),
		text:    lipgloss.Color("#E6DCFF"),
		muted:   lipgloss.Color("#A79ACF"),
		success: lipgloss.Color("#00F5D4"),
		error:   lipgloss.Color("#FF4D9D"),
		accent:  lipgloss.Color("#7DF9FF"),
	}

	viewportWidth := maxInt(m.width, 80)
	frameWidth := viewportWidth - 4
	frame := lipgloss.NewStyle().
		Background(palette.bg).
		Foreground(palette.text).
		Padding(1, 2).
		Width(frameWidth)

	titleStyle := lipgloss.NewStyle().Foreground(palette.title).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(palette.muted)
	labelStyle := lipgloss.NewStyle().Foreground(palette.accent).Bold(true)
	okStyle := lipgloss.NewStyle().Foreground(palette.success).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(palette.error).Bold(true)

	panelWidth := maxInt((viewportWidth-12)/2, 36)
	panelStyle := lipgloss.NewStyle().
		Background(palette.panel).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.accent).
		Padding(0, 1).
		Width(panelWidth)

	pct := 0.0
	if m.total > 0 {
		pct = float64(m.processed) / float64(m.total)
	}
	elapsed := time.Since(m.startedAt).Round(time.Second)
	if elapsed < 0 {
		elapsed = 0
	}
	throughput := 0.0
	if elapsed > 0 {
		throughput = float64(m.processed) / elapsed.Seconds()
	}

	progressBar := renderProgressBar(pct, panelWidth-6)

	leftPanel := strings.Join([]string{
		fmt.Sprintf("%s %d/%d", labelStyle.Render("Progress"), m.processed, m.total),
		fmt.Sprintf("%s %5.1f%%", labelStyle.Render("Fortschritt"), pct*100),
		fmt.Sprintf("%s %s", labelStyle.Render("Laufzeit"), elapsed),
		fmt.Sprintf("%s %.2f Bilder/s", labelStyle.Render("Rate"), throughput),
		progressBar,
	}, "\n")

	rightPanel := strings.Join([]string{
		fmt.Sprintf("%s %s", labelStyle.Render("Erfolg"), okStyle.Render(fmt.Sprintf("%d", m.success))),
		fmt.Sprintf("%s %s", labelStyle.Render("Fehler"), errStyle.Render(fmt.Sprintf("%d", m.failed))),
		fmt.Sprintf("%s %d", labelStyle.Render("Übersprungen"), m.skipped),
		fmt.Sprintf("%s %v", labelStyle.Render("Recursive"), m.recursive),
		fmt.Sprintf("%s %v", labelStyle.Render("Dry-run"), m.dryRun),
		fmt.Sprintf("%s %v", labelStyle.Render("Fallback mtime"), m.fallbackModTime),
		fmt.Sprintf("%s %d", labelStyle.Render("Checksum Len"), m.checksumLength),
	}, "\n")

	headline := titleStyle.Render("FocalSort Synthwave")
	help := helpStyle.Render("q: quit, c: logs löschen")

	infoWidth := maxInt(frameWidth-8, 20)
	source := wrapBlock(fmt.Sprintf("%s %s", labelStyle.Render("Import"), m.importFolder), infoWidth)
	currentFile := wrapBlock(fmt.Sprintf("%s %s", labelStyle.Render("Aktuell"), valueOrDash(m.currentFile)), infoWidth)
	lastRename := wrapBlock(fmt.Sprintf("%s %s", labelStyle.Render("Letzte Umbenennung"), valueOrDash(m.lastRename)), infoWidth)

	lastErrorValue := valueOrDash(m.lastError)
	if m.lastError != "" {
		lastErrorValue = errStyle.Render(lastErrorValue)
	}
	lastError := wrapBlock(fmt.Sprintf("%s %s", labelStyle.Render("Letzter Fehler"), lastErrorValue), infoWidth)

	top := lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Render(leftPanel), panelStyle.Render(rightPanel))

	logPanelWidth := maxInt(viewportWidth-10, 30)
	logHeight := maxInt(m.height-18, 8)
	innerLogWidth := maxInt(logPanelWidth-6, 10)
	wrappedLogs := wrapLines(m.logs, innerLogWidth)
	logLines := tailLogs(wrappedLogs, logHeight)
	if len(logLines) == 0 {
		logLines = []string{"Noch keine Logeinträge"}
	}
	logPanel := lipgloss.NewStyle().
		Background(palette.panel).
		Border(lipgloss.NormalBorder()).
		BorderForeground(palette.title).
		Padding(0, 1).
		Width(logPanelWidth).
		Height(logHeight).
		Render(strings.Join(logLines, "\n"))

	content := strings.Join([]string{
		headline,
		help,
		source,
		currentFile,
		lastRename,
		lastError,
		top,
		logPanel,
	}, "\n\n")

	return frame.Render(content)
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
	runErr    chan error
)

func StartTUI() {
	programMu.Lock()
	defer programMu.Unlock()

	if program != nil {
		return
	}

	runErr = make(chan error, 1)
	m := newModel()
	program = tea.NewProgram(m, tea.WithAltScreen())

	go func(p *tea.Program) {
		runErr <- p.Start()
		programMu.Lock()
		defer programMu.Unlock()
		program = nil
	}(program)
}

func StopTUI() {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(stopMsg{})
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

func UpdateCounters(success int, failed int, skipped int) {
	programMu.Lock()
	p := program
	programMu.Unlock()

	if p != nil {
		p.Send(countersMsg{success: success, failed: failed, skipped: skipped})
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

func WaitForExit() error {
	if runErr == nil {
		return nil
	}
	return <-runErr
}
