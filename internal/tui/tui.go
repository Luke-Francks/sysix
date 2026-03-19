package tui

import (
	"fmt"

	"github.com/Luke-Francks/sysix/internal/collector"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = ` ______________.___. _________._______  ___
 /   _____/\__  |   |/   _____/|   \   \/  /
 \_____  \  /   |   |\_____  \ |   |\     / 
 /        \ \____   |/        \|   |/     \ 
/_______  / / ______/_______  /|___/___/\  \
        \/  \/              \/           \_/`

var (
	colorAccent = lipgloss.Color("#4DA8FF")
	colorText   = lipgloss.Color("#E8F0FF")
	colorMuted  = lipgloss.Color("#6B7C99")
	colorBorder = lipgloss.Color("#1E2D45")

	logoStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			PaddingBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorText)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)
)

type model struct {
	snapshot *collector.SystemSnapshot
	err      error
}

func initialModel() model {
	snapshot, err := collector.GetSnapshot()
	return model{snapshot: snapshot, err: err}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}

	s := m.snapshot

	content := titleStyle.Render("observer") + "\n"
	content += labelStyle.Render("Host:   ") + valueStyle.Render(fmt.Sprintf("%s (%s)", s.Hostname, s.OS)) + "\n"
	content += labelStyle.Render("Uptime: ") + valueStyle.Render(fmt.Sprintf("%d hours", s.Uptime/3600)) + "\n"
	content += labelStyle.Render("CPU:    ") + valueStyle.Render(fmt.Sprintf("%.1f%%", s.CPUPercent)) + "\n"
	content += labelStyle.Render("Memory: ") + valueStyle.Render(fmt.Sprintf("%.1f%% (%d MB / %d MB)", s.MemPercent, s.MemUsed/1024/1024, s.MemTotal/1024/1024)) + "\n"
	content += labelStyle.Render("Disk:   ") + valueStyle.Render(fmt.Sprintf("%.1f%% (%d GB / %d GB)", s.DiskPercent, s.DiskUsed/1024/1024/1024, s.DiskTotal/1024/1024/1024)) + "\n"
	content += "\n" + labelStyle.Render("press q to quit")

	return logoStyle.Render(logo) + "\n" + borderStyle.Render(content)
}

func Start() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
