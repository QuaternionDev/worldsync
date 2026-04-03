package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ProviderType a támogatott cloud providereket reprezentálja
type ProviderType string

const (
	ProviderOneDrive    ProviderType = "onedrive"
	ProviderGoogleDrive ProviderType = "drive"
	ProviderProtonDrive ProviderType = "protondrive"
	ProviderWebDAV      ProviderType = "webdav"
	ProviderSFTP        ProviderType = "sftp"
	ProviderSMB         ProviderType = "smb"
	ProviderLocal       ProviderType = "local"
)

// ProviderConfig egy konfigurált cloud providert reprezentál
type ProviderConfig struct {
	Name       string            `json:"name"`        // felhasználó által adott név
	Type       ProviderType      `json:"type"`        // provider típusa
	RcloneName string            `json:"rclone_name"` // rclone remote neve
	Settings   map[string]string `json:"settings"`    // provider-specifikus beállítások
}

// Config a WorldSync fő konfigurációja
type Config struct {
	Providers      []ProviderConfig `json:"providers"`
	ActiveProvider string           `json:"active_provider"` // aktív provider neve
	SyncOnExit     bool             `json:"sync_on_exit"`    // szinkronizálás kilépéskor
	KeepSnapshots  int              `json:"keep_snapshots"`  // hány snapshot-ot őrzünk meg
}

// DefaultConfig visszaadja az alapértelmezett konfigurációt
func DefaultConfig() *Config {
	return &Config{
		Providers:     []ProviderConfig{},
		SyncOnExit:    true,
		KeepSnapshots: 5,
	}
}

// ConfigDir visszaadja a WorldSync konfigurációs mappáját
func ConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "WorldSync")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "WorldSync")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "worldsync")
	}
}

// ConfigPath visszaadja a config fájl teljes útvonalát
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// Load betölti a konfigurációt
func Load() (*Config, error) {
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Ha nem létezik, visszaadjuk az alapértelmezettet
		return DefaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save elmenti a konfigurációt
func Save(cfg *Config) error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// AddProvider hozzáad egy új providert
func (c *Config) AddProvider(p ProviderConfig) error {
	for _, existing := range c.Providers {
		if existing.Name == p.Name {
			return fmt.Errorf("már létezik '%s' nevű provider", p.Name)
		}
	}
	c.Providers = append(c.Providers, p)
	return nil
}

// GetActiveProvider visszaadja az aktív providert
func (c *Config) GetActiveProvider() *ProviderConfig {
	for i, p := range c.Providers {
		if p.Name == c.ActiveProvider {
			return &c.Providers[i]
		}
	}
	if len(c.Providers) > 0 {
		return &c.Providers[0]
	}
	return nil
}
