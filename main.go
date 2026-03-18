package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"sic-assembler/assembler"
	"sic-assembler/pass1"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ═══════════════════════════════════════════════════════════════════════════════
// MODERN THEME SYSTEM (Catppuccin & Tokyo Night)
// ═══════════════════════════════════════════════════════════════════════════════

type Theme struct {
	Name       string
	Primary    tcell.Color
	Secondary  tcell.Color
	Accent     tcell.Color
	Background tcell.Color
	Surface    tcell.Color
	Text       tcell.Color
	TextMuted  tcell.Color
	Success    tcell.Color
	Warning    tcell.Color
	Error      tcell.Color
}

var themeNames = []string{"catppuccin", "tokyonight", "monokai", "nord"}

var themes = map[string]Theme{
	"catppuccin": {
		Name: "Catppuccin", Primary: tcell.NewRGBColor(137, 180, 250), Secondary: tcell.NewRGBColor(203, 166, 247),
		Accent: tcell.NewRGBColor(148, 226, 213), Background: tcell.NewRGBColor(30, 30, 46),
		Surface: tcell.NewRGBColor(49, 50, 68), Text: tcell.NewRGBColor(205, 214, 244),
		TextMuted: tcell.NewRGBColor(108, 112, 134), Success: tcell.NewRGBColor(166, 227, 161),
		Warning: tcell.NewRGBColor(249, 226, 175), Error: tcell.NewRGBColor(243, 139, 168),
	},
	"tokyonight": {
		Name: "Tokyo Night", Primary: tcell.NewRGBColor(122, 162, 247), Secondary: tcell.NewRGBColor(187, 154, 247),
		Accent: tcell.NewRGBColor(125, 207, 200), Background: tcell.NewRGBColor(26, 27, 38),
		Surface: tcell.NewRGBColor(36, 40, 59), Text: tcell.NewRGBColor(192, 202, 245),
		TextMuted: tcell.NewRGBColor(86, 95, 137), Success: tcell.NewRGBColor(158, 206, 106),
		Warning: tcell.NewRGBColor(224, 175, 104), Error: tcell.NewRGBColor(247, 118, 142),
	},
	"monokai": {
		Name: "Monokai Pro", Primary: tcell.NewRGBColor(252, 152, 103), Secondary: tcell.NewRGBColor(171, 157, 242),
		Accent: tcell.NewRGBColor(120, 220, 232), Background: tcell.NewRGBColor(45, 42, 46),
		Surface: tcell.NewRGBColor(64, 62, 65), Text: tcell.NewRGBColor(252, 252, 250),
		TextMuted: tcell.NewRGBColor(147, 146, 147), Success: tcell.NewRGBColor(169, 220, 118),
		Warning: tcell.NewRGBColor(255, 216, 102), Error: tcell.NewRGBColor(255, 97, 136),
	},
	"nord": {
		Name: "Nord", Primary: tcell.NewRGBColor(136, 192, 208), Secondary: tcell.NewRGBColor(129, 161, 193),
		Accent: tcell.NewRGBColor(143, 188, 187), Background: tcell.NewRGBColor(46, 52, 64),
		Surface: tcell.NewRGBColor(59, 66, 82), Text: tcell.NewRGBColor(236, 239, 244),
		TextMuted: tcell.NewRGBColor(76, 86, 106), Success: tcell.NewRGBColor(163, 190, 140),
		Warning: tcell.NewRGBColor(235, 203, 139), Error: tcell.NewRGBColor(191, 97, 106),
	},
}

type ViewMode int

const (
	ViewOutput ViewMode = iota
	ViewPass1Trace
	ViewPass2Trace
	ViewErrors
)

type AppState struct {
	sync.RWMutex
	FilePath       string
	CurrentTheme   int
	IsProcessing   bool
	History        []HistoryEntry
	LastRunTime    time.Duration
	AnimationFrame int
	LastResult     *pass1.AssemblyResult
	CurrentView    ViewMode
}

type HistoryEntry struct {
	FilePath  string
	Timestamp time.Time
	Success   bool
	Duration  time.Duration
}

var state = &AppState{CurrentTheme: 0, History: make([]HistoryEntry, 0), CurrentView: ViewOutput}

