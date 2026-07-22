package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"wslc-tui-ms/internal/commands"
	"wslc-tui-ms/internal/data"
	"wslc-tui-ms/internal/ui"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewCommands view = iota
	viewPreview
	viewForm
	viewConfirm
	viewOutput
	viewLearn
)

type model struct {
	width  int
	height int

	// Navigation
	currentView   view
	activeTab     int
	sidebarIndex  int
	selectedIndex int

	// Command state
	categories   []string
	allCommands  map[string][]commands.Command
	filteredCmds []commands.Command
	inputValue   string
	inputFocused bool
	tabIndex     int

	// Preview
	previewCmd *commands.Command

	// Command form state and values remembered for this app session.
	form             *formState
	formCommand      *commands.Command
	formOptionMemory map[string]map[string]string
	pendingForm      *commands.BuildResult
	pendingArgs      []string

	// Placeholder editing (in preview)
	placeholders  []string
	phValues      map[string]string
	phActiveIndex int // -1 = no field focused
	phInput       textinput.Model
	phWarn        bool // true when Enter blocked due to empty placeholders

	// Execution
	running           bool
	cancelFn          context.CancelFunc
	executionID       int
	outputCmd         string
	outputDifficulty  string
	outputArgs        []string
	pendingCommand    string
	pendingDifficulty string

	// Output
	outputResult    *commands.ExecutionResult
	viewportContent string

	// Learn
	learnTopics       []string
	learnIndex        int
	learnDetailActive bool
	learnContent      string

	// Components
	textInput      textinput.Model
	viewport       viewport.Model
	formInput      textinput.Model
	formViewport   *viewport.Model
	formFieldLines map[int]int

	// Status
	lastCopied     string
	showTooltip    bool
	tooltipContent string

	// Clickable regions (header tabs, footer hints).
	// Pointer so mutations during value-receiver View()/render calls persist.
	clickRegions *[]clickRegion
}

type clickRegion struct {
	x1, x2 int // inclusive column range (terminal columns, 0-indexed)
	y1, y2 int // inclusive row range (terminal rows, 0-indexed)
	action string
}

type execDoneMsg struct {
	result commands.ExecutionResult
	id     int
}

func NewModelForTest(width, height int) model {
	m := NewModel()
	m.width = width
	m.height = height
	// simulate filtered commands for category 0 (Container)
	m.updateFiltered()
	return m
}

func NewModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	phi := textinput.New()
	phi.CharLimit = 200
	phi.Width = 40

	vp := viewport.New(0, 0)
	formInput := textinput.New()
	formInput.CharLimit = 200
	formInput.Width = 36
	formViewport := viewport.New(0, 0)

	categories := data.GetCategories()
	allCmds := data.AllCommands

	var initialCmds []commands.Command
	for _, cat := range categories {
		initialCmds = append(initialCmds, allCmds[cat]...)
	}

	m := model{
		currentView:    viewCommands,
		inputFocused:   true,
		categories:     categories,
		allCommands:    allCmds,
		filteredCmds:   initialCmds,
		textInput:      ti,
		viewport:       vp,
		formInput:      formInput,
		formViewport:   &formViewport,
		formFieldLines: make(map[int]int),
		phInput:        phi,
		phActiveIndex:  -1,
		learnTopics: []string{
			"Getting Started",
			"Container Operations",
			"Image Management",
			"Network & Volume",
			"Sessions",
			"System & Maintenance",
		},
	}
	m.clickRegions = &[]clickRegion{}
	m.updateFiltered()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		_, contentHeight := m.calculateLayout()
		// The output viewport is drawn inside CardStyle (1-cell border on each
		// side) which also carries a header line plus a blank separator line.
		// Account for those so the viewport's scrollable area matches what is
		// actually visible; otherwise its bottom is unreachable.
		vpWidth := m.width - 4
		if vpWidth < 1 {
			vpWidth = 1
		}
		vpHeight := contentHeight - 4
		if vpHeight < 1 {
			vpHeight = 1
		}
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
		m.formViewport.Width = vpWidth
		m.formViewport.Height = vpHeight
		return m, nil

	case execDoneMsg:
		if msg.id != m.executionID {
			return m, nil
		}
		m.running = false
		m.outputResult = &msg.result
		m.setViewportOutput()
		m.viewport.GotoTop()
		return m, nil

	case tea.KeyMsg:
		if m.running {
			switch msg.String() {
			case "esc":
				if m.cancelFn != nil {
					m.cancelFn()
					m.cancelFn = nil
				}
				m.executionID++
				m.running = false
				m.currentView = viewCommands
				m.outputResult = nil
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}

	return m, nil
}

func (m *model) setViewportOutput() {
	if m.outputResult == nil {
		m.viewport.SetContent("")
		m.viewportContent = ""
		return
	}

	var lines []string

	if m.outputResult.Error != nil && m.outputResult.ExitCode != 0 {
		lines = append(lines, ui.OutputErrorStyle.Render("  ✗ Error (exit code "+fmt.Sprintf("%d", m.outputResult.ExitCode)+")"))
		lines = append(lines, "")
	}

	output := strings.TrimSpace(m.outputResult.Output)

	if output == "" {
		if m.outputResult.Error != nil {
			lines = append(lines, ui.OutputErrorStyle.Render(fmt.Sprintf("Command failed: %s", m.outputResult.Error.Error())))
		} else {
			lines = append(lines, ui.OutputSuccessStyle.Render("  ✓ Command completed successfully (no output)"))
		}
	} else if strings.HasPrefix(output, "{") || strings.HasPrefix(output, "[") {
		prettyJSON := formatJSON(output)
		for _, line := range strings.Split(prettyJSON, "\n") {
			lines = append(lines, highlightJSONLine(line))
		}
	} else {
		lines = append(lines, output)
	}

	lines = append(lines, "")
	lines = append(lines, ui.ScrollStyle.Render(fmt.Sprintf("  Duration: %s", m.outputResult.Duration.Round(time.Millisecond))))

	if m.lastCopied != "" {
		lines = append(lines, "")
		lines = append(lines, ui.OutputSuccessStyle.Render("  ✓ Copied "+m.lastCopied+" to clipboard"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	m.viewport.SetContent(content)
	m.viewportContent = content
}

func (m model) isKnownCommand(s string) bool {
	for _, cmds := range m.allCommands {
		for _, c := range cmds {
			if c.Full == s {
				return true
			}
		}
	}
	return false
}

func (m model) executeCommand(command string) (tea.Model, tea.Cmd) {
	return m.executeCommandWithArgs(command, nil)
}

func (m model) executeCommandWithArgs(command string, structuredArgs []string) (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFn = cancel
	m.running = true
	m.executionID++
	m.outputCmd = command
	m.outputArgs = append([]string(nil), structuredArgs...)
	m.outputResult = nil
	m.lastCopied = ""
	m.currentView = viewOutput

	execID := m.executionID

	cmd := func() tea.Msg {
		start := time.Now()
		args := structuredArgs
		if args == nil {
			args = commands.ParseCommand(command)
		}
		var execCmd *exec.Cmd
		if len(args) > 0 {
			execCmd = exec.CommandContext(ctx, args[0], args[1:]...)
		} else {
			execCmd = exec.CommandContext(ctx, "cmd", "/C", command)
		}

		var stdout, stderr bytes.Buffer
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr

		if err := execCmd.Start(); err != nil {
			return execDoneMsg{
				id: execID,
				result: commands.ExecutionResult{
					Error:    err,
					ExitCode: -1,
					Duration: time.Since(start),
				},
			}
		}

		go func() {
			<-ctx.Done()
			if execCmd.Process != nil {
				if runtime.GOOS == "windows" {
					exec.Command("taskkill", "/F", "/T", "/PID",
						strconv.Itoa(execCmd.Process.Pid)).Run()
				} else {
					execCmd.Process.Kill()
				}
			}
		}()

		err := execCmd.Wait()
		duration := time.Since(start)

		output := stdout.String()
		if stderr.Len() > 0 {
			if output != "" {
				output += "\n"
			}
			output += stderr.String()
		}

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		return execDoneMsg{
			id: execID,
			result: commands.ExecutionResult{
				Output:   output,
				Error:    err,
				ExitCode: exitCode,
				Duration: duration,
			},
		}
	}

	return m, tea.Batch(tea.Sequence(tea.Println(""), cmd))
}

func (m model) executeOrConfirm(command, difficulty string, structuredArgs []string) (tea.Model, tea.Cmd) {
	if difficulty == "intermediate" || difficulty == "advanced" {
		m.pendingCommand = command
		m.pendingArgs = append([]string(nil), structuredArgs...)
		m.pendingDifficulty = difficulty
		m.currentView = viewConfirm
		return m, nil
	}
	return m.executeCommandWithDifficulty(command, difficulty, structuredArgs)
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !(m.currentView == viewCommands && m.inputFocused) {
			return m, tea.Quit
		}
	}

	switch m.currentView {
	case viewCommands:
		return m.handleCommandsKey(msg)
	case viewPreview:
		return m.handlePreviewKey(msg)
	case viewForm:
		return m.handleFormKey(msg)
	case viewConfirm:
		return m.handleConfirmationKey(msg)
	case viewOutput:
		return m.handleOutputKey(msg)
	case viewLearn:
		return m.handleLearnKey(msg)
	}

	return m, nil
}

func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Mouse wheel scrolls the output viewport.
	if m.currentView == viewOutput && !m.running {
		switch msg.Type {
		case tea.MouseWheelUp:
			m.viewport.ScrollUp(3)
			return m, nil
		case tea.MouseWheelDown:
			m.viewport.ScrollDown(3)
			return m, nil
		}
	}
	if m.currentView == viewForm {
		switch msg.Type {
		case tea.MouseWheelUp:
			m.formViewport.ScrollUp(3)
			return m, nil
		case tea.MouseWheelDown:
			m.formViewport.ScrollDown(3)
			return m, nil
		}
	}

	// Only react to an actual left-button click, never to mouse motion/drag.
	if msg.Type != tea.MouseLeft {
		return m, nil
	}

	// Check clickable regions (header tabs, footer hints) first
	if m.clickRegions != nil {
		for _, r := range *m.clickRegions {
			if msg.Y >= r.y1 && msg.Y <= r.y2 && msg.X >= r.x1 && msg.X <= r.x2 {
				return m.handleRegionClick(r.action)
			}
		}
	}

	if m.currentView != viewCommands {
		return m, nil
	}
	contentWidth, contentHeight := m.calculateLayout()
	sidebarWidth := contentWidth / 4
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}
	// Sidebar occupies columns 0..sidebarWidth+1 (incl border)
	if msg.X <= sidebarWidth+1 {
		// Sidebar first category row is at terminal row 6 (0-indexed).
		// Rows: 0-2 header, 3 sep, 4 card border, 5 title, 6 first category.
		rowY := msg.Y - 6
		if rowY >= 0 && rowY < len(m.categories) {
			m.sidebarIndex = rowY
			m.updateFiltered()
		}
		return m, nil
	}
	// Command list area: map Y to command row.
	// First command syntax row is at terminal row 6. Each command occupies
	// 3 rows (syntax, description, blank), so divide by 3.
	bodyLines := contentHeight - 2 - 2
	maxVisible := bodyLines / 3
	if maxVisible < 1 {
		maxVisible = 1
	}
	start := 0
	if m.selectedIndex >= maxVisible {
		start = m.selectedIndex - maxVisible + 1
	}
	rowY := msg.Y - 6
	itemIdx := rowY / 3
	idx := start + itemIdx
	if idx >= 0 && idx < len(m.filteredCmds) {
		m.selectedIndex = idx
		m.textInput.Blur()
		m.inputFocused = false
		m.enterPreview(m.filteredCmds[idx])
	}
	return m, nil
}

