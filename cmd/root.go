package cmd

import (
	"fmt"
	"os"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/spf13/cobra"
)

var cfg *appconfig.Config

var rootCmd = &cobra.Command{
	Use:   "worldsync",
	Short: "Minecraft world sync tool",
	Long: `WorldSync – Minecraft Java Edition cloud sync tool.
	Sync your worlds to OneDrive, Google Drive,
	Proton Drive, SFTP/SMB servers, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	configfile.Install()
	config.SetConfigPath(
		fmt.Sprintf("%s/rclone.conf", appconfig.ConfigDir()),
	)

	var err error
	cfg, err = appconfig.Load()
	if err != nil {
		fmt.Printf("Configuration error: %s\n", err)
		os.Exit(1)
	}
}