var (
	app         *tview.Application
	pages       *tview.Pages
	outputBox   *tview.TextView
	traceBox    *tview.TextView
	statusBar   *tview.TextView
	historyList *tview.List
	fileInput   *tview.InputField
	mainLayout  *tview.Flex
	headerFlex  *tview.Flex
	tabBarFlex  *tview.Flex
	contentArea *tview.Flex
	runButton   *tview.Button
	clearButton *tview.Button
	themeButton *tview.Button
	helpButton  *tview.Button
	exitButton  *tview.Button
	viewOutBtn  *tview.Button
	viewP1Btn   *tview.Button
	viewP2Btn   *tview.Button
	viewErrBtn  *tview.Button
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var dotFrames = []string{"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}

const (
	scrollLineStep = 1
	scrollPageStep = 15
)

func getTheme() Theme { return themes[themeNames[state.CurrentTheme]] }

func c(theme Theme, ct string) string {
	var color tcell.Color
	switch ct {
	case "p":
		color = theme.Primary
	case "s":
		color = theme.Secondary
	case "a":
		color = theme.Accent
	case "ok":
		color = theme.Success
	case "warn":
		color = theme.Warning
	case "err":
		color = theme.Error
	case "m":
		color = theme.TextMuted
	case "bg":
		color = theme.Background
	case "sf":
		color = theme.Surface
	default:
		color = theme.Text
	}
	r, g, b := color.RGB()
	return fmt.Sprintf("[#%02x%02x%02x]", r, g, b)
}

func progressBar(theme Theme, progress float64, width int) string {
	filled := int(progress * float64(width))
	var bar strings.Builder
	bar.WriteString(c(theme, "m") + "╭─ ")
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString(c(theme, "p") + "━[-]")
		} else {
			bar.WriteString(c(theme, "sf") + "━[-]")
		}
	}
	bar.WriteString(c(theme, "m") + " ─╮[-]")
	return bar.String()
}

func truncStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// ═══════════════════════════════════════════════════════════════════════════════
// SCROLLABLE PANEL SETUP
// ═══════════════════════════════════════════════════════════════════════════════
func makeScrollable(tv *tview.TextView) {
	tv.SetScrollable(true)
	tv.SetWrap(true)
	tv.SetWordWrap(true)
	tv.SetBorderPadding(1, 1, 2, 2) // Breathe room

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := tv.GetScrollOffset()
		switch event.Key() {
		case tcell.KeyUp:
			if row > 0 {
				tv.ScrollTo(row-scrollLineStep, 0)
			}
			return nil
		case tcell.KeyDown:
			tv.ScrollTo(row+scrollLineStep, 0)
			return nil
		case tcell.KeyPgUp:
			nr := row - scrollPageStep
			if nr < 0 {
				nr = 0
			}
			tv.ScrollTo(nr, 0)
			return nil
		case tcell.KeyPgDn:
			tv.ScrollTo(row+scrollPageStep, 0)
			return nil
		case tcell.KeyHome:
			tv.ScrollToBeginning()
			return nil
		case tcell.KeyEnd:
			tv.ScrollToEnd()
			return nil
		}
		return event
	})

	tv.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case tview.MouseScrollUp:
			row, _ := tv.GetScrollOffset()
			nr := row - 3
			if nr < 0 {
				nr = 0
			}
			tv.ScrollTo(nr, 0)
			return action, nil
		case tview.MouseScrollDown:
			row, _ := tv.GetScrollOffset()
			tv.ScrollTo(row+3, 0)
			return action, nil
		}
		return action, event
	})
}

