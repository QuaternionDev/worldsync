package cmd

import (
	"fmt"
	"os"

	"github.com/QuaternionDev/worldsync/ui/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open interactive terminal interface",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Run(cfg); err != nil {
			fmt.Printf("TUI error: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
