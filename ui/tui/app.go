package tui

import (
	"fmt"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/world"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab indexek
const (
	tabWorlds   = 0
	tabProvider = 1
	tabSettings = 2
)

// Stílusok
var (
	colorPrimary   = lipgloss.Color("#B48EAD")
	colorSecondary = lipgloss.Color("#81A1C1")
	colorMuted     = lipgloss.Color("#4C566A")
	colorError     = lipgloss.Color("#BF616A")
	colorSelected  = lipgloss.Color("#D8A9F0")
	colorBorder    = lipgloss.Color("#6B4F8E")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	styleTab = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorMuted)

	styleActiveTab = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorSelected).
			Bold(true).
			Underline(true)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorSelected).
			Bold(true)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleError = lipgloss.NewStyle().
			Foreground(colorError)

	styleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#2E1F47")).
			Foreground(lipgloss.Color("#9B7FBF")).
			Padding(0, 1)
)

// WorldItem egy világ listában való megjelenítéséhez
type WorldItem struct {
	World        world.World
	LauncherName string
}

// Model a TUI fő modellje
type Model struct {
	cfg       *appconfig.Config
	activeTab int
	worlds    []WorldItem
	cursor    int
	syncing   bool
	syncMsg   string
	err       error
	width     int
	height    int
	loading   bool
}

// Üzenetek
type worldsLoadedMsg struct {
	worlds []WorldItem
}

type syncDoneMsg struct {
	err error
}

type syncProgressMsg struct {
	msg string
}

// NewModel létrehoz egy új TUI modellt
func NewModel(cfg *appconfig.Config) Model {
	return Model{
		cfg:     cfg,
		loading: true,
	}
}

// Init inicializálja a TUI-t
func (m Model) Init() tea.Cmd {
	return loadWorldsCmd()
}

// loadWorldsCmd betölti a világokat
func loadWorldsCmd() tea.Cmd {
	return func() tea.Msg {
		var items []WorldItem

		launchers := launcher.DetectAll()
		for _, l := range launchers {
			var allWorlds []world.World
			for _, savesPath := range l.SavePaths {
				instanceName := l.InstanceNames[savesPath]
				worlds, err := world.ScanWorlds(savesPath, instanceName)
				if err != nil {
					continue
				}
				allWorlds = append(allWorlds, worlds...)
			}
			allWorlds = world.DeduplicateWorlds(allWorlds)
			for _, w := range allWorlds {
				items = append(items, WorldItem{
					World:        w,
					LauncherName: l.Name,
				})
			}
		}

		return worldsLoadedMsg{worlds: items}
	}
}

// Update kezeli az eseményeket
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case worldsLoadedMsg:
		m.worlds = msg.worlds
		m.loading = false

	case syncDoneMsg:
		m.syncing = false
		if msg.err != nil {
			m.err = msg.err
			m.syncMsg = fmt.Sprintf("Error: %s", msg.err)
		} else {
			m.syncMsg = "✓ Sync complete!"
		}

	case syncProgressMsg:
		m.syncMsg = msg.msg

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.activeTab = (m.activeTab + 1) % 3

		case "shift+tab":
			m.activeTab = (m.activeTab + 2) % 3

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.worlds)-1 {
				m.cursor++
			}

		case "s":
			if !m.syncing && m.activeTab == tabWorlds {
				m.syncing = true
				m.syncMsg = "Syncing..."
				return m, syncAllCmd(m.cfg)
			}

		case "r":
			if m.activeTab == tabWorlds {
				m.loading = true
				m.worlds = nil
				return m, loadWorldsCmd()
			}
		}
	}

	return m, nil
}

// syncAllCmd szinkronizálja az összes világot
func syncAllCmd(cfg *appconfig.Config) tea.Cmd {
	return func() tea.Msg {
		active := cfg.GetActiveProvider()
		if active == nil {
			return syncDoneMsg{err: fmt.Errorf("no active provider")}
		}

		launchers := launcher.DetectAll()
		for _, l := range launchers {
			var allWorlds []world.World
			for _, savesPath := range l.SavePaths {
				instanceName := l.InstanceNames[savesPath]
				worlds, err := world.ScanWorlds(savesPath, instanceName)
				if err != nil {
					continue
				}
				allWorlds = append(allWorlds, worlds...)
			}
			allWorlds = world.DeduplicateWorlds(allWorlds)

			for _, w := range allWorlds {
				_ = w
				// TODO: sync logika itt
			}
		}

		return syncDoneMsg{err: nil}
	}
}