// ═══════════════════════════════════════════════════════════════════════════════
// STYLING & THEMING (FLAT UI)
// ═══════════════════════════════════════════════════════════════════════════════
func applyTheme() {
	theme := getTheme()

	// Main Content Boxes (Subtle Borders)
	outputBox.SetBackgroundColor(theme.Background)
	outputBox.SetTextColor(theme.Text)
	outputBox.SetBorderColor(theme.Surface)

	traceBox.SetBackgroundColor(theme.Surface)
	traceBox.SetTextColor(theme.Text)
	traceBox.SetBorderColor(theme.Surface)

	historyList.SetBackgroundColor(theme.Surface)
	historyList.SetMainTextColor(theme.Text)
	historyList.SetSecondaryTextColor(theme.TextMuted)
	historyList.SetSelectedBackgroundColor(theme.Primary)
	historyList.SetSelectedTextColor(theme.Background)

	// Borderless Input
	fileInput.SetBackgroundColor(theme.Surface)
	fileInput.SetFieldBackgroundColor(theme.Surface)
	fileInput.SetFieldTextColor(theme.Text)
	fileInput.SetLabelColor(theme.Primary)
	fileInput.SetPlaceholderTextColor(theme.TextMuted)

	// Style flat buttons
	styleFlatButton(runButton, theme, "ok")
	styleFlatButton(clearButton, theme, "m")
	styleFlatButton(themeButton, theme, "p")
	styleFlatButton(helpButton, theme, "m")
	styleFlatButton(exitButton, theme, "err")

	mainLayout.SetBackgroundColor(theme.Background)
	headerFlex.SetBackgroundColor(theme.Background)
	tabBarFlex.SetBackgroundColor(theme.Background)
	statusBar.SetBackgroundColor(theme.Background)

	// Update active tab visuals
	updateTabs()
	updateStatus("Theme: "+theme.Name, "info")
}

func styleFlatButton(btn *tview.Button, theme Theme, ct string) {
	var txtColor tcell.Color
	switch ct {
	case "ok":
		txtColor = theme.Success
	case "err":
		txtColor = theme.Error
	case "p":
		txtColor = theme.Primary
	default:
		txtColor = theme.Text
	}

	btn.SetBackgroundColor(theme.Surface)
	btn.SetLabelColor(txtColor)
	btn.SetLabelColorActivated(theme.Background)
	btn.SetBackgroundColorActivated(txtColor)
	btn.SetBorder(false) // Kill box syndrome
}

func updateTabs() {
	theme := getTheme()
	btns := []*tview.Button{viewOutBtn, viewP1Btn, viewP2Btn, viewErrBtn}

	for i, btn := range btns {
		btn.SetBorder(false)
		if ViewMode(i) == state.CurrentView {
			// Active Tab (Highlighted)
			btn.SetBackgroundColor(theme.Primary)
			btn.SetLabelColor(theme.Background)
		} else {
			// Inactive Tab (Muted, blending into background)
			btn.SetBackgroundColor(theme.Background)
			btn.SetLabelColor(theme.TextMuted)
		}
	}
}

func updateStatus(msg string, msgType string) {
	theme := getTheme()
	var icon, color string
	switch msgType {
	case "ok":
		icon = "󰄬"
		color = c(theme, "ok")
	case "warn":
		icon = "󰀪"
		color = c(theme, "warn")
	case "err":
		icon = "󰅙"
		color = c(theme, "err")
	case "load":
		icon = spinnerFrames[state.AnimationFrame%len(spinnerFrames)]
		color = c(theme, "p")
	default:
		icon = "󰋼"
		color = c(theme, "m")
	}
	timeStr := time.Now().Format("15:04")
	vn := [...]string{"Output", "Pass 1", "Pass 2", "Errors"}[state.CurrentView]
	statusBar.SetText(fmt.Sprintf(" %s%s %s[-] %s│[-] %s%s[-] %s│[-] %s%s[-] %s│[-] View: %s %s│[-] ↑↓ PgUp/Dn Tab",
		color, icon, msg, c(theme, "m"), c(theme, "p"), getTheme().Name,
		c(theme, "m"), c(theme, "m"), timeStr, c(theme, "m"), vn, c(theme, "m")))
}

