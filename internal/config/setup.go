package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/fs/config/obscure"
)

// RunSetup elindítja az interaktív setup wizard-ot
func RunSetup(cfg *Config) error {
	configfile.Install()

	config.SetConfigPath(
		fmt.Sprintf("%s/rclone.conf", ConfigDir()),
	)

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
	// Ha már létezik, töröljük
	for _, section := range config.FileSections() {
		if section == remoteName {
			config.DeleteRemote(remoteName)
			break
		}
	}

	ctx := context.Background()

	switch providerType {
	case "onedrive":
		config.FileSetValue(remoteName, "type", "onedrive")
		config.SaveConfig()
		fmt.Println("Megnyílik a böngésző a Microsoft hitelesítéshez...")
		fmt.Println("Ha nem nyílik meg automatikusan, másold be a megjelenő URL-t.")
		fmt.Println()
		config.NewRemote(ctx, remoteName)

	case "drive":
		config.FileSetValue(remoteName, "type", "drive")
		config.FileSetValue(remoteName, "scope", "drive")
		config.SaveConfig()
		fmt.Println("Megnyílik a böngésző a Google hitelesítéshez...")
		fmt.Println("Ha nem nyílik meg automatikusan, másold be a megjelenő URL-t.")
		fmt.Println()
		config.NewRemote(ctx, remoteName)

	case "protondrive":
		email, _ := prompt("Proton email: ")
		pass, _ := promptPassword("Proton jelszó: ")
		twofa, _ := prompt("2FA kód (ha van, különben Enter): ")

		obscuredPass, err := obscure.Obscure(pass)
		if err != nil {
			return fmt.Errorf("jelszó titkosítása sikertelen: %w", err)
		}

		config.FileSetValue(remoteName, "type", "protondrive")
		config.FileSetValue(remoteName, "username", email)
		config.FileSetValue(remoteName, "password", obscuredPass)
		if twofa != "" {
			config.FileSetValue(remoteName, "2fa", twofa)
		}
		config.SaveConfig()

	case "webdav":
		url, _ := prompt("WebDAV URL (pl. https://nextcloud.example.com/remote.php/dav/files/user/): ")
		user, _ := prompt("Felhasználónév: ")
		pass, _ := promptPassword("Jelszó: ")

		obscuredPass, err := obscure.Obscure(pass)
		if err != nil {
			return fmt.Errorf("jelszó titkosítása sikertelen: %w", err)
		}

		config.FileSetValue(remoteName, "type", "webdav")
		config.FileSetValue(remoteName, "url", url)
		config.FileSetValue(remoteName, "user", user)
		config.FileSetValue(remoteName, "pass", obscuredPass)
		config.SaveConfig()

	case "sftp":
		host, _ := prompt("SFTP szerver (pl. 192.168.1.100): ")
		port, _ := prompt("Port (alapértelmezett: 22): ")
		user, _ := prompt("Felhasználónév: ")
		pass, _ := promptPassword("Jelszó: ")

		if port == "" {
			port = "22"
		}

		obscuredPass, err := obscure.Obscure(pass)
		if err != nil {
			return fmt.Errorf("jelszó titkosítása sikertelen: %w", err)
		}

		config.FileSetValue(remoteName, "type", "sftp")
		config.FileSetValue(remoteName, "host", host)
		config.FileSetValue(remoteName, "port", port)
		config.FileSetValue(remoteName, "user", user)
		config.FileSetValue(remoteName, "pass", obscuredPass)
		config.SaveConfig()

	case "smb":
		host, _ := prompt("SMB szerver (pl. 192.168.1.100): ")
		share, _ := prompt("Share neve (pl. backup): ")
		user, _ := prompt("Felhasználónév: ")
		pass, _ := promptPassword("Jelszó: ")

		obscuredPass, err := obscure.Obscure(pass)
		if err != nil {
			return fmt.Errorf("jelszó titkosítása sikertelen: %w", err)
		}

		config.FileSetValue(remoteName, "type", "smb")
		config.FileSetValue(remoteName, "host", host)
		config.FileSetValue(remoteName, "share", share)
		config.FileSetValue(remoteName, "user", user)
		config.FileSetValue(remoteName, "pass", obscuredPass)
		config.SaveConfig()

	case "local":
		path, _ := prompt("Mappa útvonala (pl. D:\\Backup\\Minecraft): ")
		config.FileSetValue(remoteName, "type", "local")
		config.FileSetValue(remoteName, "root", path)
		config.SaveConfig()
	}

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

// promptPassword jelszót olvas be
func promptPassword(label string) (string, error) {
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