func (m model) handleRegionClick(action string) (tea.Model, tea.Cmd) {
	if strings.HasPrefix(action, "form:") {
		if index, err := strconv.Atoi(strings.TrimPrefix(action, "form:")); err == nil {
			m.focusFormField(index)
		}
		return m, nil
	}
	// Placeholder value rows: "ph:<index>"
	if strings.HasPrefix(action, "ph:") {
		if idx, err := strconv.Atoi(strings.TrimPrefix(action, "ph:")); err == nil {
			m.focusPlaceholder(idx)
		}
		return m, nil
	}

	switch action {
	case "tab-commands":
		m.activeTab = 0
		m.currentView = viewCommands
	case "tab-learn":
		m.activeTab = 1
		m.currentView = viewLearn
		m.learnDetailActive = false
	case "help":
		m.showTooltip = !m.showTooltip
		if m.showTooltip {
			m.tooltipContent = m.getHelpTooltip()
		}
	case "esc":
		// mirror the key behavior contextually
		if m.running {
			if m.cancelFn != nil {
				m.cancelFn()
				m.cancelFn = nil
			}
			m.executionID++
			m.running = false
			m.currentView = viewCommands
			m.outputResult = nil
		} else if m.currentView == viewPreview {
			m.currentView = viewCommands
			m.previewCmd = nil
		} else if m.currentView == viewForm {
			m.currentView = viewCommands
			m.form = nil
			m.formCommand = nil
		} else if m.currentView == viewOutput {
			m.currentView = viewCommands
			m.outputResult = nil
		} else if m.currentView == viewLearn {
			if m.learnDetailActive {
				m.learnDetailActive = false
				m.learnContent = ""
			} else {
				m.activeTab = 0
				m.currentView = viewCommands
			}
		}
	case "enter":
		if m.currentView == viewPreview {
			if m.previewCmd != nil {
				if m.firstEmptyPlaceholder() >= 0 {
					m.phWarn = true
					m.focusPlaceholder(m.firstEmptyPlaceholder())
					return m, nil
				}
				return m.executeOrConfirm(m.substitutedCommand(), m.previewCmd.Difficulty, nil)
			}
		} else if m.currentView == viewForm {
			return m.handleFormKey(tea.KeyMsg{Type: tea.KeyEnter})
		} else if m.inputFocused {
			typed := strings.TrimSpace(m.inputValue)
			if typed != "" && !m.isKnownCommand(typed) {
				return m.executeCommand(typed)
			}
			// fall through to selecting list item
			m.textInput.Blur()
			m.inputFocused = false
			if len(m.filteredCmds) > 0 && m.selectedIndex < len(m.filteredCmds) {
				m.enterPreview(m.filteredCmds[m.selectedIndex])
			}
		}
	case "tab":
		if m.inputFocused && len(m.filteredCmds) > 0 && m.inputValue != "" {
			cmd := m.filteredCmds[m.tabIndex%len(m.filteredCmds)]
			m.inputValue = cmd.Full
			m.textInput.SetValue(cmd.Full)
			m.tabIndex++
			m.updateFiltered()
		}
	case "search":
		m.textInput.Focus()
		m.inputFocused = true
	case "learn":
		m.activeTab = 1
		m.currentView = viewLearn
		m.learnDetailActive = false
	case "category":
		// footer "1-6 Category" -> no-op on click (needs a specific number)
	case "rerun":
		if m.currentView == viewOutput {
			return m.executeOrConfirm(m.outputCmd, m.outputDifficulty, m.outputArgs)
		}
	case "copycmd":
		if m.currentView == viewOutput {
			clipboard.WriteAll(m.outputCmd)
			m.lastCopied = "command"
			m.setViewportOutput()
		}
	case "copyout":
		if m.currentView == viewOutput && m.outputResult != nil {
			clipboard.WriteAll(m.outputResult.Output)
			m.lastCopied = "output"
			m.setViewportOutput()
		}
	case "up":
		// navigation - no-op for footer click (would need direction)
	}
	return m, nil
}

func (m model) handleCommandsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputFocused || m.textInput.Focused() {
		switch msg.String() {
		case "esc":
			m.textInput.Blur()
			m.inputFocused = false
			return m, nil
		case "enter":
			typed := strings.TrimSpace(m.inputValue)
			if typed != "" && !m.isKnownCommand(typed) {
				return m.executeCommand(typed)
			}
			m.textInput.Blur()
			m.inputFocused = false
			if len(m.filteredCmds) > 0 && m.selectedIndex < len(m.filteredCmds) {
				m.enterPreview(m.filteredCmds[m.selectedIndex])
			}
			return m, nil
		case "tab":
			if len(m.filteredCmds) > 0 && m.inputValue != "" {
				cmd := m.filteredCmds[m.tabIndex%len(m.filteredCmds)]
				m.inputValue = cmd.Full
				m.textInput.SetValue(cmd.Full)
				m.tabIndex++
				m.updateFiltered()
			}
			return m, nil
		default:
			oldValue := m.textInput.Value()
			m.textInput, _ = m.textInput.Update(msg)
			newValue := m.textInput.Value()
			if oldValue != newValue {
				m.inputValue = newValue
				m.tabIndex = 0
				m.updateFiltered()
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.filteredCmds)-1 {
			m.selectedIndex++
		}
	case "tab":
		if len(m.filteredCmds) > 0 && m.inputValue != "" {
			cmd := m.filteredCmds[m.tabIndex%len(m.filteredCmds)]
			m.inputValue = cmd.Full
			m.textInput.SetValue(cmd.Full)
			m.tabIndex++
			m.updateFiltered()
		}
	case "enter":
		if len(m.filteredCmds) > 0 && m.selectedIndex < len(m.filteredCmds) {
			m.enterPreview(m.filteredCmds[m.selectedIndex])
		}
	case "/":
		m.textInput.Focus()
		m.inputFocused = true
	case "l":
		m.activeTab = 1
		m.currentView = viewLearn
	case "1", "2", "3", "4", "5", "6", "7":
		idx := int(msg.String()[0] - '1')
		if idx < len(m.categories) {
			m.sidebarIndex = idx
			m.updateFiltered()
		}
	case "?":
		m.showTooltip = !m.showTooltip
		if m.showTooltip {
			m.tooltipContent = m.getHelpTooltip()
		}
	}

	return m, nil
}

