# CLI Of Life
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/gabe565/cli-of-life)](https://github.com/gabe565/cli-of-life/releases)
[![Build](https://github.com/gabe565/cli-of-life/actions/workflows/build.yaml/badge.svg)](https://github.com/gabe565/cli-of-life/actions/workflows/build.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gabe565/cli-of-life)](https://goreportcard.com/report/github.com/gabe565/cli-of-life)

Run Conway's Game of Life in your terminal!

## Installation

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

### Homebrew (macOS, Linux)

<details>
  <summary>Click to expand</summary>

Install cli-of-life from [gabe565/homebrew-tap](https://github.com/gabe565/homebrew-tap):
```shell
brew install gabe565/tap/cli-of-life
```
</details>

### Manual Installation

<details>
  <summary>Click to expand</summary>

Download and run the [latest release binary](https://github.com/gabe565/cli-of-life/releases/latest) for your system and architecture.
</details>
