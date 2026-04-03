# worldsync

A cross-platform cloud sync tool for Minecraft Java Edition worlds. 
Supports all versions (Alpha, Beta, Release, Snapshots) and modded instances.

## How to help

If you find any issues/bugs, please create an issue on this repo.
If you want to contribute, create a fork of this project, and make a pull request later with
you new feature.
For any recommendations or further comments, drop me an email on kapcsolat@quaternion.hu.

## Features

- **Automatic launcher detection** – Vanilla, Prism Launcher, CurseForge
- **Multiple cloud providers** – OneDrive, Google Drive, Proton Drive, WebDAV, SFTP, SMB/NAS, Local
- **Delta sync** – only changed files are uploaded
- **Modpack support** – detects and displays modpack names
- **Terminal UI** – beautiful interactive interface
- **CLI** – scriptable command-line interface

## How Sync Works

### Local provider
WorldSync calculates a SHA-256 hash for every file in your world folder and compares it against the last sync state. Only changed files are copied to the backup destination. Sync state is stored in:
- Windows: `%APPDATA%\WorldSync\state\`
- Linux: `~/.config/worldsync/state/`

### Cloud providers (OneDrive, Google Drive, etc.)
WorldSync uses rclone's delta sync under the hood. It compares file sizes and modification times, and only uploads changed files to `WorldSync/worlds/<world_name>/` in your cloud storage.

### Current limitations
- Sync is **manual** – you need to run `worldsync sync` or use TUI after closing Minecraft
- Download/restore from cloud is not yet implemented
- Conflict detection is not yet implemented – if you played on two machines offline, the latest sync wins
- Auto-sync on game exit is planned for a future release

## Installation

### Windows

1. Download the latest `worldsync.exe` from the [Releases](https://github.com/QuaternionDev/worldsync/releases) page
2. Place it somewhere on your `PATH` (e.g. `C:\Program Files\WorldSync\`)
3. Open a terminal and run:
```powershell
worldsync provider add
```

### Linux

1. Download the latest `worldsync` binary from the [Releases](https://github.com/QuaternionDev/worldsync/releases) page
2. Make it executable and move it to your PATH:
```bash
chmod +x worldsync
sudo mv worldsync /usr/local/bin/
```

3. Set up a provider:
```bash
worldsync provider add
```

## Usage

### Terminal UI
```bash
worldsync tui
```

### CLI Commands
```bash
# Show all worlds and sync status
worldsync status

# Sync all worlds to the active provider
worldsync sync

# Sync using a specific provider
worldsync sync --provider "My OneDrive"

# Manage providers
worldsync provider list
worldsync provider add
worldsync provider remove "My OneDrive"
worldsync provider set "My Google Drive"
```

## Supported Cloud Providers

| Provider | Authentication |
|---|---|
| OneDrive | OAuth2 (browser) |
| Google Drive | OAuth2 (browser) |
| Proton Drive | Email + Password + 2FA |
| WebDAV | Username + Password |
| SFTP | Username + Password |
| SMB / NAS | Username + Password |
| Local folder | Path |

## Building from Source

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- Git

### Windows
```powershell
git clone https://github.com/QuaternionDev/worldsync.git
cd worldsync
go build -o worldsync.exe -ldflags="-s -w" .
```

### Linux
```bash
git clone https://github.com/QuaternionDev/worldsync.git
cd worldsync
go build -o worldsync -ldflags="-s -w" .
```

## Supported Launchers

- **Vanilla** (Official Minecraft Launcher)
- **Prism Launcher**
- **CurseForge**

More launchers coming soon.

## Behind the scenes

This project was made because I wanted to make a way to sync my MC worlds on my Dual Boot Arch machine.
I never touched Go before, but with the help of the internet and Claude, I learned some things, and managed
to put together a working tool for myself, and publish it to the internet too in case someone would need something similar.

## Disclaimer

This project was built with the assistance of Claude by Anthropic. The code was generated and iterated through a conversational development process. Most of this repo is vibe-coded. (Documentation is written by me, and TUI design was created by me too)


## License

MIT License – see [LICENSE](LICENSE) for details.