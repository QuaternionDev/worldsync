package cmd

import (
	"fmt"
	"os"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/storage"
	"github.com/QuaternionDev/worldsync/internal/sync"
	"github.com/QuaternionDev/worldsync/internal/world"
	"github.com/rclone/rclone/fs/config"
	"github.com/spf13/cobra"
)

var providerFlag string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Szinkronizálja az összes világot",
	Long:  `Megkeresi az összes Minecraft világot és szinkronizálja a konfigurált cloud providerrel.`,
	Run:   runSync,
}

func init() {
	syncCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Adott provider használata (név)")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
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

	// Engine inicializálása
	stateDir := fmt.Sprintf("%s/state", appconfig.ConfigDir())
	engine, err := sync.NewEngine(stateDir)
	if err != nil {
		fmt.Printf("Engine hiba: %s\n", err)
		os.Exit(1)
	}

	// Launchers keresése
	fmt.Println("Launchers keresése...")
	launchers := launcher.DetectAll()

	if len(launchers) == 0 {
		fmt.Println("Nem található egyetlen launcher sem.")
		return
	}

	totalWorlds := 0
	syncedWorlds := 0

	for _, l := range launchers {
		fmt.Printf("\n✓ %s\n", l.Name)

		var allWorlds []world.World
		for _, savesPath := range l.SavePaths {
			worlds, err := world.ScanWorlds(savesPath)
			if err != nil {
				continue
			}
			allWorlds = append(allWorlds, worlds...)
		}

		allWorlds = world.DeduplicateWorlds(allWorlds)

		if len(allWorlds) == 0 {
			fmt.Println("  (nincs világ)")
			continue
		}

		totalWorlds += len(allWorlds)

		for _, w := range allWorlds {
			fmt.Printf("  → %s szinkronizálása...\n", w.Name)

			switch active.Type {
			case appconfig.ProviderLocal:
				destDir, _ := config.FileGetValue(active.RcloneName, "root")
				if destDir == "" {
					fmt.Printf("  ✗ Nincs megadva célmappa\n")
					continue
				}
				if _, err := engine.SyncToLocal(w.Path, destDir); err != nil {
					fmt.Printf("  ✗ Hiba: %s\n", err)
					continue
				}

			default:
				provider, err := storage.NewRcloneProvider(
					active.Name,
					active.RcloneName,
					"WorldSync/worlds",
				)
				if err != nil {
					fmt.Printf("  ✗ Provider hiba: %s\n", err)
					continue
				}
				if err := provider.SyncWorld(w.Path, w.Name); err != nil {
					fmt.Printf("  ✗ Sync hiba: %s\n", err)
					continue
				}
			}

			fmt.Printf("  ✓ %s kész\n", w.Name)
			syncedWorlds++
		}
	}

	fmt.Printf("\n%d/%d világ szinkronizálva.\n", syncedWorlds, totalWorlds)
}
