package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/whitehai11/AWaN-Installer/installer"
)

type progressMsg struct {
	step  installer.Step
	index int
	total int
}

type installDoneMsg struct {
	result *installer.Result
	err    error
}

type tuiModel struct {
	installer     *installer.Installer
	selected      int
	options       []string
	progressCh    chan tea.Msg
	installing    bool
	stepIndex     int
	stepTotal     int
	currentStep   string
	status        string
	errMessage    string
	result        *installer.Result
	titleStyle    lipgloss.Style
	activeStyle   lipgloss.Style
	plainStyle    lipgloss.Style
	panelStyle    lipgloss.Style
	errorStyle    lipgloss.Style
	successStyle  lipgloss.Style
}

// RunInstaller runs the Bubble Tea installer flow.
func RunInstaller(flow *installer.Installer) error {
	model := tuiModel{
		installer: flow,
		options:   []string{"Install GUI version", "Install TUI version"},
		status:    "Choose which AWaN interface to install",
		titleStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		activeStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("29")).Padding(0, 1),
		plainStyle: lipgloss.NewStyle().Padding(0, 1),
		panelStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
		errorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
	}

	_, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	return err
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if !m.installing && m.selected > 0 {
				m.selected--
			}
		case "down":
			if !m.installing && m.selected < len(m.options)-1 {
				m.selected++
			}
		case "enter":
			if m.result != nil {
				return m, tea.Quit
			}
			if m.installing {
				return m, nil
			}

			m.installing = true
			m.errMessage = ""
			m.status = "Starting installation..."
			m.progressCh = make(chan tea.Msg, 16)
			return m, tea.Batch(m.installCmd(), m.listenProgress())
		}

	case progressMsg:
		m.stepIndex = msg.index + 1
		m.stepTotal = msg.total
		m.currentStep = msg.step.Title
		m.status = fmt.Sprintf("Step %d/%d: %s", m.stepIndex, m.stepTotal, msg.step.Title)
		return m, m.listenProgress()

	case installDoneMsg:
		m.installing = false
		m.progressCh = nil
		if msg.err != nil {
			m.errMessage = msg.err.Error()
			m.status = "Installation failed"
			return m, nil
		}

		m.result = msg.result
		m.status = "AWaN installed successfully"
		return m, nil
	}

	return m, nil
}

func (m tuiModel) View() string {
	lines := []string{
		m.titleStyle.Render("AWaN Terminal Installer"),
		"",
		"Select installation target:",
		"",
	}

	for index, option := range m.options {
		label := option
		if index == m.selected && !m.installing && m.result == nil {
			label = m.activeStyle.Render(label)
		} else {
			label = m.plainStyle.Render(label)
		}
		lines = append(lines, label)
	}

	if m.installing {
		lines = append(lines, "", "Installation progress:")
		for index, step := range m.installer.Steps(selectedTarget(m.selected)) {
			prefix := "  "
			if index+1 == m.stepIndex {
				prefix = "> "
			} else if index+1 < m.stepIndex {
				prefix = "x "
			}
			lines = append(lines, prefix+step.Title)
		}
	}

	if m.result != nil {
		lines = append(lines, "", m.successStyle.Render("Installation complete"))
		lines = append(lines, "Target: "+m.result.Target, "Location: "+m.result.RootPath)
		lines = append(lines, "", "Press Enter to exit.")
	}

	if m.errMessage != "" {
		lines = append(lines, "", m.errorStyle.Render(m.errMessage))
	}

	lines = append(lines, "", m.status)
	return m.panelStyle.Render(strings.Join(lines, "\n"))
}

func (m tuiModel) installCmd() tea.Cmd {
	target := selectedTarget(m.selected)
	progressCh := m.progressCh

	return func() tea.Msg {
		result, err := m.installer.Run(target, func(step installer.Step, index, total int) {
			progressCh <- progressMsg{step: step, index: index, total: total}
		})
		close(progressCh)
		return installDoneMsg{result: result, err: err}
	}
}

func (m tuiModel) listenProgress() tea.Cmd {
	progressCh := m.progressCh
	return func() tea.Msg {
		if progressCh == nil {
			return nil
		}
		msg, ok := <-progressCh
		if !ok {
			return nil
		}
		return msg
	}
}

func selectedTarget(index int) string {
	if index == 0 {
		return installer.TargetGUI
	}
	return installer.TargetTUI
}
