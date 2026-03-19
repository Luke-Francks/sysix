package tui

import (
	"fmt"
	"time"

	"github.com/System9-Software/sysix/internal/collector"
	"github.com/System9-Software/sysix/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `  ____            _      
 / ___| _   _ ___(_)_  __
 \___ \| | | / __| \ \/ /
  ___) | |_| \__ \  |  |
 |____/ \__, |___/_/_/\_\
        |___/            `

var (
	colorAccent = lipgloss.Color("#4DA8FF")
	colorText   = lipgloss.Color("#E8F0FF")
	colorMuted  = lipgloss.Color("#6B7C99")
	colorBorder = lipgloss.Color("#1E2D45")
	colorGood   = lipgloss.Color("#39D98A")
	colorWarn   = lipgloss.Color("#F5A623")
	colorBad    = lipgloss.Color("#FF5B5B")

	logoStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(10)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorText)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Width(45)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

type tickMsg time.Time

type model struct {
	snapshot *collector.SystemSnapshot
	network  *collector.NetworkStats
	ports    []collector.Port
	cfg      *config.Config
	width    int
	height   int
	err      error
}

func tick(rate int) tea.Cmd {
	return tea.Tick(time.Duration(rate)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {
	cfg, _ := config.Load()
	snapshot, err := collector.GetSnapshot()
	network, _ := collector.GetNetwork()
	ports, _ := collector.GetPorts()
	return model{
		snapshot: snapshot,
		network:  network,
		ports:    ports,
		cfg:      cfg,
		err:      err,
	}
}

func (m model) Init() tea.Cmd {
	return tick(m.cfg.RefreshRate)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.snapshot, m.err = collector.GetSnapshot()
		m.network, _ = collector.GetNetwork()
		m.ports, _ = collector.GetPorts()
		return m, tick(m.cfg.RefreshRate)
	}
	return m, nil
}

func statusColor(percent float64) lipgloss.Color {
	if percent >= 90 {
		return colorBad
	} else if percent >= 70 {
		return colorWarn
	}
	return colorGood
}

func formatBytes(b uint64) string {
	if b >= 1024*1024*1024 {
		return fmt.Sprintf("%.1f GB", float64(b)/1024/1024/1024)
	} else if b >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	} else if b >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%d B", b)
}

func row(label, value string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render(label),
		valueStyle.Render(value),
	) + "\n"
}

func colorRow(label, value string, color lipgloss.Color) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render(label),
		lipgloss.NewStyle().Foreground(color).Render(value),
	) + "\n"
}

func (m model) renderSystem() string {
	s := m.snapshot
	content := panelTitleStyle.Render("[ System ]") + "\n"
	content += row("Host:", fmt.Sprintf("%s (%s)", s.Hostname, s.OS))
	content += row("Uptime:", fmt.Sprintf("%d hours", s.Uptime/3600))
	content += colorRow("CPU:", fmt.Sprintf("%.1f%%", s.CPUPercent), statusColor(s.CPUPercent))
	content += colorRow("Memory:", fmt.Sprintf("%.1f%% (%d MB / %d MB)", s.MemPercent, s.MemUsed/1024/1024, s.MemTotal/1024/1024), statusColor(s.MemPercent))
	content += colorRow("Disk:", fmt.Sprintf("%.1f%% (%d GB / %d GB)", s.DiskPercent, s.DiskUsed/1024/1024/1024, s.DiskTotal/1024/1024/1024), statusColor(s.DiskPercent))
	return panelStyle.Render(content)
}

func (m model) renderNetwork() string {
	content := panelTitleStyle.Render("[ Network ]") + "\n"
	if m.network != nil {
		content += row("Sent:", formatBytes(m.network.BytesSent))
		content += row("Received:", formatBytes(m.network.BytesRecv))
		content += row("Pkts Out:", fmt.Sprintf("%d", m.network.PacketsSent))
		content += row("Pkts In:", fmt.Sprintf("%d", m.network.PacketsRecv))
	} else {
		content += valueStyle.Render("unavailable")
	}
	return panelStyle.Render(content)
}

func (m model) renderProcesses() string {
	content := panelTitleStyle.Render("[ Processes ]") + "\n"
	content += lipgloss.NewStyle().Foreground(colorMuted).Render(
		fmt.Sprintf("%-6s %-22s %6s %8s", "PID", "Name", "CPU%", "Mem"),
	) + "\n"
	count := 0
	for _, p := range m.snapshot.Processes {
		if p.MemMB > 1 && count < 8 {
			content += valueStyle.Render(
				fmt.Sprintf("%-6d %-22s %5.1f%% %6.0fMB", p.PID, truncate(p.Name, 22), p.CPUPercent, p.MemMB),
			) + "\n"
			count++
		}
	}
	return panelStyle.Width(50).Render(content)
}

func (m model) renderPorts() string {
	content := panelTitleStyle.Render("[ Ports ]") + "\n"
	content += lipgloss.NewStyle().Foreground(colorMuted).Render(
		fmt.Sprintf("%-8s %-6s %-10s %s", "Port", "Type", "Status", "PID"),
	) + "\n"
	count := 0
	for _, p := range m.ports {
		if count < 8 {
			content += valueStyle.Render(
				fmt.Sprintf("%-8d %-6s %-10s %d", p.Port, p.Type, p.Status, p.PID),
			) + "\n"
			count++
		}
	}
	return panelStyle.Width(50).Render(content)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-2] + ".."
	}
	return s
}

func (m model) renderQuickRef() string {
	content := panelTitleStyle.Render("[ Quick Reference ]") + "\n"
	content += row("q", "quit sysix")
	content += row("ctrl+c", "force quit")
	content += row("watch", "launch TUI")
	content += row("serve", "launch web UI")
	return panelStyle.Width(45).Render(content)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}
	if m.width == 0 {
		return "loading..."
	}

	header := logoStyle.Render(logo)

	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		m.renderSystem(),
		m.renderNetwork(),
		m.renderQuickRef(),
	)

	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		m.renderProcesses(),
		m.renderPorts(),
	)

	dashboard := lipgloss.JoinHorizontal(lipgloss.Top,
		leftCol,
		rightCol,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		dashboard,
	)
}

func Start() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