// enterPreview switches to the preview view for the given command and
// initializes placeholder editing state.
func (m *model) enterPreview(cmd commands.Command) {
	if cmd.Schema != nil {
		m.openForm(cmd)
		return
	}
	m.previewCmd = &cmd
	m.currentView = viewPreview
	m.placeholders = commands.ExtractPlaceholders(cmd.Full)
	m.phValues = make(map[string]string, len(m.placeholders))
	m.phActiveIndex = -1
	m.phWarn = false
	m.phInput.Blur()
	m.phInput.SetValue("")
}

func (m *model) openForm(command commands.Command) {
	m.openCommandForm(command)
	m.formCommand = &command
	m.currentView = viewForm
	m.formInput.Blur()
	if m.form != nil && m.form.fieldCount() > 0 {
		m.form.focusedField = 0
		m.syncFormInput()
	}
	m.formViewport.GotoTop()
}

func formCommandBase(command string) []string {
	fields := strings.Fields(command)
	for i, field := range fields {
		if strings.HasPrefix(field, "-") || strings.Contains(field, "{") {
			return fields[:i]
		}
	}
	return fields
}

func (m *model) syncFormInput() {
	if m.form == nil {
		return
	}
	index := m.form.focusedField
	if index < len(m.form.argumentRows) {
		value := ""
		if len(m.form.argumentRows[index]) == 1 {
			value = m.form.argumentRows[index][0]
		}
		m.formInput.SetValue(value)
		m.formInput.Focus()
		m.formInput.CursorEnd()
		return
	}
	optionIndex := index - len(m.form.argumentRows)
	if optionIndex >= 0 && optionIndex < len(m.form.options) {
		option := m.form.options[optionIndex]
		if option.kind == commands.OptionKindText || option.kind == commands.OptionKindNumeric {
			m.formInput.SetValue(option.value)
			m.formInput.Focus()
			m.formInput.CursorEnd()
		} else {
			m.formInput.Blur()
		}
	}
}

func (m *model) ensureFormFieldVisible() {
	if m.form == nil || m.formViewport == nil {
		return
	}
	line, ok := m.formFieldLines[m.form.focusedField]
	if !ok || m.formViewport.Height < 1 {
		return
	}
	if line < m.formViewport.YOffset {
		m.formViewport.SetYOffset(line)
	} else if line >= m.formViewport.YOffset+m.formViewport.Height {
		m.formViewport.SetYOffset(line - m.formViewport.Height + 1)
	}
}

func (m *model) focusFormField(index int) {
	if m.form == nil || index < 0 || index >= m.form.fieldCount() {
		return
	}
	m.form.focusedField = index
	m.syncFormInput()
	m.ensureFormFieldVisible()
}

func (m *model) formBuild() commands.BuildResult {
	if m.form == nil || m.formCommand == nil {
		return commands.BuildResult{}
	}
	result := commands.Build(formCommandBase(m.formCommand.Full), m.form.commandSchema, m.form.argumentRows, m.form.optionValues())
	m.form.buildResult = result
	return result
}

func (m model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.form == nil {
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.currentView = viewCommands
		m.form = nil
		return m, nil
	case "tab", "down":
		m.form.moveFocus(1)
		m.syncFormInput()
		m.ensureFormFieldVisible()
	case "shift+tab", "up":
		m.form.moveFocus(-1)
		m.syncFormInput()
		m.ensureFormFieldVisible()
	case "left", "right", " ":
		index := m.form.focusedField - len(m.form.argumentRows)
		if index >= 0 && index < len(m.form.options) {
			option := &m.form.options[index]
			if option.kind == commands.OptionKindBoolean {
				if option.value == "true" {
					option.value = "false"
				} else {
					option.value = "true"
				}
			} else if option.kind == commands.OptionKindSelect && len(option.choices) > 0 {
				choice := 0
				for i, candidate := range option.choices {
					if candidate == option.value {
						choice = i
					}
				}
				if msg.String() == "left" {
					choice--
				} else {
					choice++
				}
				if choice < 0 {
					choice = len(option.choices) - 1
				}
				if choice >= len(option.choices) {
					choice = 0
				}
				option.value = option.choices[choice]
			}
		}
	case "enter":
		result := m.formBuild()
		if len(result.Errors) > 0 {
			m.form.validationError = result.Errors[0]
			return m, nil
		}
		m.rememberFormOptions()
		pending := result
		m.pendingForm = &pending
		return m.executeOrConfirm(result.Display, m.formCommand.Difficulty, result.Args)
	case "+", "=":
		if m.form.addRepeatableRow() {
			m.form.focusedField = m.form.fieldCount() - 1
			m.syncFormInput()
		}
		return m, nil
	case "-":
		argumentIndex := m.form.focusedField
		if argumentIndex >= 0 && argumentIndex < len(m.form.argumentRows) && m.form.removeRepeatableRow(argumentIndex) {
			m.syncFormInput()
		}
		return m, nil
	default:
		if m.formInput.Focused() {
			var cmd tea.Cmd
			m.formInput, cmd = m.formInput.Update(msg)
			value := m.formInput.Value()
			if m.form.focusedField < len(m.form.argumentRows) {
				m.form.argumentRows[m.form.focusedField] = []string{value}
			} else {
				option := m.form.focusedField - len(m.form.argumentRows)
				if option >= 0 && option < len(m.form.options) {
					m.form.options[option].value = value
				}
			}
			m.form.validationError = nil
			return m, cmd
		}
	}
	return m, nil
}

// focusPlaceholder activates the placeholder field at index i (if valid),
// loading its current value into the input.
func (m *model) focusPlaceholder(i int) {
	if i < 0 || i >= len(m.placeholders) {
		return
	}
	m.phActiveIndex = i
	m.phInput.SetValue(m.phValues[m.placeholders[i]])
	m.phInput.CursorEnd()
	m.phInput.Focus()
}

// substitutedCommand returns previewCmd.Full with placeholder values applied.
func (m model) substitutedCommand() string {
	if m.previewCmd == nil {
		return ""
	}
	return commands.SubstitutePlaceholders(m.previewCmd.Full, m.phValues)
}

// firstEmptyPlaceholder returns the index of the first placeholder with no
// value, or -1 if all are filled.
func (m model) firstEmptyPlaceholder() int {
	for i, name := range m.placeholders {
		if strings.TrimSpace(m.phValues[name]) == "" {
			return i
		}
	}
	return -1
}

func (m model) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Commands without placeholders keep the simple behavior.
	if len(m.placeholders) == 0 {
		switch msg.String() {
		case "esc":
			m.currentView = viewCommands
			m.previewCmd = nil
			return m, nil
		case "enter":
			if m.previewCmd != nil {
				return m.executeOrConfirm(m.previewCmd.Full, m.previewCmd.Difficulty, nil)
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		// First esc unfocuses the active field; second goes back.
		if m.phActiveIndex >= 0 {
			m.phActiveIndex = -1
			m.phInput.Blur()
			m.phWarn = false
			return m, nil
		}
		m.currentView = viewCommands
		m.previewCmd = nil
		return m, nil

	case "enter":
		if empty := m.firstEmptyPlaceholder(); empty >= 0 {
			m.phWarn = true
			m.focusPlaceholder(empty)
			return m, nil
		}
		return m.executeOrConfirm(m.substitutedCommand(), m.previewCmd.Difficulty, nil)

	case "tab", "down":
		next := m.phActiveIndex + 1
		if next >= len(m.placeholders) {
			next = 0
		}
		m.focusPlaceholder(next)
		return m, nil

	case "shift+tab", "up":
		prev := m.phActiveIndex - 1
		if prev < 0 {
			prev = len(m.placeholders) - 1
		}
		m.focusPlaceholder(prev)
		return m, nil

	default:
		if m.phActiveIndex >= 0 {
			var cmd tea.Cmd
			m.phInput, cmd = m.phInput.Update(msg)
			m.phValues[m.placeholders[m.phActiveIndex]] = m.phInput.Value()
			m.phWarn = false
			return m, cmd
		}
	}
	return m, nil
}

func (m model) handleConfirmationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y", "Y":
		command := m.pendingCommand
		args := m.pendingArgs
		difficulty := m.pendingDifficulty
		m.pendingCommand = ""
		m.pendingArgs = nil
		m.pendingDifficulty = ""
		return m.executeCommandWithDifficulty(command, difficulty, args)
	case "esc", "n", "N", "q":
		m.pendingCommand = ""
		m.pendingArgs = nil
		m.pendingDifficulty = ""
		m.currentView = viewPreview
	}
	return m, nil
}

