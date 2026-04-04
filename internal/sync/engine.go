package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileState egy fájl állapotát tárolja
type FileState struct {
	Hash     string    `json:"hash"`
	Modified time.Time `json:"modified"`
	Size     int64     `json:"size"`
}

// WorldState egy világ összes fájljának állapotát tárolja
type WorldState struct {
	WorldName string               `json:"world_name"`
	LastSync  time.Time            `json:"last_sync"`
	Files     map[string]FileState `json:"files"`
}

// SyncResult a szinkronizálás eredménye
type SyncResult struct {
	Uploaded []string
	Skipped  []string
	Deleted  []string
	Errors   []string
}

// Engine a szinkronizálást végzi
type Engine struct {
	StateDir string // ahol a .worldsync state fájlok tárolódnak
}

// NewEngine létrehoz egy új Engine-t
func NewEngine(stateDir string) (*Engine, error) {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, err
	}
	return &Engine{StateDir: stateDir}, nil
}

// SyncToLocal szinkronizál egy világot egy lokális célmappába
func (e *Engine) SyncToLocal(worldPath, destBase string) (*SyncResult, error) {
	worldName := filepath.Base(worldPath)
	destPath := filepath.Join(destBase, worldName)
	result := &SyncResult{}

	// Célmappa létrehozása ha nem létezik
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return nil, fmt.Errorf("célmappa létrehozása sikertelen: %w", err)
	}

	// Előző state betöltése
	oldState, _ := e.loadState(worldName)

	// Jelenlegi fájlok bejárása
	newState := &WorldState{
		WorldName: worldName,
		LastSync:  time.Now(),
		Files:     make(map[string]FileState),
	}

	err := filepath.Walk(worldPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Relatív útvonal a world mappához képest
		relPath, err := filepath.Rel(worldPath, path)
		if err != nil {
			return err
		}

		// Hash számítása
		hash, err := hashFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", relPath, err))
			return nil
		}

		newState.Files[relPath] = FileState{
			Hash:     hash,
			Modified: info.ModTime(),
			Size:     info.Size(),
		}

		// Kell-e feltölteni?
		if oldState != nil {
			if old, exists := oldState.Files[relPath]; exists && old.Hash == hash {
				result.Skipped = append(result.Skipped, relPath)
				return nil
			}
		}

		// Fájl másolása a célba
		destFile := filepath.Join(destPath, relPath)
		if err := copyFile(path, destFile); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", relPath, err))
			return nil
		}

		result.Uploaded = append(result.Uploaded, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// State mentése
	if err := e.saveState(worldName, newState); err != nil {
		return nil, fmt.Errorf("state mentése sikertelen: %w", err)
	}

	fmt.Printf("  ✓ %s: %d feltöltve, %d kihagyva, %d hiba\n",
		worldName,
		len(result.Uploaded),
		len(result.Skipped),
		len(result.Errors),
	)

	return result, nil
}

// PullFromRemote letölti a remote világokat a lokális saves mappába
func (e *Engine) PullFromLocal(srcBase, savesPath string) error {
	// Megkeresi a srcBase-ben lévő összes világot
	entries, err := os.ReadDir(srcBase)
	if err != nil {
		return fmt.Errorf("backup mappa olvasása sikertelen: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		worldName := entry.Name()
		srcWorldPath := filepath.Join(srcBase, worldName)
		destWorldPath := filepath.Join(savesPath, worldName)

		// Ellenőrizzük hogy lokálisan létezik-e már
		if dirExists(destWorldPath) {
			fmt.Printf("  ↔ %s már létezik lokálisan, kihagyva\n", worldName)
			continue
		}

		// Letöltés (másolás)
		fmt.Printf("  ↓ %s letöltése...\n", worldName)
		if err := copyDir(srcWorldPath, destWorldPath); err != nil {
			fmt.Printf("  ✗ %s: %s\n", worldName, err)
			continue
		}

		fmt.Printf("  ✓ %s letöltve\n", worldName)
	}

	return nil
}

// copyDir rekurzívan másol egy mappát
func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return copyFile(path, destPath)
	})
}

// dirExists ellenőrzi hogy egy mappa létezik-e
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// --- State kezelés ---

func (e *Engine) loadState(worldName string) (*WorldState, error) {
	path := filepath.Join(e.StateDir, worldName+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state WorldState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (e *Engine) saveState(worldName string, state *WorldState) error {
	path := filepath.Join(e.StateDir, worldName+".json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// --- Segédfüggvények ---

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
