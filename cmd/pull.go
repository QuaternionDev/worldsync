package cmd

import (
	"fmt"
	"os"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/storage"
	worldsync "github.com/QuaternionDev/worldsync/internal/sync"
	"github.com/rclone/rclone/fs/config"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Letölti a cloud-on lévő világokat lokálisan",
	Long:  `Megkeresi a cloud-on lévő világokat és letölti azokat amelyek lokálisan nem léteznek.`,
	Run:   runPull,
}

func init() {
	pullCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Adott provider használata (név)")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) {
	// Provider kiválasztása
	var active *appconfig.ProviderConfig

	if providerFlag != "" {
		for i, p := range cfg.Providers {
			if p.Name == providerFlag {
				active = &cfg.Providers[i]
				break
			}
		}
		if active == nil {
			fmt.Printf("Nem található '%s' nevű provider\n", providerFlag)
			os.Exit(1)
		}
	} else {
		active = cfg.GetActiveProvider()
	}

	if active == nil {
		fmt.Println("Nincs konfigurált provider. Futtasd: worldsync provider add")
		os.Exit(1)
	}

	fmt.Printf("Provider: %s (%s)\n", active.Name, active.Type)
	fmt.Println()

	// Launchers keresése – meghatározzuk hova töltjük le
	fmt.Println("Launchers keresése...")
	launchers := launcher.DetectAll()

	if len(launchers) == 0 {
		fmt.Println("Nem található egyetlen launcher sem.")
		return
	}

	// Megkérdezzük melyik saves mappába töltse le
	fmt.Println("\nHova töltse le a világokat?")
	for i, l := range launchers {
		if len(l.SavePaths) > 0 {
			fmt.Printf("  %d. %s (%s)\n", i+1, l.Name, l.SavePaths[0])
		}
	}
	fmt.Println()

	var choice int
	fmt.Print("Választás: ")
	fmt.Scan(&choice)

	if choice < 1 || choice > len(launchers) {
		fmt.Println("Érvénytelen választás.")
		os.Exit(1)
	}

	selectedLauncher := launchers[choice-1]
	if len(selectedLauncher.SavePaths) == 0 {
		fmt.Println("Nincs saves mappa a kiválasztott launchernél.")
		os.Exit(1)
	}

	savesPath := selectedLauncher.SavePaths[0]
	fmt.Printf("\nCélmappa: %s\n\n", savesPath)

	// Pull logika
	stateDir := fmt.Sprintf("%s/state", appconfig.ConfigDir())
	engine, err := worldsync.NewEngine(stateDir)
	if err != nil {
		fmt.Printf("Engine hiba: %s\n", err)
		os.Exit(1)
	}

	switch active.Type {
	case appconfig.ProviderLocal:
		destDir, _ := config.FileGetValue(active.RcloneName, "root")
		if destDir == "" {
			fmt.Println("Nincs megadva forrás mappa.")
			os.Exit(1)
		}
		if err := engine.PullFromLocal(destDir, savesPath); err != nil {
			fmt.Printf("Pull hiba: %s\n", err)
			os.Exit(1)
		}

	default:
		provider, err := storage.NewRcloneProvider(
			active.Name,
			active.RcloneName,
			"WorldSync/worlds",
		)
		if err != nil {
			fmt.Printf("Provider hiba: %s\n", err)
			os.Exit(1)
		}

		// Remote világok listázása
		fmt.Println("Remote világok keresése...")
		worlds, err := provider.ListWorlds()
		if err != nil {
			fmt.Printf("Listázási hiba: %s\n", err)
			os.Exit(1)
		}

		if len(worlds) == 0 {
			fmt.Println("Nem található világ a remote-on.")
			return
		}

		fmt.Printf("%d világ található a remote-on.\n\n", len(worlds))

		downloaded := 0
		for _, worldName := range worlds {
			fmt.Printf("  ↓ %s letöltése...\n", worldName)
			if err := provider.PullWorld(worldName, savesPath); err != nil {
				fmt.Printf("  ✗ %s: %s\n", worldName, err)
				continue
			}
			fmt.Printf("  ✓ %s letöltve\n", worldName)
			downloaded++
		}

		fmt.Printf("\n%d/%d világ letöltve.\n", downloaded, len(worlds))
	}
}
