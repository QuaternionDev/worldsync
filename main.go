package main

import (
	"fmt"
	"os"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/storage"
	"github.com/QuaternionDev/worldsync/internal/sync"
	"github.com/QuaternionDev/worldsync/internal/world"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
)

func main() {
	fmt.Println("WorldSync v0.1.0")
	fmt.Println()

	// rclone config inicializálása
	configfile.Install()
	config.SetConfigPath(
		fmt.Sprintf("%s/rclone.conf", appconfig.ConfigDir()),
	)

	// Konfiguráció betöltése
	cfg, err := appconfig.Load()
	if err != nil {
		fmt.Printf("Konfiguráció hiba: %s\n", err)
		os.Exit(1)
	}

	// Ha nincs provider konfigurálva, futtatjuk a setup wizard-ot
	if len(cfg.Providers) == 0 {
		fmt.Println("Nincs konfigurált provider. Indítjuk a setup wizard-ot...")
		fmt.Println()
		if err := appconfig.RunSetup(cfg); err != nil {
			fmt.Printf("Setup hiba: %s\n", err)
			os.Exit(1)
		}
		fmt.Println()
	}

	// Aktív provider
	active := cfg.GetActiveProvider()
	if active == nil {
		fmt.Println("Nincs aktív provider!")
		os.Exit(1)
	}

	fmt.Printf("Aktív provider: %s (%s)\n", active.Name, active.Type)
	fmt.Println()

	// State mappa
	stateDir := fmt.Sprintf("%s/state", appconfig.ConfigDir())
	engine, err := sync.NewEngine(stateDir)
	if err != nil {
		fmt.Printf("Engine hiba: %s\n", err)
		os.Exit(1)
	}

	// Launchers keresése
	fmt.Println("Launchers keresése...")
	fmt.Println()

	launchers := launcher.DetectAll()
	if len(launchers) == 0 {
		fmt.Println("Nem található egyetlen launcher sem.")
		return
	}

	for _, l := range launchers {
		fmt.Printf("✓ %s\n", l.Name)

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
			fmt.Println()
			continue
		}

		fmt.Printf("  %d világ szinkronizálása...\n", len(allWorlds))

		for _, w := range allWorlds {
			switch active.Type {
			case appconfig.ProviderLocal:
				// Lokális: az rclone.conf-ban tárolt root mappába
				destDir, _ := config.FileGetValue(active.RcloneName, "root")
				if destDir == "" {
					fmt.Printf("  ✗ %s: nincs megadva célmappa\n", w.Name)
					continue
				}
				engine.SyncToLocal(w.Path, destDir)

			default:
				// Cloud: rclone sync
				provider, err := storage.NewRcloneProvider(
					active.Name,
					active.RcloneName,
					"WorldSync/worlds",
				)
				if err != nil {
					fmt.Printf("  ✗ %s: provider hiba: %s\n", w.Name, err)
					continue
				}
				if err := provider.SyncWorld(w.Path, w.Name); err != nil {
					fmt.Printf("  ✗ %s: sync hiba: %s\n", w.Name, err)
					continue
				}
				fmt.Printf("  ✓ %s szinkronizálva\n", w.Name)
			}
		}

		fmt.Println()
	}

	fmt.Println("Kész!")
}