// ═══════════════════════════════════════════════════════════════════════════════
// SLEEK FORMATTERS
// ═══════════════════════════════════════════════════════════════════════════════
func formatAssemblyOutput(result string, filePath string, duration time.Duration) string {
	theme := getTheme()
	p := c(theme, "p")
	m := c(theme, "m")
	ok := c(theme, "ok")
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s╭─ ASSEMBLY COMPLETE ─────────────────────────[-]\n", ok))
	out.WriteString(fmt.Sprintf("%s│[-] File: %s\n", m, filepath.Base(filePath)))
	out.WriteString(fmt.Sprintf("%s│[-] Time: %.3fs\n", m, duration.Seconds()))
	out.WriteString(fmt.Sprintf("%s╰─────────────────────────────────────────────[-]\n\n", m))

	for _, line := range strings.Split(result, "\n") {
		if strings.TrimSpace(line) == "" {
			out.WriteString("\n")
			continue
		}
		if strings.Contains(line, "Symbol Table") {
			out.WriteString(fmt.Sprintf("%s%s[-]\n", p, line))
		} else if strings.HasPrefix(strings.TrimSpace(line), "H^") || strings.HasPrefix(strings.TrimSpace(line), "T^") || strings.HasPrefix(strings.TrimSpace(line), "E^") {
			out.WriteString(fmt.Sprintf("%s%s[-]\n", ok, line))
		} else {
			out.WriteString(fmt.Sprintf("%s%s[-]\n", m, line))
		}
	}
	return out.String()
}

func formatErrorReport(errs []*assembler.AssemblyError, filePath string) string {
	theme := getTheme()
	m := c(theme, "m")
	ec := c(theme, "err")
	w := c(theme, "warn")
	a := c(theme, "a")
	ok := c(theme, "ok")
	text := c(theme, "text") // Using the parsed text color to fix the bug
	var out strings.Builder

	out.WriteString(fmt.Sprintf("%s╭─ ASSEMBLY ERRORS (%d found) ───────────────[-]\n", ec, len(errs)))
	out.WriteString(fmt.Sprintf("%s│[-] File: %s\n", m, filepath.Base(filePath)))
	out.WriteString(fmt.Sprintf("%s╰─────────────────────────────────────────────[-]\n\n", m))

	for i, e := range errs {
		out.WriteString(fmt.Sprintf("%s  ERROR %d/%d [-]\n", ec, i+1, len(errs)))
		out.WriteString(fmt.Sprintf("%s ┃[-] %sLine[-]   %s%d[-]\n", m, a, w, e.Line))
		out.WriteString(fmt.Sprintf("%s ┃[-] %sStmt[-]   %s%s[-]\n", m, a, m, truncStr(e.Statement, 50)))
		out.WriteString(fmt.Sprintf("%s ┃[-] %sType[-]   %s%s[-]\n", m, a, ec, e.Type))
		out.WriteString(fmt.Sprintf("%s ┃[-] %sCause[-]  %s%s[-]\n", m, a, text, e.Cause))
		out.WriteString(fmt.Sprintf("%s ┃[-] %sFix[-]    %s%s[-]\n\n", m, a, ok, e.Fix))
	}
	return out.String()
}

func formatGenericError(err error, filePath string) string {
	theme := getTheme()
	m := c(theme, "m")
	ec := c(theme, "err")
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s╭─ ASSEMBLY FAILED ──────────────────────────[-]\n", ec))
	out.WriteString(fmt.Sprintf("%s│[-] File:  %s\n", m, filepath.Base(filePath)))
	out.WriteString(fmt.Sprintf("%s│[-] Error: %v\n", m, err))
	out.WriteString(fmt.Sprintf("%s╰─────────────────────────────────────────────[-]\n\n", m))
	return out.String()
}