// View rendereli a TUI-t
func (m Model) View() string {
	if m.width == 0 {
		return "Betöltés..."
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s",
		m.renderHeader(),
		m.renderTabs(),
		m.renderContent(),
		m.renderStatusBar(),
	)
}

func (m Model) renderHeader() string {
	title := styleTitle.Render("⛏  WorldSync")

	provider := ""
	if active := m.cfg.GetActiveProvider(); active != nil {
		provider = styleMuted.Render(fmt.Sprintf("Provider: %s (%s)", active.Name, active.Type))
	} else {
		provider = styleError.Render("No provider configured")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		title,
		"  ",
		provider,
	)
}

func (m Model) renderTabs() string {
	tabs := []string{"Worlds", "Provider", "Settings"}
	rendered := []string{}

	for i, tab := range tabs {
		if i == m.activeTab {
			rendered = append(rendered, styleActiveTab.Render(tab))
		} else {
			rendered = append(rendered, styleTab.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderContent() string {
	switch m.activeTab {
	case tabWorlds:
		return m.renderWorlds()
	case tabProvider:
		return m.renderProvider()
	case tabSettings:
		return m.renderSettings()
	}
	return ""
}

func (m Model) renderWorlds() string {
	if m.loading {
		return styleBox.Render(styleMuted.Render("Világok betöltése..."))
	}

	if len(m.worlds) == 0 {
		return styleBox.Render(styleMuted.Render("Nem található egyetlen világ sem."))
	}

	content := ""
	currentLauncher := ""

	for i, item := range m.worlds {
		// Launcher fejléc ha új launcher
		if item.LauncherName != currentLauncher {
			if currentLauncher != "" {
				content += "\n"
			}
			currentLauncher = item.LauncherName
			content += styleMuted.Render(fmt.Sprintf("  %s\n", currentLauncher))
		}

		// Világ sor
		cursor := "  "
		name := item.World.Name
		info := fmt.Sprintf("%-12s %-10s %s",
			item.World.Version,
			item.World.GameMode,
			world.FormatSize(item.World.SizeBytes),
		)

		// Modpack megjelenítése ha van
		modpack := ""
		if item.World.ModpackName != "" {
			modpack = styleMuted.Render(fmt.Sprintf(" [%s]", item.World.ModpackName))
		}

		if i == m.cursor {
			cursor = styleSelected.Render("▶ ")
			name = styleSelected.Render(fmt.Sprintf("%-28s", name))
			info = styleSelected.Render(info)
			if item.World.ModpackName != "" {
				modpack = styleSelected.Render(fmt.Sprintf(" [%s]", item.World.ModpackName))
			}
		} else {
			name = fmt.Sprintf("%-28s", name)
			info = styleMuted.Render(info)
		}

		content += fmt.Sprintf("  %s%s%s  %s\n", cursor, name, modpack, info)
	}

	// Sync üzenet
	if m.syncMsg != "" {
		if m.err != nil {
			content += "\n" + styleError.Render(m.syncMsg)
		} else {
			content += "\n" + styleSelected.Render(m.syncMsg)
		}
	}

	return styleBox.
		Width(m.width - 4).
		Render(content)
}

func (m Model) renderProvider() string {
	if len(m.cfg.Providers) == 0 {
		return styleBox.Render(
			styleMuted.Render("Provider is not configured.\n\n") +
				"Press [A] to add a new provider.",
		)
	}

	content := ""
	for _, p := range m.cfg.Providers {
		active := ""
		if p.Name == m.cfg.ActiveProvider {
			active = styleSelected.Render(" ← aktív")
		}
		content += fmt.Sprintf("  • %s (%s)%s\n", p.Name, p.Type, active)
	}

	return styleBox.
		Width(m.width - 4).
		Render(content)
}

func (m Model) renderSettings() string {
	content := fmt.Sprintf(
		"  Snapshot megőrzés:  %d\n"+
			"  Sync kilépéskor:    %v\n"+
			"  Config directory:       %s\n",
		m.cfg.KeepSnapshots,
		m.cfg.SyncOnExit,
		appconfig.ConfigDir(),
	)

	return styleBox.
		Width(m.width - 4).
		Render(content)
}

func (m Model) renderStatusBar() string {
	keys := "[Tab] Navigáció  [↑/↓] Mozgás  [S] Sync  [R] Frissítés  [Q] Kilépés"
	if m.syncing {
		keys = styleMuted.Render("Sync in progress...")
	}
	return styleStatusBar.
		Width(m.width).
		Render(keys)
}

// Run elindítja a TUI-t
func Run(cfg *appconfig.Config) error {
	m := NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
