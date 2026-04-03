package cmd

import (
	"fmt"

	"github.com/QuaternionDev/worldsync/internal/launcher"
	"github.com/QuaternionDev/worldsync/internal/world"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Megjeleníti az összes világot és azok állapotát",
	Run:   runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	active := cfg.GetActiveProvider()
	if active != nil {
		fmt.Printf("Aktív provider: %s (%s)\n", active.Name, active.Type)
	} else {
		fmt.Println("Aktív provider: nincs")
	}
	fmt.Println()

	launchers := launcher.DetectAll()
	if len(launchers) == 0 {
		fmt.Println("Nem található egyetlen launcher sem.")
		return
	}

	totalWorlds := 0

	for _, l := range launchers {
		fmt.Printf("✓ %s\n", l.Name)

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
			fmt.Println("  (nincs világ)")
			continue
		}

		for _, w := range allWorlds {
			fmt.Printf("  %-30s %-20s %-10s %-10s %s\n",
				w.Name,
				w.Version,
				w.GameMode,
				w.Difficulty,
				world.FormatSize(w.SizeBytes),
			)
			totalWorlds++
		}

		fmt.Println()
	}

	fmt.Printf("Összesen: %d világ\n", totalWorlds)
}