func (m model) executeCommandWithDifficulty(command, difficulty string, structuredArgs []string) (tea.Model, tea.Cmd) {
	result, cmd := m.executeCommandWithArgs(command, structuredArgs)
	updated := result.(model)
	updated.outputDifficulty = difficulty
	return updated, cmd
}

func (m model) handleOutputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.currentView = viewCommands
		m.outputResult = nil
		m.lastCopied = ""
		m.viewport.SetContent("")
		m.viewportContent = ""
		return m, nil
	case "r":
		return m.executeCommandWithArgs(m.outputCmd, m.outputArgs)
	case "c":
		clipboard.WriteAll(m.outputCmd)
		m.lastCopied = "command"
		m.setViewportOutput()
	case "y":
		if m.outputResult != nil {
			clipboard.WriteAll(m.outputResult.Output)
			m.lastCopied = "output"
			m.setViewportOutput()
		}
	case "g", "home":
		m.viewport.GotoTop()
	case "G", "end":
		m.viewport.GotoBottom()
	default:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m model) handleLearnKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.learnDetailActive {
			m.learnDetailActive = false
			m.learnContent = ""
		} else {
			m.activeTab = 0
			m.currentView = viewCommands
		}
		return m, nil
	case "up", "k":
		if !m.learnDetailActive && m.learnIndex > 0 {
			m.learnIndex--
		}
	case "down", "j":
		if !m.learnDetailActive && m.learnIndex < len(m.learnTopics)-1 {
			m.learnIndex++
		}
	case "enter":
		if !m.learnDetailActive {
			m.learnDetailActive = true
			m.learnContent = m.getLearnContent(m.learnTopics[m.learnIndex])
		}
	}
	return m, nil
}

func (m *model) updateFiltered() {
	category := m.categories[m.sidebarIndex]

	if m.inputValue == "" {
		m.filteredCmds = m.allCommands[category]
	} else {
		var all []commands.Command
		for _, cmds := range m.allCommands {
			all = append(all, cmds...)
		}
		m.filteredCmds = commands.FilterByPrefix(m.inputValue, all)
		if len(m.filteredCmds) == 0 {
			results := commands.Autocomplete(m.inputValue, all)
			m.filteredCmds = make([]commands.Command, len(results))
			for i, r := range results {
				m.filteredCmds[i] = r.Command
			}
		}
	}

	m.selectedIndex = 0
}

// ─── Layout helpers ─────────────────────────────────────────────────

func (m model) calculateLayout() (int, int) {
	contentWidth := m.width
	// Header card(3) + separator(1) + empty line in commands(1) + input box(3) + status bar(3) = 11
	contentHeight := m.height - 11
	if contentHeight < 5 {
		contentHeight = 5
	}
	return contentWidth, contentHeight
}

func truncateString(s string, maxLen int) string {
	if maxLen < 1 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// ─── Rendering ──────────────────────────────────────────────────────

func (m model) View() string {
	if m.clickRegions == nil {
		m.clickRegions = &[]clickRegion{}
	}
	*m.clickRegions = (*m.clickRegions)[:0]

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	switch m.currentView {
	case viewCommands:
		b.WriteString(m.renderCommandsView())
	case viewPreview:
		b.WriteString(m.renderPreviewView())
	case viewForm:
		b.WriteString(m.renderFormView())
	case viewConfirm:
		b.WriteString(m.renderConfirmationView())
	case viewOutput:
		b.WriteString(m.renderOutputView())
	case viewLearn:
		b.WriteString(m.renderLearnView())
	}
	b.WriteString("\n")

	// Count lines rendered so far so the status bar (a 3-line bordered card)
	// can register footer click regions at their true terminal row. The
	// footer content sits on the middle line of that card.
	linesSoFar := strings.Count(b.String(), "\n")
	footerContentRow := linesSoFar + 1

	b.WriteString(m.renderStatusBar(footerContentRow))

	return b.String()
}

func (m model) renderHeader() string {
	// Left side: title + tagline
	title := ui.HeaderTitleStyle.Render("  ▓  WSLC TUI  ")
	tagline := ui.HeaderTaglineStyle.Render("Native containers for WSL")
	leftSide := lipgloss.JoinHorizontal(lipgloss.Center, title, tagline)

	// Right side: tabs + help
	commandsTab := ui.HeaderTabStyle.Render("Commands")
	learnTab := ui.HeaderTabStyle.Render("Learn")
	if m.activeTab == 0 {
		commandsTab = ui.HeaderTabActiveStyle.Render("Commands")
	} else {
		learnTab = ui.HeaderTabActiveStyle.Render("Learn")
	}
	helpHint := ui.HeaderHelpStyle.Render("[?] Help")
	rightSide := lipgloss.JoinHorizontal(lipgloss.Center, commandsTab, learnTab, helpHint)

	// Register clickable regions for the header tabs / help.
	// Content sits inside a bordered card (1-col border, no padding) and the
	// left/right blocks are joined left-anchored with a 2-space separator:
	//   col 0 = border, col 1 = leftSide start, then 2 spaces, then rightSide.
	rightStart := 1 + lipgloss.Width(leftSide) + 2
	cx := rightStart
	addRegion := func(s string, action string) {
		if m.clickRegions == nil {
			return
		}
		w := lipgloss.Width(s)
		*m.clickRegions = append(*m.clickRegions, clickRegion{
			x1: cx, x2: cx + w - 1,
			y1: 1, y2: 1, // header content is on row 1 (0-indexed)
			action: action,
		})
		cx += w
	}
	addRegion(commandsTab, "tab-commands")
	addRegion(learnTab, "tab-learn")
	addRegion(helpHint, "help")

	// Join with flexible spacing
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, "  ", rightSide)

	// Wrap in card border
	return ui.CardActiveStyle.
		Width(m.width - 2).
		Render(content)
}

func (m model) renderCommandsView() string {
	sidebarWeight := 1
	listWeight := 3
	contentWidth, contentHeight := m.calculateLayout()
	totalWeight := sidebarWeight + listWeight
	sidebarWidth := (contentWidth * sidebarWeight) / totalWeight
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}
	listWidth := contentWidth - sidebarWidth - 2 // -2 for gap

	sidebar := m.renderSidebar(sidebarWidth)
	content := m.renderCommandList(listWidth, contentHeight)
	input := m.renderInput()

	// Sidebar + command list side by side
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", content)

	var parts []string
	parts = append(parts, mainContent)
	parts = append(parts, "")
	parts = append(parts, input)

	if m.showTooltip {
		parts = append(parts, m.renderTooltip())
	}

	return strings.Join(parts, "\n")
}

