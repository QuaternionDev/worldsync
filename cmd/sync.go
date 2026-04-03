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
	Short: "Sync all worlds",
	Long:  `Finds all Minecraft worlds and syncs them with the configured cloud provider.`,
	Run:   runSync,
}

func init() {
	syncCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Use a specific provider (name from 'worldsync provider list')")
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
			fmt.Printf("No provider found with name '%s'\n", providerFlag)
			os.Exit(1)
		}
	} else {
		active = cfg.GetActiveProvider()
	}

	if active == nil {
		fmt.Println("No configured provider. Run: worldsync provider add")
		os.Exit(1)
	}

	fmt.Printf("Provider: %s (%s)\n", active.Name, active.Type)
	fmt.Println()

	// Engine inicializálása
	stateDir := fmt.Sprintf("%s/state", appconfig.ConfigDir())
	engine, err := sync.NewEngine(stateDir)
	if err != nil {
		fmt.Printf("Engine error: %s\n", err)
		os.Exit(1)
	}

	// Launchers keresése
	fmt.Println("Searching for launchers...")
	launchers := launcher.DetectAll()

	if len(launchers) == 0 {
		fmt.Println("No launchers found.")
		return
	}

	totalWorlds := 0
	syncedWorlds := 0

	for _, l := range launchers {
		fmt.Printf("\n✓ %s\n", l.Name)

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

		if len(allWorlds) == 0 {
			fmt.Println("  (no worlds found)")
			continue
		}

		totalWorlds += len(allWorlds)

		for _, w := range allWorlds {
			fmt.Printf("  → %s syncing...\n", w.Name)

			switch active.Type {
			case appconfig.ProviderLocal:
				destDir, _ := config.FileGetValue(active.RcloneName, "root")
				if destDir == "" {
					fmt.Printf("  ✗ No destination directory specified\n")
					continue
				}
				if _, err := engine.SyncToLocal(w.Path, destDir); err != nil {
					fmt.Printf("  ✗ Error: %s\n", err)
					continue
				}

			default:
				provider, err := storage.NewRcloneProvider(
					active.Name,
					active.RcloneName,
					"WorldSync/worlds",
				)
				if err != nil {
					fmt.Printf("  ✗ Provider error: %s\n", err)
					continue
				}
				if err := provider.SyncWorld(w.Path, w.Name); err != nil {
					fmt.Printf("  ✗ Sync error: %s\n", err)
					continue
				}
			}

			fmt.Printf("  ✓ %s done\n", w.Name)
			syncedWorlds++
		}
	}

	fmt.Printf("\n%d/%d worlds synced.\n", syncedWorlds, totalWorlds)
}
