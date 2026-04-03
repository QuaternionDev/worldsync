package world

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Tnze/go-mc/nbt"
)

// GameMode a játékmódot reprezentálja
type GameMode int

const (
	Survival  GameMode = 0
	Creative  GameMode = 1
	Adventure GameMode = 2
	Spectator GameMode = 3
)

func (g GameMode) String() string {
	switch g {
	case Survival:
		return "Survival"
	case Creative:
		return "Creative"
	case Adventure:
		return "Adventure"
	case Spectator:
		return "Spectator"
	default:
		return "Unknown"
	}
}

// Difficulty a nehézségi szintet reprezentálja
type Difficulty int

const (
	Peaceful Difficulty = 0
	Easy     Difficulty = 1
	Normal   Difficulty = 2
	Hard     Difficulty = 3
)

func (d Difficulty) String() string {
	switch d {
	case Peaceful:
		return "Peaceful"
	case Easy:
		return "Easy"
	case Normal:
		return "Normal"
	case Hard:
		return "Hard"
	default:
		return "Unknown"
	}
}

// World egy Minecraft világot reprezentál
type World struct {
	Name       string
	Path       string
	Version    string
	GameMode   GameMode
	Difficulty Difficulty
	Seed       int64
	LastPlayed time.Time
	SizeBytes  int64
}

// levelDat a level.dat NBT struktúrája
// levelDat a level.dat NBT struktúrája – minden verzióhoz
type levelDat struct {
	Data struct {
		LevelName        string `nbt:"LevelName"`
		LastPlayed       int64  `nbt:"LastPlayed"`
		GameType         int32  `nbt:"GameType"`
		Difficulty       int8   `nbt:"Difficulty"`
		RandomSeed       int64  `nbt:"RandomSeed"`
		VersionInt       int32  `nbt:"version"` // régi formátum (beta/alpha)
		WorldGenSettings struct {
			Seed int64 `nbt:"seed"`
		} `nbt:"WorldGenSettings"`
	} `nbt:"Data"`
}

// levelDatNewVersion újabb világokhoz ahol Version struktúra van (1.9+)
type levelDatNewVersion struct {
	Data struct {
		LevelName        string `nbt:"LevelName"`
		LastPlayed       int64  `nbt:"LastPlayed"`
		GameType         int32  `nbt:"GameType"`
		Difficulty       int8   `nbt:"Difficulty"`
		RandomSeed       int64  `nbt:"RandomSeed"`
		VersionInt       int32  `nbt:"version"`
		WorldGenSettings struct {
			Seed int64 `nbt:"seed"`
		} `nbt:"WorldGenSettings"`
		Version struct {
			Name string `nbt:"Name"`
		} `nbt:"Version"`
	} `nbt:"Data"`
}

// ScanWorlds megkeresi és beolvassa az összes világot egy saves mappában
func ScanWorlds(savesPath string) ([]World, error) {
	entries, err := os.ReadDir(savesPath)
	if err != nil {
		return nil, err
	}

	var worlds []World

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		worldPath := filepath.Join(savesPath, entry.Name())

		// Kihagyjuk ha nincs level.dat (nem valódi világ, pl. NEI mappa)
		levelDatPath := filepath.Join(worldPath, "level.dat")
		if _, err := os.Stat(levelDatPath); os.IsNotExist(err) {
			continue
		}

		w, err := readWorld(worldPath)
		if err != nil {
			fmt.Printf("  [debug] %s: %s\n", entry.Name(), err)
			continue
		}

		worlds = append(worlds, *w)
	}

	return worlds, nil
}

// readWorld beolvassa egy világ adatait a level.dat fájlból
func readWorld(worldPath string) (*World, error) {
	levelDatPath := filepath.Join(worldPath, "level.dat")

	// Először próbáljuk az újabb formátumot (Version struktúrával)
	version, data, err := parseLevelDat(levelDatPath)
	if err != nil {
		return nil, err
	}

	seed := data.Data.WorldGenSettings.Seed
	if seed == 0 {
		seed = data.Data.RandomSeed
	}

	size, err := dirSize(worldPath)
	if err != nil {
		size = 0
	}

	return &World{
		Name:       data.Data.LevelName,
		Path:       worldPath,
		Version:    version,
		GameMode:   GameMode(data.Data.GameType),
		Difficulty: Difficulty(data.Data.Difficulty),
		Seed:       seed,
		LastPlayed: time.UnixMilli(data.Data.LastPlayed),
		SizeBytes:  size,
	}, nil
}

func parseLevelDat(path string) (string, *levelDat, error) {
	// Először próbáljuk az újabb formátumot (1.9+ ahol Version struktúra van)
	withVersion, err := decodeLevelDat[levelDatNewVersion](path)
	if err == nil && withVersion.Data.Version.Name != "" {
		base := &levelDat{}
		base.Data.LevelName = withVersion.Data.LevelName
		base.Data.LastPlayed = withVersion.Data.LastPlayed
		base.Data.GameType = withVersion.Data.GameType
		base.Data.Difficulty = withVersion.Data.Difficulty
		base.Data.RandomSeed = withVersion.Data.RandomSeed
		base.Data.WorldGenSettings.Seed = withVersion.Data.WorldGenSettings.Seed
		return withVersion.Data.Version.Name, base, nil
	}

	// Régi formátum (alpha/beta/modded) – Version mező int vagy nem létezik
	plain, err := decodeLevelDat[levelDat](path)
	if err != nil {
		return "", nil, err
	}

	// VersionInt alapján megpróbáljuk meghatározni a verziót
	version := legacyVersionName(plain.Data.VersionInt)

	return version, plain, nil
}

// legacyVersionName a régi integer verziószámot olvasható névvé alakítja
func legacyVersionName(v int32) string {
	switch v {
	case 0:
		return "Alpha"
	case 19132, 19133:
		return "Beta 1.7-"
	case 19134:
		return "Beta 1.8+"
	default:
		if v > 0 {
			return fmt.Sprintf("Legacy (v%d)", v)
		}
		return "Legacy (Alpha/Beta)"
	}
}

// DeduplicateWorlds eltávolítja a duplikált világokat seed és méret alapján
func DeduplicateWorlds(worlds []World) []World {
	seen := make(map[string]bool)
	var result []World

	for _, w := range worlds {
		// Kulcs: név + seed kombináció
		key := fmt.Sprintf("%s_%d", w.Name, w.Seed)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, w)
	}

	return result
}

func decodeLevelDat[T any](path string) (*T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	var data T
	_, err = nbt.NewDecoder(gr).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// dirSize rekurzívan kiszámítja egy mappa méretét
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// FormatSize olvasható formátumra hozza a méretet
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