func formatPass1Trace(result *pass1.AssemblyResult) string {
	theme := getTheme()
	p := c(theme, "p")
	s := c(theme, "s")
	a := c(theme, "a")
	m := c(theme, "m")
	ok := c(theme, "ok")
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s╭─ 󰒍 PASS 1 — Define Symbols ────────────────[-]\n", p))
	out.WriteString(fmt.Sprintf("%s│[-] %sScans source, updates LOCCTR, builds SYMTAB.[-]\n", m, m))
	out.WriteString(fmt.Sprintf("%s╰─────────────────────────────────────────────[-]\n\n", p))

	if result == nil || len(result.Pass1Trace) == 0 {
		out.WriteString(fmt.Sprintf(" %s󰋼 No trace data. Run assembler first.[-]\n", m))
		return out.String()
	}

	for _, ev := range result.Pass1Trace {
		switch ev.Kind {
		case assembler.TracePass1Start:
			out.WriteString(fmt.Sprintf(" %s▶[-] %s%s[-]\n\n", ok, s, ev.Detail))
		case assembler.TracePass1End:
			out.WriteString(fmt.Sprintf("\n %s■[-] %s%s[-]\n", ok, s, ev.Detail))
		case assembler.TracePass1Line, assembler.TraceLocctrUpdate:
			lbl := ev.Label
			if lbl == "" {
				lbl = "-"
			}
			out.WriteString(fmt.Sprintf(" %sL%-3d[-] %sLOC=%s[-] %s%-6s[-] %s%-5s[-] %s%s[-]\n", m, ev.Line, p, ev.Locctr, a, lbl, ok, ev.Opcode, m, ev.Operand))
			out.WriteString(fmt.Sprintf("      %s%s[-]\n", m, ev.Detail))
		case assembler.TraceSymbolAdded:
			out.WriteString(fmt.Sprintf("      %s+ SYMTAB[-]: %s%s[-] %s= %s[-]\n", ok, a, ev.Label, p, ev.Locctr))
		}
	}
	if result.SymbolTable != nil && len(result.SymbolTable) > 0 {
		out.WriteString(fmt.Sprintf("\n %s╭─ Final SYMTAB ────────────────────────────[-]\n", m))
		out.WriteString(fmt.Sprintf(" %s│ %s%-10s  %s[-]\n", m, a, "SYMBOL", "ADDRESS"))
		labels := make([]string, 0, len(result.SymbolTable))
		for k := range result.SymbolTable {
			labels = append(labels, k)
		}
		sort.Strings(labels)
		for _, l := range labels {
			out.WriteString(fmt.Sprintf(" %s│[-] %s%-10s[-]  %s%s[-]\n", m, ok, l, p, result.SymbolTable[l]))
		}
		out.WriteString(fmt.Sprintf(" %s╰─────────────────────────────────────────────[-]\n", m))
	}
	return out.String()
}

func formatPass2Trace(result *pass1.AssemblyResult) string {
	theme := getTheme()
	p := c(theme, "p")
	s := c(theme, "s")
	a := c(theme, "a")
	m := c(theme, "m")
	ok := c(theme, "ok")
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s╭─ 󰒍 PASS 2 — Generate Object Code ──────────[-]\n", s))
	out.WriteString(fmt.Sprintf("%s│[-] %sTranslates opcodes, resolves addresses.[-]\n", m, m))
	out.WriteString(fmt.Sprintf("%s╰─────────────────────────────────────────────[-]\n\n", s))

	if result == nil || len(result.Pass2Trace) == 0 {
		out.WriteString(fmt.Sprintf(" %s󰋼 No trace data. Run assembler first.[-]\n", m))
		return out.String()
	}
	for _, ev := range result.Pass2Trace {
		switch ev.Kind {
		case assembler.TracePass2Start:
			out.WriteString(fmt.Sprintf(" %s▶[-] %s%s[-]\n\n", ok, s, ev.Detail))
		case assembler.TracePass2End:
			out.WriteString(fmt.Sprintf("\n %s■[-] %s%s[-]\n", ok, s, ev.Detail))
		case assembler.TraceOpcodeGenerate:
			out.WriteString(fmt.Sprintf(" %sL%-3d[-] %s%-5s[-] %s%-12s[-]\n      %s→ %s[-]\n", m, ev.Line, ok, ev.Opcode, a, ev.Operand, p, ev.Detail))
		case assembler.TraceForwardRef:
			out.WriteString(fmt.Sprintf("      %s↳[-] %s%s[-]\n", s, m, ev.Detail))
		case assembler.TracePass2Line:
			out.WriteString(fmt.Sprintf(" %sL%-3d[-] %s%-5s[-] %s%-12s[-]  %s%s[-]\n", m, ev.Line, ok, ev.Opcode, a, ev.Operand, m, ev.Detail))
		case assembler.TraceObjectRecord:
			out.WriteString(fmt.Sprintf(" %s%s[-]\n", ok, ev.Detail))
		}
	}
	return out.String()
}