func (m model) renderSidebar(width int) string {
	var lines []string
	lines = append(lines, ui.SidebarTitleStyle.Render(" CATEGORIES"))
	lines = append(lines, "")

	maxCatWidth := width - 6 // -2 borders, -2 padding, -2 for icon

	for i, cat := range m.categories {
		icon := ui.GetCategoryIcon(cat)
		count := len(m.allCommands[cat])
		countStr := fmt.Sprintf("%d", count)

		if i == m.sidebarIndex {
			item := fmt.Sprintf("%s  %s", icon, cat)
			item = truncateString(item, maxCatWidth)
			item = padRight(item, maxCatWidth) // ensure full-width bg
			countPart := fmt.Sprintf(" %s", ui.SidebarItemCountStyle.Render(countStr))
			lines = append(lines, ui.SidebarItemActiveBgStyle.Render(item)+countPart)
		} else {
			item := fmt.Sprintf(" %s  %s", icon, cat)
			item = truncateString(item, maxCatWidth)
			countPart := fmt.Sprintf(" %s", ui.SidebarItemCountStyle.Render(countStr))
			lines = append(lines, ui.SidebarItemStyle.Render(item)+countPart)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return ui.CardStyle.
		Width(width).
		Render(content)
}

func (m model) renderCommandList(listWidth, contentHeight int) string {
	if listWidth < 30 {
		listWidth = 30
	}

	if len(m.filteredCmds) == 0 {
		empty := lipgloss.JoinVertical(lipgloss.Center, "",
			ui.CmdDescStyle.Render("No commands found"),
			"",
		)
		return ui.CardStyle.Width(listWidth).Render(empty)
	}

	// Header
	cat := m.categories[m.sidebarIndex]
	cmdCount := fmt.Sprintf("%d commands", len(m.filteredCmds))
	header := fmt.Sprintf(" %s  %s",
		ui.CmdHeaderStyle.Render(cat),
		ui.SidebarItemCountStyle.Render(cmdCount),
	)

	// Command list. Each command occupies 3 rows (syntax, description, blank),
	// and the card body reserves the header + a blank line (2 rows) plus the
	// 2 border rows. Compute how many commands fit in the available height.
	bodyLines := contentHeight - 2 - 2 // -2 borders, -2 header+blank
	maxVisible := bodyLines / 3
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := 0
	if m.selectedIndex >= maxVisible {
		start = m.selectedIndex - maxVisible + 1
	}

	end := start + maxVisible
	if end > len(m.filteredCmds) {
		end = len(m.filteredCmds)
	}

	maxTextWidth := listWidth - 4 // -2 borders, -2 padding

	var cmdLines []string
	cmdLines = append(cmdLines, header)
	cmdLines = append(cmdLines, "")

	for i := start; i < end; i++ {
		cmd := m.filteredCmds[i]
		badge := ui.GetDifficultyBadge(cmd.Difficulty)
		isSelected := i == m.selectedIndex

		syntax := cmd.Full
		syntax = truncateString(syntax, maxTextWidth)

		// Command syntax (first line)
		var syntaxLine string
		if isSelected {
			syntaxLine = ui.CmdCursorStyle.Render("▸ ") + badge + " " + ui.CmdSyntaxStyle.Render(syntax)
		} else {
			syntaxLine = "  " + badge + " " + ui.CmdSyntaxStyle.Render(syntax)
		}

		desc := truncateString(cmd.Description, maxTextWidth)
		descLine := "    " + ui.CmdDescStyle.Render(desc)

		cmdLines = append(cmdLines, syntaxLine)
		cmdLines = append(cmdLines, descLine)

		if i < end-1 {
			cmdLines = append(cmdLines, "")
		}
	}

	// Scroll indicator
	if len(m.filteredCmds) > maxVisible {
		cmdLines = append(cmdLines, "")
		scrollInfo := fmt.Sprintf("↑ %d/%d ↓", m.selectedIndex+1, len(m.filteredCmds))
		cmdLines = append(cmdLines, ui.ScrollStyle.Render("  "+scrollInfo))
	}

	innerHeight := contentHeight - 2
	targetLines := innerHeight
	for len(cmdLines) < targetLines {
		cmdLines = append(cmdLines, "")
	}

	return ui.CardStyle.
		Width(listWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, cmdLines...))
}

func (m model) renderInput() string {
	prompt := ui.InputPromptStyle.Render("❯ ")
	input := m.textInput.View()

	inputContent := lipgloss.JoinHorizontal(lipgloss.Center, prompt, input)

	style := ui.InputBoxFocusedStyle
	if !m.inputFocused {
		style = ui.InputBoxStyle
	}

	return style.
		Width(m.width - 2).
		Render(inputContent)
}

func (m model) renderTooltip() string {
	return "\n" + ui.TooltipStyle.Render(m.tooltipContent)
}

// renderStyledCommand renders previewCmd.Full with each {placeholder} token
// styled according to its state (empty/filled/active).
func (m model) renderStyledCommand() string {
	if m.previewCmd == nil {
		return ""
	}
	full := m.previewCmd.Full
	if len(m.placeholders) == 0 {
		return ui.PreviewCmdStyle.Render(full)
	}

	// Map placeholder name -> index for active/value lookup.
	idxOf := make(map[string]int, len(m.placeholders))
	for i, n := range m.placeholders {
		idxOf[n] = i
	}

	var b strings.Builder
	rest := full
	for {
		loc := commands.FindPlaceholderIndex(rest)
		if loc == nil {
			b.WriteString(ui.PreviewCmdStyle.Render(rest))
			break
		}
		b.WriteString(ui.PreviewCmdStyle.Render(rest[:loc[0]]))
		token := rest[loc[0]:loc[1]]
		name := token[1 : len(token)-1]
		i := idxOf[name]
		v := m.phValues[name]
		switch {
		case i == m.phActiveIndex:
			shown := v
			if strings.TrimSpace(shown) == "" {
				shown = token
			}
			b.WriteString(ui.PlaceholderActiveStyle.Render(shown))
		case strings.TrimSpace(v) != "":
			b.WriteString(ui.PlaceholderFilledStyle.Render(v))
		default:
			b.WriteString(ui.PlaceholderEmptyStyle.Render(token))
		}
		rest = rest[loc[1]:]
	}
	return b.String()
}

func (m model) renderPreviewView() string {
	if m.previewCmd == nil {
		return "No command selected"
	}

	cmd := m.previewCmd
	_, contentHeight := m.calculateLayout()
	innerWidth := m.width - 6

	var lines []string

	// Difficulty badge
	badge := ui.GetDifficultyLabel(cmd.Difficulty)

	// Command section (with styled, editable placeholders)
	lines = append(lines, ui.SectionLabelStyle.Render("  COMMAND"))
	lines = append(lines, "  "+ui.PreviewCmdStyle.Render("$ ")+m.renderStyledCommand())
	lines = append(lines, "")

	// Description
	lines = append(lines, ui.SectionLabelStyle.Render("  DESCRIPTION"))
	lines = append(lines, "  "+ui.PreviewDescStyle.Render(cmd.Description))
	lines = append(lines, "")

	// Values section (only when the command has placeholders).
	// Rows here are clickable to focus a field. The preview card starts at
	// terminal row 4 (header=3 lines + 1 blank), its own top border is row 4,
	// the card header is row 5, so content line index i renders at row 6+i.
	if len(m.placeholders) > 0 {
		lines = append(lines, ui.SectionLabelStyle.Render("  VALUES"))
		for i, name := range m.placeholders {
			rowLineIndex := len(lines) // index of the row we're about to add
			terminalRow := 6 + rowLineIndex

			label := ui.PlaceholderLabelStyle.Render(padRight(name, 16))
			var valueField string
			if i == m.phActiveIndex {
				valueField = m.phInput.View()
			} else {
				v := m.phValues[name]
				if strings.TrimSpace(v) == "" {
					valueField = ui.PlaceholderEmptyStyle.Render("‹empty›")
				} else {
					valueField = ui.PlaceholderFilledStyle.Render(v)
				}
			}
			cursor := "  "
			if i == m.phActiveIndex {
				cursor = ui.CmdCursorStyle.Render("▸ ")
			}
			lines = append(lines, cursor+label+"  "+valueField)

			// Register a clickable region spanning this row's inner width.
			if m.clickRegions != nil {
				*m.clickRegions = append(*m.clickRegions, clickRegion{
					x1: 1, x2: m.width - 2,
					y1: terminalRow, y2: terminalRow,
					action: fmt.Sprintf("ph:%d", i),
				})
			}
		}
		lines = append(lines, "")
	}

	// Flags
	if len(cmd.Flags) > 0 {
		lines = append(lines, ui.SectionLabelStyle.Render("  FLAGS"))

		// Table header
		flagHeader := fmt.Sprintf("  %s  %s",
			ui.CmdHeaderStyle.Render(padRight("Flag", 22)),
			ui.CmdHeaderStyle.Render("Description"),
		)
		lines = append(lines, flagHeader)

		// Table separator
		lines = append(lines, ui.SidebarItemCountStyle.Render("  "+strings.Repeat("─", innerWidth-2)))

		// Flag rows
		for _, flag := range cmd.Flags {
			flagName := flag.Long
			if flag.Short != "" {
				flagName = flag.Short + ", " + flag.Long
			}
			flagDefault := ""
			if flag.Default != "" {
				flagDefault = ui.PreviewDefaultStyle.Render(fmt.Sprintf(" (default: %s)", flag.Default))
			}

			maxFlagWidth := innerWidth - 26 // account for flag name column
			flagDesc := truncateString(flag.Description, maxFlagWidth)

			row := fmt.Sprintf("  %s  %s%s",
				ui.PreviewFlagStyle.Render(padRight(flagName, 22)),
				ui.PreviewDescStyle.Render(flagDesc),
				flagDefault,
			)
			lines = append(lines, row)
		}
		lines = append(lines, "")
	}

	// Examples
	if len(cmd.Examples) > 0 {
		lines = append(lines, ui.SectionLabelStyle.Render("  EXAMPLES"))
		for _, ex := range cmd.Examples {
			lines = append(lines, "  "+ui.PreviewExampleStyle.Render("$ "+ex))
		}
		lines = append(lines, "")
	}

	// Placeholder hint / warning line.
	if len(m.placeholders) > 0 {
		if m.phWarn && m.firstEmptyPlaceholder() >= 0 {
			lines = append(lines, "  "+ui.PlaceholderWarnStyle.Render("⚠ Fill all fields before running"))
		} else {
			lines = append(lines, "  "+ui.PreviewDefaultStyle.Render("Tab/↑↓ next field • click a value to edit • Enter run • Esc back"))
		}
	}

	// Pad to fill available space
	innerHeight := contentHeight - 2
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	// Wrap in card with header
	headerContent := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.PreviewHeaderStyle.Render("  ▓  PREVIEW"),
		"  ",
		badge,
	)

	return ui.CardStyle.
		Width(m.width - 2).
		Render(headerContent + "\n" + content)
}

