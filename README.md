# Clipsyboogie - cli clipboard logging

`clipsyboogie` is a simple cli tool written in Go to capture and retrieve the current system clipboard contents.

![macOS cli](https://img.shields.io/badge/macOS-cli-blue?logo=apple)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/balintb/clipsyboogie)
![GitHub License](https://img.shields.io/github/license/balintb/clipsyboogie)
![GitHub branch check runs](https://img.shields.io/github/check-runs/balintb/clipsyboogie/main)

> [!CAUTION]
> `clipsyboogie` has been tested on macOS only. `install` and `uninstall` are macOS-specific.

## Installation

### Homebrew

The easiest way to install `clipsyboogie` is via [Homebrew](https://brew.sh):

```console
brew install balintb/tap/clipsyboogie
```

If this fails, you might need to add the tap first:

```console
brew tap balintb/tap
```

And retry `brew install balintb/tap/clipsyboogie`.

### Go

To install via Go,

```console
go install github.com/balintb/clipsyboogie@latest
```

Make sure you have Go environment variables set up, especially `GOPATH`, and your `PATH` with `$GOPATH/bin`.

### LaunchAgent

To install a LaunchAgent (run in the background), use

```console
clipsyboogie install
```

## Configuration

Current configuration options are limited to command line flags and environment variables, notably

- `--interval` or `$CBG_INTERVAL`: polling interval in ms. Any clipboard contents that live for a shorter period of time are not guaranteed to be captured. Minimum value has been set at `10` for performance reasons.
- `--run-at-load` to enable running LaunchAgent on load. Default `true`.

Pull requests are welcome for a proper file-based config. 

## Usage

`clipsyboogie [global options] command [command options]`

### Commands

```console
add, a        Record clipboard content
get, g        Get latest N entries
install, i    Install LaunchAgent
uninstall, u  Uninstall LaunchAgent
listen, l     Listen (poll) for clipboard changes
help, h       Shows a list of commands or help for one command
```

To start in the background, `clipsyboogie install` creates a plist in `~/Library/LaunchAgents/com.balintb.clipsyboogie.plist`, which can be loaded with `launchctl load ~/Library/LaunchAgents/com.balintb.clipsyboogie.plist`.

## Retrieving stored clipboard content

`clipsyboogie` stores clipboard content in `~/.clipsyboogie/clips.db` as an SQLite database. You can open it with `sqlite` or use `clipsyboogie get [N]` to retrieve the last `N` (default `1`) entries.

## Roadmap

- [ ] Add `load` / `unload` commands to load/unload via `launchctl`
- [ ] File-based configuration
- [ ] Configurable retention period or number of items to store
- [x] Homebrew tap

## License

[MIT](LICENSE)

Clipsyboogie Copyright [@balintb](https://balint.click/github)