// ═══════════════════════════════════════════════════════════════════════════════
// VIEW SWITCHING
// ═══════════════════════════════════════════════════════════════════════════════
func switchView(view ViewMode) {
	state.Lock()
	state.CurrentView = view
	result := state.LastResult
	state.Unlock()
	theme := getTheme()

	updateTabs() // Refresh tab visuals

	switch view {
	case ViewOutput:
		if result != nil && result.Output != "" {
			traceBox.SetText(outputBox.GetText(true))
		} else {
			traceBox.SetText(fmt.Sprintf("\n %s󰋼 Run assembler to see output.[-]", c(theme, "m")))
		}
	case ViewPass1Trace:
		traceBox.SetText(formatPass1Trace(result))
	case ViewPass2Trace:
		traceBox.SetText(formatPass2Trace(result))
	case ViewErrors:
		if result != nil && len(result.Errors) > 0 {
			traceBox.SetText(formatErrorReport(result.Errors, state.FilePath))
		} else {
			traceBox.SetText(fmt.Sprintf("\n %s󰄬 No errors detected.[-]\n\n %sRun assembler to see error analysis.[-]", c(theme, "ok"), c(theme, "m")))
		}
	}
	traceBox.ScrollToBeginning()
	updateStatus("", "info")
}

// ═══════════════════════════════════════════════════════════════════════════════
// ACTIONS
// ═══════════════════════════════════════════════════════════════════════════════
func runAssembler() {
	state.RLock()
	fp := state.FilePath
	state.RUnlock()
	theme := getTheme()
	if fp == "" {
		updateStatus("No file specified", "err")
		outputBox.SetText(fmt.Sprintf("\n %s󰅙 Please enter a file path above", c(theme, "warn")))
		return
	}
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		updateStatus("File not found", "err")
		outputBox.SetText(formatGenericError(fmt.Errorf("file does not exist: %s", fp), fp))
		return
	}
	state.Lock()
	state.IsProcessing = true
	state.Unlock()

	go func() {
		for i := 0; ; i++ {
			state.RLock()
			pr := state.IsProcessing
			state.RUnlock()
			if !pr {
				break
			}
			state.Lock()
			state.AnimationFrame = i
			state.Unlock()
			progress := float64(i%30) / 30.0
			spinner := spinnerFrames[i%len(spinnerFrames)]
			th := getTheme()
			app.QueueUpdateDraw(func() {
				updateStatus("Assembling...", "load")
				var lt strings.Builder
				lt.WriteString(fmt.Sprintf("\n\n\n %s%s[-] %sProcessing %s[-]\n\n", c(th, "p"), spinner, c(th, "m"), filepath.Base(fp)))
				lt.WriteString(fmt.Sprintf(" %s\n\n", progressBar(th, progress, 40)))
				outputBox.SetText(lt.String())
			})
			time.Sleep(60 * time.Millisecond)
		}
	}()

	go func() {
		st := time.Now()
		result, err := pass1.RunPass1WithTrace(fp)
		dur := time.Since(st)
		if dur < 500*time.Millisecond {
			time.Sleep(500*time.Millisecond - dur)
		}
		state.Lock()
		state.IsProcessing = false
		state.LastRunTime = dur
		state.LastResult = result
		entry := HistoryEntry{FilePath: fp, Timestamp: time.Now(), Success: err == nil, Duration: dur}
		state.History = append([]HistoryEntry{entry}, state.History...)
		if len(state.History) > 8 {
			state.History = state.History[:8]
		}
		state.Unlock()
		app.QueueUpdateDraw(func() {
			if err != nil {
				if me, ok := err.(*pass1.MultiAssemblyError); ok {
					updateStatus(fmt.Sprintf("Failed: %d error(s)", len(me.Errors)), "err")
					outputBox.SetText(formatErrorReport(me.Errors, fp))
					state.Lock()
					state.CurrentView = ViewErrors
					state.Unlock()
					switchView(ViewErrors)
				} else {
					updateStatus("Failed", "err")
					outputBox.SetText(formatGenericError(err, fp))
				}
			} else {
				updateStatus(fmt.Sprintf("Done in %.2fs", dur.Seconds()), "ok")
				outputBox.SetText(formatAssemblyOutput(result.Output, fp, dur))
				state.Lock()
				state.CurrentView = ViewOutput
				state.Unlock()
				switchView(ViewOutput)
			}
			outputBox.ScrollToBeginning()
			refreshHistory()
		})
	}()
}

