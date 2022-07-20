# `pro` - Pull Request Opener

A single command to open current PR in browser. Supports GitHub and GitLab.

## Usage

Open Pull Request for current branch in browser:

```bash
pro
```

Print Pull Request URL instead of opening it:

```bash
pro -p
```

Authorize `pro` to access your GitHub account:

```bash
pro auth github
```

Authorize `pro` to access your GitLab account:

```bash
pro auth gitlab
```

Tokens are stored in `~/.config/pro/config.yml` by default.

## Installation

### Homebrew

```bash
brew install wowu/tap/pro
```

### Compile from source

Install go 1.18 (`brew install go` or [offical docs](https://go.dev/doc/install)), then compile the binary from source with:

```bash
go install github.com/wowu/pro@latest
```

`pro` binary will be installed in `$GOPATH/bin`, most likely `~/go/bin/pro`.

### Other platforms

Download binaries from the [releases page](https://github.com/wowu/pro/releases/latest).
