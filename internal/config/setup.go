package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
)

// RunSetup elindítja az interaktív setup wizard-ot
func RunSetup(cfg *Config) error {
	configfile.Install()

	fmt.Println("╔════════════════════════════════╗")
	fmt.Println("║   WorldSync – Provider Setup   ║")
	fmt.Println("╚════════════════════════════════╝")
	fmt.Println()

	provider, err := selectProvider()
	if err != nil {
		return err
	}

	name, err := prompt("Add meg a provider nevét (pl. 'Saját OneDrive'): ")
	if err != nil {
		return err
	}

	rcloneName := sanitizeName(name)

	fmt.Println()
	fmt.Println("A böngésző megnyílik a hitelesítéshez...")
	fmt.Println()

	// rclone interaktív konfiguráció futtatása
	if err := configureRclone(rcloneName, string(provider)); err != nil {
		return fmt.Errorf("rclone konfiguráció hiba: %w", err)
	}

	providerCfg := ProviderConfig{
		Name:       name,
		Type:       provider,
		RcloneName: rcloneName,
		Settings:   map[string]string{},
	}

	if err := cfg.AddProvider(providerCfg); err != nil {
		return err
	}

	// Ha ez az első provider, állítsuk be aktívnak
	if cfg.ActiveProvider == "" {
		cfg.ActiveProvider = name
	}

	if err := Save(cfg); err != nil {
		return err
	}

	fmt.Printf("\n✓ '%s' sikeresen konfigurálva!\n", name)
	return nil
}

func selectProvider() (ProviderType, error) {
	fmt.Println("Válassz cloud providert:")
	fmt.Println("  1. OneDrive")
	fmt.Println("  2. Google Drive")
	fmt.Println("  3. Proton Drive")
	fmt.Println("  4. WebDAV (Nextcloud, ownCloud, stb.)")
	fmt.Println("  5. SFTP")
	fmt.Println("  6. SMB / NAS")
	fmt.Println("  7. Lokális mappa")
	fmt.Println()

	choice, err := prompt("Választás (1-7): ")
	if err != nil {
		return "", err
	}

	switch strings.TrimSpace(choice) {
	case "1":
		return ProviderOneDrive, nil
	case "2":
		return ProviderGoogleDrive, nil
	case "3":
		return ProviderProtonDrive, nil
	case "4":
		return ProviderWebDAV, nil
	case "5":
		return ProviderSFTP, nil
	case "6":
		return ProviderSMB, nil
	case "7":
		return ProviderLocal, nil
	default:
		return "", fmt.Errorf("érvénytelen választás: %s", choice)
	}
}

func configureRclone(remoteName, providerType string) error {
	config.SetConfigPath(
		fmt.Sprintf("%s/rclone.conf", ConfigDir()),
	)

	fs := config.FileSections()
	for _, section := range fs {
		if section == remoteName {
			fmt.Printf("'%s' már létezik, felülírjuk...\n", remoteName)
			config.DeleteRemote(remoteName)
			break
		}
	}

	ctx := context.Background()
	config.NewRemote(ctx, remoteName)
	return nil
}

// prompt beolvas egy sort a terminálból
func prompt(label string) (string, error) {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// sanitizeName rclone-kompatibilis nevet készít
func sanitizeName(name string) string {
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, result)
	return result
}