func refreshHistory() {
	theme := getTheme()
	historyList.Clear()
	for _, e := range state.History {
		var ic, tc string
		if e.Success {
			ic = c(theme, "ok") + "󰄬[-]"
			tc = c(theme, "m")
		} else {
			ic = c(theme, "err") + "󰅙[-]"
			tc = c(theme, "err")
		}
		historyList.AddItem(fmt.Sprintf("%s %s", ic, filepath.Base(e.FilePath)),
			fmt.Sprintf("   %s%s · %.2fs[-]", tc, e.Timestamp.Format("15:04:05"), e.Duration.Seconds()), 0, nil)
	}
}

func clearOutput() {
	theme := getTheme()
	outputBox.SetText(fmt.Sprintf("\n %s󰋼 Output cleared[-]", c(theme, "m")))
	traceBox.SetText(fmt.Sprintf("\n %s󰋼 Trace cleared[-]", c(theme, "m")))
	state.Lock()
	state.LastResult = nil
	state.Unlock()
	pass1.ClearTables()
	updateStatus("Cleared", "info")
}

func cycleTheme() {
	state.Lock()
	state.CurrentTheme = (state.CurrentTheme + 1) % len(themeNames)
	state.Unlock()
	applyTheme()
	refreshHistory()
	theme := getTheme()
	var pv strings.Builder
	pv.WriteString(fmt.Sprintf("\n\n %s󰏘 Theme: %s%s[-]\n\n", c(theme, "p"), c(theme, "Text"), theme.Name))
	for _, pair := range [][2]string{{"p", "Primary"}, {"s", "Secondary"}, {"a", "Accent"}, {"ok", "Success"}, {"err", "Error"}, {"bg", "Background"}, {"sf", "Surface"}} {
		pv.WriteString(fmt.Sprintf(" %s██████[-] %s\n", c(theme, pair[0]), pair[1]))
	}
	outputBox.SetText(pv.String())
}

func showWelcome() {
	theme := getTheme()
	p := c(theme, "p")
	m := c(theme, "m")
	a := c(theme, "a")
	outputBox.SetText(fmt.Sprintf(`

 %s 󰜎 S I C   A S S E M B L E R   v3.0[-]

 %s ────────────────────────────────────────[-]

 %s 1.[-] Enter path to %s.asm[-] file above
 %s 2.[-] Press %s󰐊 run[-] or %sCtrl+R[-]
 %s 3.[-] Use tabs to view Output & Traces

 %s ────────────────────────────────────────[-]
 %s Ctrl+H[-] help  %s·[-]  %sCtrl+T[-] theme  %s·[-]  %sCtrl+Q[-] quit

`, p, m, m, a, m, p, p, m, m, m, m, m, m, m))
}

