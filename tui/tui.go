package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app *tview.Application
var logView *tview.TextView
var statusView *tview.TextView
var menu *tview.Modal
var exitChan chan struct{}

func StartTUI() {
	app = tview.NewApplication()
	exitChan = make(chan struct{})

	logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { app.Draw() })

	statusView = tview.NewTextView()
	statusView.SetDynamicColors(true)
	statusView.SetTextAlign(tview.AlignCenter)
	statusView.SetBackgroundColor(tcell.ColorBlue)
	statusView.SetText("[::b][white]Processed: 0/0")

	menu = tview.NewModal().
		SetText("Choose an option").
		AddButtons([]string{"Continue", "Abort", "Log"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Abort" {
				close(exitChan)
				app.Stop()
			} else if buttonLabel == "Continue" {
				app.SetRoot(layout(), true)
			} else {
				app.SetRoot(layout(), true)
				LogMessage("Log from Menu")
			}
		})

	app.SetRoot(layout(), true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'm' {
			app.SetRoot(menu, true).SetFocus(menu)
		}
		return event
	})

	go app.Run()
}

func layout() *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(logView, 0, 1, false).
		AddItem(statusView, 1, 1, false)
}

func LogMessage(message string) {
	fmt.Fprintln(logView, message)
	logView.ScrollToEnd()
	app.Draw()
}

func UpdateStatus(processed int, total int) {
	statusView.SetText(fmt.Sprintf("[::b][white]Processed: %d/%d", processed, total))
	app.Draw()
}

func WaitForExit() {
	<-exitChan
}
