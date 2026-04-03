package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/sync"
	"github.com/QuaternionDev/worldsync/internal/world"
)

func main() {
	fmt.Println("WorldSync v0.1.0")
	fmt.Println()

	// State mappa: %APPDATA%\WorldSync\state
	stateDir := filepath.Join(os.Getenv("APPDATA"), "WorldSync", "state")
	// Célmappa: %APPDATA%\WorldSync\backup
	destDir := filepath.Join(os.Getenv("APPDATA"), "WorldSync", "backup")

	engine, err := sync.NewEngine(stateDir)
	if err != nil {
		fmt.Printf("Engine hiba: %s\n", err)
		return
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