// ═══════════════════════════════════════════════════════════════════════════════
// MAIN & LAYOUT
// ═══════════════════════════════════════════════════════════════════════════════
func main() {
	app = tview.NewApplication()
	app.EnableMouse(true)

	theme := getTheme()
	pages = tview.NewPages()

	outputBox = tview.NewTextView().SetDynamicColors(true)
	makeScrollable(outputBox)
	outputBox.SetBorder(true)

	traceBox = tview.NewTextView().SetDynamicColors(true)
	makeScrollable(traceBox)

	historyList = tview.NewList().ShowSecondaryText(true)
	historyList.SetBorder(false) // Clean look

	// Borderless Input with custom label
	fileInput = tview.NewInputField().
		SetLabel(" 󰈔 File: ").
		SetFieldWidth(40).
		SetPlaceholder("path/to/source.asm").
		SetChangedFunc(func(t string) { state.Lock(); state.FilePath = t; state.Unlock() })

	statusBar = tview.NewTextView().SetDynamicColors(true)

	// Buttons
	runButton = tview.NewButton(" 󰐊 Run ")
	clearButton = tview.NewButton(" 󰎚 Clear ")
	themeButton = tview.NewButton(" 󰏘 Theme ")
	helpButton = tview.NewButton(" 󰋖 ")
	exitButton = tview.NewButton(" 󰗼 ")

	runButton.SetSelectedFunc(runAssembler)
	clearButton.SetSelectedFunc(clearOutput)
	themeButton.SetSelectedFunc(cycleTheme)
	exitButton.SetSelectedFunc(func() { pass1.ClearTables(); app.Stop() })

	// Tab Buttons
	viewOutBtn = tview.NewButton(" 󰈔 Output ")
	viewP1Btn = tview.NewButton(" 󰒍 Pass 1 ")
	viewP2Btn = tview.NewButton(" 󰒍 Pass 2 ")
	viewErrBtn = tview.NewButton(" 󰅙 Errors ")

	viewOutBtn.SetSelectedFunc(func() { switchView(ViewOutput) })
	viewP1Btn.SetSelectedFunc(func() { switchView(ViewPass1Trace) })
	viewP2Btn.SetSelectedFunc(func() { switchView(ViewPass2Trace) })
	viewErrBtn.SetSelectedFunc(func() { switchView(ViewErrors) })

	// ─── UNIFIED HEADER (No borders, inline) ───
	headerFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(fileInput, 0, 1, true).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 2, 0, false).
		AddItem(runButton, 9, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false).
		AddItem(clearButton, 11, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false).
		AddItem(themeButton, 11, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false).
		AddItem(helpButton, 5, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false).
		AddItem(exitButton, 5, 0, false)

	// ─── TAB BAR ───
	tabBarFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(viewOutBtn, 12, 0, false).
		AddItem(viewP1Btn, 12, 0, false).
		AddItem(viewP2Btn, 12, 0, false).
		AddItem(viewErrBtn, 12, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 0, 1, false)

	// ─── CONTENT AREA ───
	// Right panel integrates tracebox and a clean history list at bottom
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(traceBox, 0, 4, false). // Increased flex proportion to give traceBox more room
		AddItem(tview.NewTextView().SetText(" 󰋚 History").SetTextColor(theme.TextMuted).SetBackgroundColor(theme.Surface), 1, 0, false).
		AddItem(historyList, 8, 0, false) // Fixed height for history

	contentArea = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(outputBox, 0, 1, false).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false). // Gap
		AddItem(rightPanel, 0, 1, false)

	// ─── MASTER LAYOUT ───
	mainLayout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false). // Top padding
		AddItem(headerFlex, 1, 0, true).
		AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false). // Gap
		AddItem(tabBarFlex, 1, 0, false).
		AddItem(contentArea, 0, 1, false).
		AddItem(statusBar, 1, 0, false)

	pages.AddPage("main", mainLayout, true, true)
	applyTheme()
	showWelcome()

	focusables := []tview.Primitive{fileInput, runButton, clearButton, themeButton, helpButton, exitButton, viewOutBtn, viewP1Btn, viewP2Btn, viewErrBtn, outputBox, traceBox}
	fi := 0

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			fi = (fi + 1) % len(focusables)
			app.SetFocus(focusables[fi])
			return nil
		case tcell.KeyBacktab:
			fi = (fi - 1 + len(focusables)) % len(focusables)
			app.SetFocus(focusables[fi])
			return nil
		case tcell.KeyCtrlR:
			runAssembler()
			return nil
		case tcell.KeyCtrlL:
			clearOutput()
			return nil
		case tcell.KeyCtrlT:
			cycleTheme()
			return nil
		case tcell.KeyCtrlQ:
			pass1.ClearTables()
			app.Stop()
			return nil
		case tcell.KeyEscape:
			app.SetFocus(fileInput)
			fi = 0
			return nil
		case tcell.KeyF1:
			switchView(ViewOutput)
			return nil
		case tcell.KeyF2:
			switchView(ViewPass1Trace)
			return nil
		case tcell.KeyF3:
			switchView(ViewPass2Trace)
			return nil
		case tcell.KeyF4:
			switchView(ViewErrors)
			return nil
		}
		return event
	})

	app.SetRoot(pages, true).SetFocus(fileInput)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
