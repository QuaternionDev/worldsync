package cmd

import (
	"fmt"
	"os"

	appconfig "github.com/QuaternionDev/worldsync/internal/config"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Provider kezelés",
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Konfigurált providerek listája",
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.Providers) == 0 {
			fmt.Println("Nincs konfigurált provider.")
			return
		}

		for _, p := range cfg.Providers {
			active := ""
			if p.Name == cfg.ActiveProvider {
				active = " ← aktív"
			}
			fmt.Printf("  • %s (%s)%s\n", p.Name, p.Type, active)
		}
	},
}

var providerAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Új provider hozzáadása",
	Run: func(cmd *cobra.Command, args []string) {
		if err := appconfig.RunSetup(cfg); err != nil {
			fmt.Printf("Hiba: %s\n", err)
			os.Exit(1)
		}
	},
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove [név]",
	Short: "Provider törlése",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		found := false

		newProviders := []appconfig.ProviderConfig{}
		for _, p := range cfg.Providers {
			if p.Name == name {
				found = true
				continue
			}
			newProviders = append(newProviders, p)
		}

		if !found {
			fmt.Printf("Nem található '%s' nevű provider\n", name)
			os.Exit(1)
		}

		cfg.Providers = newProviders
		if cfg.ActiveProvider == name {
			cfg.ActiveProvider = ""
			if len(cfg.Providers) > 0 {
				cfg.ActiveProvider = cfg.Providers[0].Name
			}
		}

		if err := appconfig.Save(cfg); err != nil {
			fmt.Printf("Mentési hiba: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ '%s' törölve\n", name)
	},
}

var providerSetCmd = &cobra.Command{
	Use:   "set [név]",
	Short: "Aktív provider beállítása",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		found := false

		for _, p := range cfg.Providers {
			if p.Name == name {
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Nem található '%s' nevű provider\n", name)
			os.Exit(1)
		}

		cfg.ActiveProvider = name
		if err := appconfig.Save(cfg); err != nil {
			fmt.Printf("Mentési hiba: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Aktív provider: %s\n", name)
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerSetCmd)
	rootCmd.AddCommand(providerCmd)
}