func (m model) renderFormView() string {
	if m.form == nil || m.formCommand == nil {
		return "No command selected"
	}
	result := m.formBuild()
	if m.formViewport.Width < 1 {
		m.formViewport.Width = m.width - 4
	}
	if m.formViewport.Height < 1 {
		_, contentHeight := m.calculateLayout()
		m.formViewport.Height = contentHeight - 4
	}
	if m.formViewport.Width < 1 {
		m.formViewport.Width = 1
	}
	if m.formViewport.Height < 1 {
		m.formViewport.Height = 1
	}
	if m.formFieldLines == nil {
		m.formFieldLines = make(map[int]int)
	}
	for field := range m.formFieldLines {
		delete(m.formFieldLines, field)
	}
	registerField := func(field, line int) {
		m.formFieldLines[field] = line
		terminalRow := 6 + line - m.formViewport.YOffset
		if terminalRow >= 6 && terminalRow < 6+m.formViewport.Height && m.clickRegions != nil {
			*m.clickRegions = append(*m.clickRegions, clickRegion{
				x1: 1, x2: m.width - 2,
				y1: terminalRow, y2: terminalRow,
				action: "form:" + strconv.Itoa(field),
			})
		}
	}
	lines := []string{
		ui.SectionLabelStyle.Render("  ARGUMENTS"),
	}
	for i, argument := range m.form.commandSchema.Arguments {
		for row := i; row < len(m.form.argumentRows) && (argument.Repeatable || row == i); row++ {
			line := len(lines)
			value := ""
			if len(m.form.argumentRows[row]) == 1 {
				value = m.form.argumentRows[row][0]
			}
			field := value
			if m.form.focusedField == row && m.formInput.Focused() {
				field = m.formInput.View()
			}
			if field == "" {
				field = ui.PlaceholderEmptyStyle.Render("<empty>")
			}
			marker := "  "
			if m.form.focusedField == row {
				marker = ui.CmdCursorStyle.Render("▸ ")
			}
			lines = append(lines, marker+ui.PlaceholderLabelStyle.Render(argument.Name)+"  "+field)
			registerField(row, line)
			if !argument.Repeatable {
				break
			}
		}
	}
	if len(m.form.commandSchema.Arguments) == 0 {
		lines = append(lines, "  "+ui.PreviewDefaultStyle.Render("No positional arguments"))
	}
	if len(m.form.commandSchema.Arguments) > 0 && m.form.commandSchema.Arguments[len(m.form.commandSchema.Arguments)-1].Repeatable {
		lines = append(lines, "  "+ui.PreviewDefaultStyle.Render("+ add argument row   - remove focused row"))
	}
	lines = append(lines, "", ui.SectionLabelStyle.Render("  OPTIONS"))
	for i, option := range m.form.options {
		fieldIndex := len(m.form.argumentRows) + i
		line := len(lines)
		value := option.value
		switch option.kind {
		case commands.OptionKindBoolean:
			if value == "true" {
				value = "[x]"
			} else {
				value = "[ ]"
			}
		case commands.OptionKindSelect:
			value = "<" + value + ">"
		}
		if fieldIndex == m.form.focusedField && m.formInput.Focused() && (option.kind == commands.OptionKindText || option.kind == commands.OptionKindNumeric) {
			value = m.formInput.View()
		}
		marker := "  "
		if fieldIndex == m.form.focusedField {
			marker = ui.CmdCursorStyle.Render("▸ ")
		}
		lines = append(lines, marker+ui.PreviewFlagStyle.Render(option.flag)+"  "+value)
		registerField(fieldIndex, line)
	}
	lines = append(lines, "", ui.SectionLabelStyle.Render("  GENERATED COMMAND"), "  "+ui.PreviewCmdStyle.Render("$ "+result.Display))
	if m.form.validationError != nil {
		lines = append(lines, "  "+ui.PlaceholderWarnStyle.Render("⚠ "+m.form.validationError.Error()))
	}
	lines = append(lines, "", ui.SectionLabelStyle.Render("  EXAMPLES / HELP"))
	for _, example := range m.formCommand.Examples {
		lines = append(lines, "  "+ui.PreviewExampleStyle.Render("$ "+example))
	}
	lines = append(lines, "", ui.PreviewDefaultStyle.Render("↑↓/Tab navigate • Space toggle • ←→ select • Enter submit • Esc back"))

	m.formViewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, lines...))
	content := m.formViewport.View()
	header := ui.PreviewHeaderStyle.Render("  ▓  GUIDED FORM")
	return ui.CardStyle.Width(m.width - 2).Render(header + "\n" + content)
}

func (m model) renderConfirmationView() string {
	_, contentHeight := m.calculateLayout()
	innerHeight := contentHeight - 2
	if innerHeight < 6 {
		innerHeight = 6
	}

	var lines []string
	if m.pendingDifficulty == "advanced" {
		lines = append(lines, ui.PlaceholderWarnStyle.Render("  ⚠ HIGH-RISK COMMAND"))
		lines = append(lines, "  This command can delete data, stop workloads, or change system state.")
	} else {
		lines = append(lines, ui.PreviewExampleStyle.Render("  ⚠ REVIEW BEFORE RUNNING"))
		lines = append(lines, "  This command can change containers, networks, images, or files.")
	}
	lines = append(lines, "")
	lines = append(lines, ui.SectionLabelStyle.Render("  COMMAND"))
	lines = append(lines, "  "+ui.PreviewCmdStyle.Render("$ "+m.pendingCommand))
	lines = append(lines, "")
	lines = append(lines, ui.PreviewDefaultStyle.Render("  Enter/Y: run    Esc/N: cancel"))
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return ui.CardStyle.
		Width(m.width - 2).
		Render(ui.PreviewHeaderStyle.Render("  ▓  CONFIRM COMMAND") + "\n" + content)
}

func (m model) renderRunningView() string {
	_, contentHeight := m.calculateLayout()
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "  "+ui.OutputJsonStringStyle.Render("⟳  Executing command..."))
	lines = append(lines, "  "+ui.CmdDescStyle.Render(m.outputCmd))
	lines = append(lines, "")
	lines = append(lines, "  "+ui.ActionHintKeyStyle.Render("Esc")+ui.ActionHintStyle.Render(" Cancel"))

	innerHeight := contentHeight - 2
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	headerContent := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.OutputHeaderStyle.Render("  ▓  RUNNING"),
	)

	return ui.CardActiveStyle.
		Width(m.width - 2).
		Render(headerContent + "\n" + content)
}

