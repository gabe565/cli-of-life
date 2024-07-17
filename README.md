# CLI Of Life
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/gabe565/cli-of-life)](https://github.com/gabe565/cli-of-life/releases)
[![Build](https://github.com/gabe565/cli-of-life/actions/workflows/build.yaml/badge.svg)](https://github.com/gabe565/cli-of-life/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gabe565/cli-of-life)](https://goreportcard.com/report/github.com/gabe565/cli-of-life)

Play Conway's Game of Life in your terminal!

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/dfda81d9-2211-4ee9-ae10-201716e5a218">
    <img width="720" alt="cli-of-life demo" src="https://github.com/user-attachments/assets/c5fdf882-af73-4f3e-90cb-e53fd2dcbf35">
  </picture>
</p>

## Installation

### Docker

<details>
  <summary>Click to expand</summary>

A Docker image is available at [ghcr.io/gabe565/cli-of-life](https://ghcr.io/gabe565/cli-of-life)

```shell
sudo docker run --rm -it ghcr.io/gabe565/cli-of-life
```
</details>

### Homebrew (macOS, Linux)

<details>
  <summary>Click to expand</summary>

Install cli-of-life from [gabe565/homebrew-tap](https://github.com/gabe565/homebrew-tap):
```shell
brew install gabe565/tap/cli-of-life
```
</details>

### APT (Ubuntu, Debian)

<details>
  <summary>Click to expand</summary>

1. If you don't have it already, install the `ca-certificates` package
   ```shell
   sudo apt install ca-certificates
   ```

2. Add gabe565 apt repository
   ```
   echo 'deb [trusted=yes] https://apt.gabe565.com /' | sudo tee /etc/apt/sources.list.d/gabe565.list
   ```

3. Update apt repositories
   ```shell
   sudo apt update
   ```

4. Install cli-of-life
   ```shell
   sudo apt install cli-of-life
   ```
</details>

### RPM (CentOS, RHEL)

<details>
  <summary>Click to expand</summary>

1. If you don't have it already, install the `ca-certificates` package
   ```shell
   sudo dnf install ca-certificates
   ```

2. Add gabe565 rpm repository to `/etc/yum.repos.d/gabe565.repo`
   ```ini
   [gabe565]
   name=gabe565
   baseurl=https://rpm.gabe565.com
   enabled=1
   gpgcheck=0
   ```

3. Install cli-of-life
   ```shell
   sudo dnf install cli-of-life
   ```
</details>

### AUR (Arch Linux)

<details>
  <summary>Click to expand</summary>

Install [cli-of-life-bin](https://aur.archlinux.org/packages/cli-of-life-bin) with your [AUR helper](https://wiki.archlinux.org/index.php/AUR_helpers) of choice.
</details>

### Manual Installation

<details>
  <summary>Click to expand</summary>

Download and run the [latest release binary](https://github.com/gabe565/cli-of-life/releases/latest) for your system and architecture.
</details>

## Usage
Run `cli-of-life` in a terminal to play.

By default, the grid will be empty, but rle/plaintext files can be loaded with `cli-of-life --file FILE.rle`

See [usage docs](docs/cli-of-life.md) for cli flag documentation.

### Keybinds

| Key     | Description                               |
|---------|-------------------------------------------|
| mouse   | Place cells                               |
| `space` | Play/pause                                |
| `m`     | Toggle between modes: smart, place, erase |
| `↑↓←→`  | Move the game board                       |
| `w`     | Toggle wrapping                           |
| `<`/`>` | Change playback speed                     |
| `t`     | Tick                                      |
| `r`     | Reset                                     |
| `q`     | Quit                                      |
