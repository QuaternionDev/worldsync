package main

import (
	"fmt"
	"os"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/sync"
	"github.com/QuaternionDev/worldsync/internal/world"
)

func main() {
	fmt.Println("WorldSync v0.1.0")
	fmt.Println()

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

	// Aktív provider kijelzése
	active := cfg.GetActiveProvider()
	if active != nil {
		fmt.Printf("Aktív provider: %s (%s)\n", active.Name, active.Type)
		fmt.Println()
	}

	// State és backup mappák
	stateDir := fmt.Sprintf("%s/state", appconfig.ConfigDir())
	destDir := fmt.Sprintf("%s/backup", appconfig.ConfigDir())

	engine, err := sync.NewEngine(stateDir)
	if err != nil {
		fmt.Printf("Engine hiba: %s\n", err)
		os.Exit(1)
	}

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
			engine.SyncToLocal(w.Path, destDir)
		}

		fmt.Println()
	}

	fmt.Printf("Kész! Backup helye: %s\n", destDir)
}