func (m model) renderOutputView() string {
	if m.running {
		return m.renderRunningView()
	}

	if m.outputResult == nil {
		return "No output"
	}

	viewportContent := m.viewport.View()

	// Status indicator
	statusIcon := "  ✓ "
	statusColor := ui.OutputSuccessStyle
	if m.outputResult.ExitCode != 0 {
		statusIcon = "  ✗ "
		statusColor = ui.OutputErrorStyle
	}

	// Scroll indicator: shown only when content overflows the viewport.
	scrollInfo := ""
	if m.viewport.TotalLineCount() > m.viewport.Height {
		pct := int(m.viewport.ScrollPercent() * 100)
		if pct < 0 {
			pct = 0
		} else if pct > 100 {
			pct = 100
		}
		label := fmt.Sprintf(" %d%% ", pct)
		if m.viewport.AtTop() {
			label = " TOP "
		} else if m.viewport.AtBottom() {
			label = " END "
		}
		scrollInfo = ui.ScrollStyle.Render(label)
	}

	left := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.OutputHeaderStyle.Render("  ▓  OUTPUT"),
		"  ",
		statusColor.Render(statusIcon+m.outputCmd),
	)

	headerContent := left
	if scrollInfo != "" {
		gap := m.width - 4 - lipgloss.Width(left) - lipgloss.Width(scrollInfo)
		if gap < 1 {
			gap = 1
		}
		headerContent = left + strings.Repeat(" ", gap) + scrollInfo
	}

	return ui.CardStyle.
		Width(m.width - 2).
		Render(headerContent + "\n" + viewportContent)
}

func (m model) renderLearnView() string {
	if m.learnDetailActive {
		return m.renderLearnDetail()
	}

	sidebarWeight := 1
	contentWeight := 3
	contentWidth, contentHeight := m.calculateLayout()
	totalWeight := sidebarWeight + contentWeight
	sidebarWidth := (contentWidth * sidebarWeight) / totalWeight
	contentAreaWidth := contentWidth - sidebarWidth - 2

	// Sidebar
	var sidebarLines []string
	sidebarLines = append(sidebarLines, ui.SidebarTitleStyle.Render(" TOPICS"))
	sidebarLines = append(sidebarLines, "")

	for i, topic := range m.learnTopics {
		if i == m.learnIndex {
			sidebarLines = append(sidebarLines, ui.LearnTopicActiveStyle.Render(" ▸ "+topic))
		} else {
			sidebarLines = append(sidebarLines, ui.LearnTopicStyle.Render("   "+topic))
		}
	}

	// Pad sidebar to fill available space
	innerHeight := contentHeight - 2
	for len(sidebarLines) < innerHeight {
		sidebarLines = append(sidebarLines, "")
	}

	sidebar := ui.CardStyle.
		Width(sidebarWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, sidebarLines...))

	// Content area (topic list with descriptions)
	var contentLines []string
	contentLines = append(contentLines, ui.CmdHeaderStyle.Render(" LEARN WSLC"))
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, ui.CmdDescStyle.Render("  Select a topic to learn about WSLC"))
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, "")

	for i, topic := range m.learnTopics {
		icon := "▸"
		if i != m.learnIndex {
			icon = "·"
		}

		style := ui.LearnTopicStyle
		if i == m.learnIndex {
			style = ui.LearnTopicActiveStyle
		}

		contentLines = append(contentLines, "  "+style.Render(icon+" "+topic))
		contentLines = append(contentLines, "")
	}

	// Pad content to fill available space
	for len(contentLines) < innerHeight {
		contentLines = append(contentLines, "")
	}

	content := ui.CardStyle.
		Width(contentAreaWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, contentLines...))

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", content)
}

func (m model) renderLearnDetail() string {
	_, contentHeight := m.calculateLayout()
	innerWidth := m.width - 6

	var lines []string
	lines = append(lines, ui.SectionLabelStyle.Render("  TOPIC: "+strings.ToUpper(m.learnTopics[m.learnIndex])))
	lines = append(lines, "")

	// Render the content line by line for proper styling
	maxContentWidth := innerWidth - 4
	for _, line := range strings.Split(m.learnContent, "\n") {
		styledLine := styleLearnLine(line)
		styledLine = truncateString(styledLine, maxContentWidth)
		lines = append(lines, "  "+styledLine)
	}

	// Pad to fill available space
	innerHeight := contentHeight - 2
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	headerContent := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.LearnTopicActiveStyle.Render("  ▓  LEARN"),
	)

	return ui.CardStyle.
		Width(innerWidth).
		Render(headerContent + "\n" + content)
}

func styleLearnLine(line string) string {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return ""
	}

	// Section headers (lines ending with ":" that are all caps or short)
	if strings.HasSuffix(trimmed, ":") && len(trimmed) < 40 && !strings.HasPrefix(trimmed, "$") && !strings.HasPrefix(trimmed, "wslc") {
		return ui.SectionLabelStyle.Render(trimmed)
	}

	// Command examples (start with "wslc" or "$")
	if strings.HasPrefix(trimmed, "wslc") {
		return ui.PreviewCmdStyle.Render("  " + trimmed)
	}
	if strings.HasPrefix(trimmed, "$ ") {
		return ui.PreviewExampleStyle.Render("  " + trimmed)
	}

	// Numbered list items
	if len(trimmed) > 2 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
		parts := strings.SplitN(trimmed, ": ", 2)
		if len(parts) == 2 {
			return ui.PreviewFlagStyle.Render(parts[0]+": ") + ui.PreviewDescStyle.Render(parts[1])
		}
	}

	// Default text
	return ui.PreviewDescStyle.Render(trimmed)
}

func (m model) renderStatusBar(contentRowY int) string {
	var parts []string

	if m.running {
		parts = append(parts, renderKeyHint("Esc", "Cancel"))
	} else if m.currentView == viewCommands {
		if m.inputFocused || m.textInput.Focused() {
			parts = append(parts, renderKeyHint("Esc", "Unfocus"))
			parts = append(parts, renderKeyHint("Tab", "Complete"))
			parts = append(parts, renderKeyHint("Enter", "Select"))
		} else {
			parts = append(parts, renderKeyHint("↑↓", "Navigate"))
			parts = append(parts, renderKeyHint("Enter", "Run/Preview"))
			parts = append(parts, renderKeyHint("Tab", "Complete"))
			parts = append(parts, renderKeyHint("/", "Search"))
			parts = append(parts, renderKeyHint("l", "Learn"))
			parts = append(parts, renderKeyHint("1-7", "Category"))
			parts = append(parts, renderKeyHint("?", "Help"))
		}
	} else if m.currentView == viewPreview {
		if len(m.placeholders) > 0 {
			parts = append(parts, renderKeyHint("Tab", "Next field"))
			parts = append(parts, renderKeyHint("Enter", "Execute"))
			parts = append(parts, renderKeyHint("Esc", "Back"))
		} else {
			parts = append(parts, renderKeyHint("Enter", "Execute"))
			parts = append(parts, renderKeyHint("Esc", "Back"))
		}
	} else if m.currentView == viewForm {
		parts = append(parts, renderKeyHint("↑↓/Tab", "Navigate"))
		parts = append(parts, renderKeyHint("Enter", "Submit"))
		parts = append(parts, renderKeyHint("Esc", "Back"))
	} else if m.currentView == viewConfirm {
		parts = append(parts, renderKeyHint("Enter/Y", "Run"))
		parts = append(parts, renderKeyHint("Esc/N", "Cancel"))
	} else if m.currentView == viewOutput {
		parts = append(parts, renderKeyHint("↑↓/wheel", "Scroll"))
		parts = append(parts, renderKeyHint("PgUp/PgDn", "Page"))
		parts = append(parts, renderKeyHint("g/G", "Top/End"))
		parts = append(parts, renderKeyHint("r", "Re-run"))
		parts = append(parts, renderKeyHint("y", "Copy"))
		parts = append(parts, renderKeyHint("Esc", "Back"))
	} else if m.currentView == viewLearn {
		parts = append(parts, renderKeyHint("↑↓", "Navigate"))
		parts = append(parts, renderKeyHint("Enter", "Open"))
		parts = append(parts, renderKeyHint("Esc", "Back"))
	}

	hints := lipgloss.JoinHorizontal(lipgloss.Center, parts...)

	// Register clickable regions for the footer key hints.
	// contentRowY is the true terminal row of the footer content line,
	// computed by View() from the actual rendered output height.
	y := contentRowY
	cx := 2
	for _, p := range parts {
		w := lipgloss.Width(p)
		if w > 0 && m.clickRegions != nil {
			*m.clickRegions = append(*m.clickRegions, clickRegion{
				x1: cx, x2: cx + w - 1,
				y1: y, y2: y,
				action: hintAction(p),
			})
		}
		cx += w
	}

	return ui.CardStyle.
		Width(m.width - 2).
		Render(" " + hints)
}

func renderKeyHint(key, label string) string {
	return "  " + ui.ActionHintKeyStyle.Render(key) + " " + ui.ActionHintStyle.Render(label) + "  "
}

