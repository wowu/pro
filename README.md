# `pro` - Pull Request Opener <!-- omit in toc -->

[![](https://img.shields.io/github/v/release/wowu/pro?label=version)](https://github.com/wowu/pro/releases/latest)
[![](https://img.shields.io/badge/platform-windows%20%7C%20macos%20%7C%20linux-lightgrey)](#installation)

A single command to open current PR in browser. Supports GitHub and GitLab. Available for macOS, Linux and Windows.

![pro](pro.png)

- [Demo](#demo)
- [Usage](#usage)
  - [Authorize GitHub / GitLab](#authorize-github--gitlab)
  - [Open  Pull Request in default browser](#open--pull-request-in-default-browser)
- [Installation](#installation)
  - [Homebrew (macOS/Linux)](#homebrew-macoslinux)
  - [Scoop (Windows)](#scoop-windows)
  - [dpkg (Ubuntu/Debian)](#dpkg-ubuntudebian)
  - [Download latest binary (Linux)](#download-latest-binary-linux)
  - [Go package](#go-package)
  - [Compile from source](#compile-from-source)
  - [Precompiled binaries](#precompiled-binaries)

## Demo

![pro](pro.gif)

## Usage

### Authorize GitHub / GitLab

`pro` uses GitHub or GitLab API to find Pull Request related to current branch. Access is granted via personal access tokens.

Use `auth` command to login:

```bash
# GitHub
pro auth github

# GitLab
pro auth gitlab
```

You will be asked to generate access token and paste it. Tokens are stored in `~/.config/pro/config.yml` by default.

### Open  Pull Request in default browser

To open current Pull Request simply type:

```bash
pro
```

If you're on the main branch (`main`, `master`, `trunk`, etc.) repository homepage will be opened instead. If no PR matching current branch is found, a URL to create new Pull Request will be printed.

Use `-p | --print` flag to print the Pull Request URL instead of opening it in default browser:

```bash
pro -p
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew install wowu/tap/pro
```

### Scoop (Windows)

```bash
scoop bucket add wowu https://github.com/wowu/scoop.git
scoop install wowu/pro
```

### dpkg (Ubuntu/Debian)

```bash
wget https://github.com/wowu/pro/releases/latest/download/pro-linux-amd64.deb
dpkg -i pro-linux-amd64.deb
```

Replace `amd64` with `arm64` if you are on ARM.

### Download latest binary (Linux)

```bash
curl -sSL -o /usr/local/bin/pro https://github.com/wowu/pro/releases/latest/download/pro-linux-amd64
chmod +x /usr/local/bin/pro
```

Replace `amd64` with `arm64` if you are on ARM.

### Go package

```bash
go install github.com/wowu/pro@latest
```

Go 1.18 is required. `pro` binary will be installed in `$GOPATH/bin` (most likely `~/go/bin/pro`).

### Compile from source

1. Install Go 1.18 (`brew install go` or [see offical docs](https://go.dev/doc/install))
2. Clone the repository and build the project:

    ```bash
    git clone git@github.com:Wowu/pro.git && cd pro
    go build
    ```

### Precompiled binaries

Download binaries from the [releases page](https://github.com/wowu/pro/releases/latest).
