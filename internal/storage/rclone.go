package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "github.com/rclone/rclone/backend/drive"
	_ "github.com/rclone/rclone/backend/onedrive"
	_ "github.com/rclone/rclone/backend/protondrive"
	_ "github.com/rclone/rclone/backend/sftp"
	_ "github.com/rclone/rclone/backend/smb"
	_ "github.com/rclone/rclone/backend/webdav"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/sync"
)

// RcloneProvider egy rclone-alapú storage providert reprezentál
type RcloneProvider struct {
	name       string
	remoteName string // pl. "onedrive", "gdrive"
	remotePath string // pl. "WorldSync/worlds"
}

// NewRcloneProvider létrehoz egy új rclone providert
func NewRcloneProvider(name, remoteName, remotePath string) (*RcloneProvider, error) {
	// rclone config inicializálása
	configfile.Install()

	return &RcloneProvider{
		name:       name,
		remoteName: remoteName,
		remotePath: remotePath,
	}, nil
}

func (r *RcloneProvider) Name() string {
	return r.name
}

// remoteFsPath visszaadja a teljes rclone remote útvonalat
func (r *RcloneProvider) remoteFsPath(subPath string) string {
	if subPath == "" {
		return fmt.Sprintf("%s:%s", r.remoteName, r.remotePath)
	}
	return fmt.Sprintf("%s:%s/%s", r.remoteName, r.remotePath, subPath)
}

func (r *RcloneProvider) Upload(localPath, remotePath string) error {
	ctx := context.Background()

	// Forrás fájl megnyitása
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("forrás megnyitása sikertelen: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Remote fs létrehozása
	remoteDir := filepath.Dir(r.remoteFsPath(remotePath))
	remoteFs, err := fs.NewFs(ctx, remoteDir)
	if err != nil {
		return fmt.Errorf("remote fs hiba: %w", err)
	}

	// Feltöltés
	_, err = operations.RcatSize(ctx, remoteFs, filepath.Base(remotePath), srcFile, srcInfo.Size(), srcInfo.ModTime(), nil)
	return err
}

func (r *RcloneProvider) Download(remotePath, localPath string) error {
	ctx := context.Background()

	remoteFs, err := fs.NewFs(ctx, r.remoteFsPath(filepath.Dir(remotePath)))
	if err != nil {
		return err
	}

	remoteObj, err := remoteFs.NewObject(ctx, filepath.Base(remotePath))
	if err != nil {
		return err
	}

	// Lokális mappa létrehozása
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	// Letöltés
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	reader, err := remoteObj.Open(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(localFile, reader)
	return err
}

func (r *RcloneProvider) Delete(remotePath string) error {
	ctx := context.Background()

	remoteFs, err := fs.NewFs(ctx, r.remoteFsPath(filepath.Dir(remotePath)))
	if err != nil {
		return err
	}

	obj, err := remoteFs.NewObject(ctx, filepath.Base(remotePath))
	if err != nil {
		return err
	}

	return obj.Remove(ctx)
}

func (r *RcloneProvider) List(remotePath string) ([]FileInfo, error) {
	ctx := context.Background()

	remoteFs, err := fs.NewFs(ctx, r.remoteFsPath(remotePath))
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	err = operations.ListFn(ctx, remoteFs, func(obj fs.Object) {
		files = append(files, FileInfo{
			Path:    obj.Remote(),
			Size:    obj.Size(),
			ModTime: obj.ModTime(ctx).String(),
			IsDir:   false,
		})
	})

	return files, err
}

func (r *RcloneProvider) Exists(remotePath string) (bool, error) {
	ctx := context.Background()

	remoteFs, err := fs.NewFs(ctx, r.remoteFsPath(filepath.Dir(remotePath)))
	if err != nil {
		return false, err
	}

	_, err = remoteFs.NewObject(ctx, filepath.Base(remotePath))
	if err == fs.ErrorObjectNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RcloneProvider) Reader(remotePath string) (io.ReadCloser, error) {
	ctx := context.Background()

	remoteFs, err := fs.NewFs(ctx, r.remoteFsPath(filepath.Dir(remotePath)))
	if err != nil {
		return nil, err
	}

	obj, err := remoteFs.NewObject(ctx, filepath.Base(remotePath))
	if err != nil {
		return nil, err
	}

	return obj.Open(ctx)
}

// SyncWorld szinkronizál egy teljes világ mappát a remote-ra
func (r *RcloneProvider) SyncWorld(localWorldPath, worldName string) error {
	ctx := context.Background()

	// Forrás fs
	srcFs, err := fs.NewFs(ctx, localWorldPath)
	if err != nil {
		return fmt.Errorf("forrás fs hiba: %w", err)
	}

	// Cél fs
	dstFs, err := fs.NewFs(ctx, r.remoteFsPath(worldName))
	if err != nil {
		return fmt.Errorf("cél fs hiba: %w", err)
	}

	// Sync – csak a változott fájlok kerülnek fel
	return sync.Sync(ctx, dstFs, srcFs, false)
}

// ConfigPath visszaadja az rclone config fájl helyét
func ConfigPath() string {
	return config.GetConfigPath()
}
