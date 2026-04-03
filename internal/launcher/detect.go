package launcher

import (
	"os"
	"path/filepath"
	"runtime"
)

// Launcher egy felismert Minecraft launchert reprezentál
type Launcher struct {
	Name      string   // pl. "Vanilla", "Prism Launcher"
	SavePaths []string // az összes megtalált saves/ mappa
}

// DetectAll megkeresi az összes telepített launchert
func DetectAll() []Launcher {
	var launchers []Launcher

	detectors := []func() *Launcher{
		detectVanilla,
		detectPrism,
		detectCurseForge,
	}

	for _, detect := range detectors {
		if l := detect(); l != nil {
			launchers = append(launchers, *l)
		}
	}

	return launchers
}

// detectVanilla az Official Launcher saves mappáit keresi
func detectVanilla() *Launcher {
	base := vanillaBasePath()
	savesPath := filepath.Join(base, "saves")

	if !dirExists(savesPath) {
		return nil
	}

	return &Launcher{
		Name:      "Vanilla",
		SavePaths: []string{savesPath},
	}
}

// detectPrism a Prism Launcher instance-eit keresi
func detectPrism() *Launcher {
	candidates := prismBasePaths()
	var saves []string

	for _, base := range candidates {
		instancesDir := filepath.Join(base, "instances")
		if !dirExists(instancesDir) {
			continue
		}

		entries, err := os.ReadDir(instancesDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Prism használhat "minecraft" vagy ".minecraft" mappát is
			savesPath := filepath.Join(instancesDir, entry.Name(), "minecraft", "saves")
			if !dirExists(savesPath) {
				savesPath = filepath.Join(instancesDir, entry.Name(), ".minecraft", "saves")
			}

			if dirExists(savesPath) {
				saves = append(saves, savesPath)
			}
		}
	}

	if len(saves) == 0 {
		return nil
	}

	return &Launcher{Name: "Prism Launcher", SavePaths: saves}
}

// detectCurseForge a CurseForge instance-eit keresi
func detectCurseForge() *Launcher {
	candidates := curseForgeBasePaths()
	var saves []string

	for _, base := range candidates {
		instancesDir := filepath.Join(base, "minecraft", "Instances")
		if !dirExists(instancesDir) {
			continue
		}

		entries, err := os.ReadDir(instancesDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			savesPath := filepath.Join(instancesDir, entry.Name(), "saves")
			if dirExists(savesPath) {
				saves = append(saves, savesPath)
			}
		}
	}

	if len(saves) == 0 {
		return nil
	}

	return &Launcher{Name: "CurseForge", SavePaths: saves}
}

// --- Platform-specifikus útvonalak ---

func vanillaBasePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), ".minecraft")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "minecraft")
	default: // linux
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".minecraft")
	}
}

func prismBasePaths() []string {
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		localappdata := os.Getenv("LOCALAPPDATA")
		return []string{
			filepath.Join(appdata, "PrismLauncher"),
			filepath.Join(localappdata, "PrismLauncher"),
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		return []string{
			filepath.Join(home, "Library", "Application Support", "PrismLauncher"),
		}
	default:
		home, _ := os.UserHomeDir()
		return []string{
			filepath.Join(home, ".local", "share", "PrismLauncher"),
		}
	}
}

func curseForgeBasePaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join(os.Getenv("USERPROFILE"), "curseforge"),
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		return []string{
			filepath.Join(home, "curseforge"),
		}
	default:
		home, _ := os.UserHomeDir()
		return []string{
			filepath.Join(home, "curseforge"),
		}
	}
}

// dirExists ellenőrzi, hogy egy mappa létezik-e
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