func hintAction(s string) string {
	clean := stripAnsi(s)
	clean = strings.TrimSpace(clean)
	// clean looks like "Esc Unfocus" or "↑↓ Navigate"
	fields := strings.SplitN(clean, " ", 2)
	if len(fields) == 0 {
		return ""
	}
	key := fields[0]
	switch key {
	case "Esc":
		return "esc"
	case "Tab":
		return "tab"
	case "Enter":
		return "enter"
	case "↑↓":
		return "up" // navigate
	case "/":
		return "search"
	case "l":
		return "learn"
	case "1-7":
		return "category"
	case "?":
		return "help"
	case "r":
		return "rerun"
	case "c":
		return "copycmd"
	case "y":
		return "copyout"
	}
	return ""
}

func stripAnsi(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (m model) getHelpTooltip() string {
	return `WSLC TUI - Microsoft WSL Containers

Navigation:
  ↑/k       Move up
  ↓/j       Move down
  Tab       Accept autocomplete
  Enter     Select command
  Esc       Back / Cancel

Commands:
  1-7       Jump to category
  /         Focus search
  l         Switch to Learn tab
  ?         Toggle this help
  q         Quit

Output View:
  ↑/↓       Scroll
  PgUp/PgDn Scroll page
  r         Re-run command
  c         Copy command
  y         Copy output`
}

func (m model) getLearnContent(topic string) string {
	switch topic {
	case "Getting Started":
		return `GETTING STARTED WITH WSLC

WSLC (Windows Subsystem for Linux Containers) is Microsoft's native
container runtime built into WSL 2.9.3+. It provides a Docker-like CLI
for managing Linux containers without Docker Desktop.

Basic Commands:
  wslc container ls              List running containers
  wslc image ls                  List available images
  wslc info                      Show system information

Quick Start:
  1. Pull an image:     wslc pull ubuntu:latest
  2. Run a container:   wslc run -d --name myapp ubuntu:latest
  3. Execute command:   wslc exec -it myapp bash
  4. View logs:         wslc logs myapp
  5. Stop it:           wslc stop myapp

Tip: Use Tab to autocomplete commands in the command palette!`

	case "Container Operations":
		return `CONTAINER OPERATIONS

List Containers:
  wslc container ls                          List running containers
  wslc container ls --all                    List all containers
  wslc container ls --all --format json      JSON output for scripting

Run Containers:
  wslc run -d --name myapp ubuntu:latest     Detached mode
  wslc run -it --rm ubuntu:latest bash       Interactive shell
  wslc run -d -p 8080:80 --name web nginx    Port mapping
  wslc run --gpus all --name gpu-app ubuntu  GPU-enabled container

Execute Commands:
  wslc exec -it CONTAINER bash       Interactive shell
  wslc exec CONTAINER command        Run single command
  wslc exec -u root CONTAINER cmd    Run as root

View Logs:
  wslc logs CONTAINER                Show logs
  wslc logs -f CONTAINER             Follow logs
  wslc logs --tail 100 CONTAINER     Last 100 lines
  wslc logs --timestamps CONTAINER   Show with timestamps

Lifecycle:
  wslc start CONTAINER     Start a stopped container
  wslc stop CONTAINER      Stop a running container
  wslc kill CONTAINER      Force stop with signal

Cleanup:
  wslc remove CONTAINER        Remove a container
  wslc remove -f CONTAINER     Force remove
  wslc container prune         Remove all stopped containers`

	case "Image Management":
		return `IMAGE MANAGEMENT

Pull Images:
  wslc pull IMAGE              Pull from registry
  wslc pull ubuntu:latest      Pull specific tag

List Images:
  wslc image ls                List local images
  wslc image ls --all          Show all images

Tag Images:
  wslc tag SOURCE TARGET       Create new tag
  wslc tag myapp:latest myapp:v1.0

Remove Images:
  wslc rmi IMAGE               Remove an image
  wslc rmi -f IMAGE            Force remove

Inspect Images:
  wslc inspect IMAGE           Show detailed info

Build Images:
  wslc build -t myapp:latest .         Build from Containerfile
  wslc build -f Dockerfile.prod .      Custom Dockerfile
  wslc build --label env=prod .        With metadata labels

Save/Load:
  wslc save -o backup.tar IMAGE        Save to tar
  wslc load -i backup.tar              Load from tar`

	case "Network & Volume":
		return `NETWORK & VOLUME MANAGEMENT

Network Commands:
  wslc network ls                          List networks
  wslc network create my-network           Create a network
  wslc network rm my-network               Remove a network
  wslc network inspect my-network          Show network details
  wslc network connect my-net container    Attach container to network
  wslc network disconnect my-net container Detach container from network
  wslc network prune                       Remove unused networks

Volume Commands:
  wslc volume ls                           List volumes
  wslc volume create myvolume              Create a volume
  wslc volume rm myvolume                  Remove a volume
  wslc volume inspect myvolume             Show volume details
  wslc volume prune                        Remove unused volumes

Usage with Containers:
  wslc run -v myvolume:/data --name app ubuntu
  wslc run --network my-network --name app ubuntu
  wslc run -v /host/path:/container/path --name app ubuntu`

	case "Sessions":
		return `SESSION MANAGEMENT

WSLC sessions are named containers with configurable storage and
resource limits. A default session is created on demand.

List Sessions:
  wslc session ls              List all sessions

Enter a Session:
  wslc session enter my-session     Connect to session

Run in Session:
  wslc session run my-session echo hello
  wslc session run --cpus 4 --memory 4096 my-session bash

Open Shell:
  wslc session shell my-session     Interactive shell

Terminate:
  wslc session terminate my-session     Release session resources

Session Resources:
  Default storage: 32 GB per session
  Configurable CPU and memory limits
  VHD-backed storage for performance`

	case "System & Maintenance":
		return `SYSTEM & MAINTENANCE

System Info:
  wslc info                  Show system information

Cleanup:
  wslc system prune          Remove all unused data
  wslc system prune -f       Skip confirmation
  wslc system prune --volumes  Also prune volumes

Version:
  wslc version               Show wslc version

Tips:
  - Run cleanup regularly to free disk space
  - Check 'wslc info' for system status
  - Use 'wslc container prune' for just stopped containers
  - WSLC runs natively on WSL 2.9.3+ — no Docker daemon needed`
	}

	return "No content available for this topic."
}

func formatJSON(s string) string {
	// A single JSON value (object or array) indents directly.
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err == nil {
		return buf.String()
	}

	// Otherwise, treat it as JSON Lines (one JSON value per line), which is
	// what tools like `wslc container ls --format json` emit for multiple
	// items. Indent each valid line; leave any non-JSON line untouched.
	lines := strings.Split(s, "\n")
	formatted := make([]string, 0, len(lines))
	any := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		var b bytes.Buffer
		if err := json.Indent(&b, []byte(trimmed), "", "  "); err == nil {
			if any {
				formatted = append(formatted, "")
			}
			formatted = append(formatted, b.String())
			any = true
		} else {
			formatted = append(formatted, line)
		}
	}
	if !any {
		return s
	}
	return strings.Join(formatted, "\n")
}

func highlightJSONLine(line string) string {
	trimmed := strings.TrimSpace(line)

	if idx := strings.Index(trimmed, ":"); idx > 0 && strings.HasPrefix(trimmed, `"`) {
		keyEnd := strings.Index(trimmed[1:], `"`) + 1
		if keyEnd > 0 {
			key := trimmed[:keyEnd+1]
			rest := trimmed[keyEnd+1:]
			return "  " + ui.OutputJsonKeyStyle.Render(key) + rest
		}
	}

	if strings.HasPrefix(trimmed, `"`) {
		return ui.OutputJsonStringStyle.Render(line)
	}

	if trimmed == "true" || trimmed == "false" || trimmed == "null" {
		return ui.OutputJsonNumberStyle.Render(line)
	}

	if len(trimmed) > 0 && ((trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-') {
		return ui.OutputJsonNumberStyle.Render(line)
	}

	return line
}

// padRight pads a string with spaces to reach the target visual width.
func padRight(s string, targetWidth int) string {
	w := termWidth(s)
	for w < targetWidth {
		s += " "
		w++
	}
	return s
}

// termWidth returns the visual width of a string in terminal columns.
func termWidth(s string) int {
	w := 0
	for _, r := range s {
		if isCJK(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3000 && r <= 0x303F) ||
		(r >= 0x3040 && r <= 0x309F) ||
		(r >= 0x30A0 && r <= 0x30FF) ||
		(r >= 0xFF00 && r <= 0xFFEF)
}
