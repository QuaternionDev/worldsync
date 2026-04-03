package storage

import "io"

// Provider egy cloud/lokális tárhely providert reprezentál
type Provider interface {
	// Name visszaadja a provider nevét
	Name() string

	// Upload feltölt egy fájlt
	Upload(localPath, remotePath string) error

	// Download letölt egy fájlt
	Download(remotePath, localPath string) error

	// Delete töröl egy fájlt
	Delete(remotePath string) error

	// List listázza a remote mappa tartalmát
	List(remotePath string) ([]FileInfo, error)

	// Exists ellenőrzi, hogy egy fájl létezik-e
	Exists(remotePath string) (bool, error)

	// Reader visszaad egy io.ReadCloser-t egy remote fájlhoz
	Reader(remotePath string) (io.ReadCloser, error)
}

// FileInfo egy remote fájl metaadatait tartalmazza
type FileInfo struct {
	Path    string
	Size    int64
	ModTime string
	IsDir   bool
}
