package main

import (
	"sic-assembler/pass1"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var filePath string

func main() {
	app := tview.NewApplication()

	primaryColor := tcell.ColorPurple
	accentColor := tcell.ColorLightSkyBlue
	backgroundColor := tcell.ColorBlack

	title := tview.NewTextView().
		SetText("[::b][::u]SIC Assembler - TUI Edition[::-]\n").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	title.SetBorderPadding(1, 1, 0, 0)

	outputBox := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetChangedFunc(func() {
			app.Draw()
		})
	outputBox.SetBorder(true).
		SetBorderColor(primaryColor).
		SetTitle(" [Output] ").
		SetTitleAlign(tview.AlignCenter)

	fileInput := tview.NewInputField().
		SetLabel("[skyblue]File Path:[-] ").
		SetFieldWidth(40).
		SetLabelColor(accentColor).
		SetFieldBackgroundColor(backgroundColor).
		SetChangedFunc(func(text string) {
			filePath = text
		})

	runAction := func() {
		if filePath == "" {
			outputBox.SetText("[red]Error:[-] Please specify a file path.")
			return
		}
		result, err := pass1.RunPass1(filePath)
		if err != nil {
			outputBox.SetText("[red]Error")
		}
		outputBox.SetText(result)
		pass1.ClearTables()

	}

	runButton := tview.NewButton("[Run Assembler]").SetSelectedFunc(runAction)

	clearButton := tview.NewButton("[Clear Output]").SetSelectedFunc(func() {
		pass1.ClearTables()
		outputBox.SetText("")

	})

	exitButton := tview.NewButton("[Exit]").SetSelectedFunc(func() {
		pass1.ClearTables()
		app.Stop()
	})

	runButton.SetBackgroundColor(accentColor)
	runButton.SetLabelColor(backgroundColor)

	clearButton.SetBackgroundColor(accentColor)
	clearButton.SetLabelColor(backgroundColor)

	exitButton.SetBackgroundColor(accentColor)
	exitButton.SetLabelColor(backgroundColor)

	buttonsFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(runButton, 0, 1, true).
		AddItem(clearButton, 0, 1, false).
		AddItem(exitButton, 0, 1, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(title, 3, 1, false).
		AddItem(fileInput, 2, 1, true).
		AddItem(buttonsFlex, 1, 1, false).
		AddItem(outputBox, 0, 1, true)

	app.SetRoot(layout, true).SetFocus(fileInput)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:

			if app.GetFocus() == fileInput {
				app.SetFocus(runButton)
			} else if app.GetFocus() == runButton {
				app.SetFocus(clearButton)
			} else if app.GetFocus() == clearButton {
				app.SetFocus(exitButton)
			} else {
				app.SetFocus(fileInput)
			}
			return nil
		default:
			return event
		}
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
