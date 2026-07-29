package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelUpdateAndView(t *testing.T) {
	t.Parallel()

	m := newModel()

	updated, cmd := m.Update(logMsg("first"))
	if cmd != nil {
		t.Fatalf("expected nil command after log update")
	}
	m = updated.(model)

	updated, cmd = m.Update(countersMsg{success: 2, failed: 1, skipped: 3, alreadyNamed: 4})
	if cmd != nil {
		t.Fatalf("expected nil command after counter update")
	}
	m = updated.(model)

	updated, cmd = m.Update(statusMsg{processed: 3, total: 10})
	if cmd != nil {
		t.Fatalf("expected nil command after status update")
	}
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "FocalSort Synthwave") {
		t.Fatalf("missing title in view: %q", view)
	}
	if !strings.Contains(view, "Progress") {
		t.Fatalf("missing progress in view: %q", view)
	}
	if !strings.Contains(view, "first") {
		t.Fatalf("missing log entry in view: %q", view)
	}
	if !strings.Contains(view, "█") {
		t.Fatalf("missing progress bar in view: %q", view)
	}
	if !strings.Contains(view, "Bereits benannt") {
		t.Fatalf("missing already named counter in view: %q", view)
	}
}

func TestModelQuitAndClearMessages(t *testing.T) {
	t.Parallel()

	m := newModel()
	updated, _ := m.Update(logMsg("one"))
	m = updated.(model)
	updated, _ = m.Update(logMsg("two"))
	m = updated.(model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd != nil {
		t.Fatalf("expected nil command for clear logs")
	}
	m = updated.(model)
	if len(m.logs) != 0 {
		t.Fatalf("expected logs cleared, got %d", len(m.logs))
	}

	updated, cmd = m.Update(stopMsg{})
	if cmd == nil {
		t.Fatalf("expected quit command for stopMsg")
	}
	m = updated.(model)
	if !m.quitting {
		t.Fatalf("expected quitting state after stopMsg")
	}

	m = newModel()
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected quit command for q key")
	}
	m = updated.(model)
	if !m.quitting {
		t.Fatalf("expected quitting state after q key")
	}

	m = newModel()
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command for ctrl+c")
	}
	m = updated.(model)
	if !m.quitting {
		t.Fatalf("expected quitting state after ctrl+c")
	}
}

func TestRenderProgressBarBounds(t *testing.T) {
	t.Parallel()

	bar := renderProgressBar(1.5, 8)
	if !strings.Contains(bar, "████████") {
		t.Fatalf("expected full bar, got %q", bar)
	}

	bar = renderProgressBar(-1, 8)
	if !strings.Contains(bar, "░░░░░░░░") {
		t.Fatalf("expected empty bar, got %q", bar)
	}
}

func TestHardWrap(t *testing.T) {
	t.Parallel()

	wrapped := hardWrap("abcdef", 3)
	if len(wrapped) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(wrapped))
	}
	if wrapped[0] != "abc" || wrapped[1] != "def" {
		t.Fatalf("unexpected wrapped output: %#v", wrapped)
	}
}
