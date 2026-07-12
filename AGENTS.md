# pro

## Overview

`pro` (Pull Request Opener) is a Go CLI that opens the pull/merge request for the current git branch in the browser. It supports GitHub and GitLab, and runs on macOS, Linux, and Windows.

## Commands

- Build: `go build` (produces the `pro` binary in the repo root)
- Run without building: `go run .`
- Lint: `golangci-lint run` (CI uses golangci-lint; no config file, so defaults apply)
- Format: `gofmt -w .`
- Tests: there is no test suite. `go test ./...` is a no-op today.
- Bump version + tag a release: `bin/bump_version vX.Y.Z` (requires `gsed`; edits the `Version:` string in `pro.go`, commits, and optionally tags). Pushing a `v*` tag triggers the goreleaser release workflow.

## Architecture

The flow is: `pro.go` (CLI wiring) → `command/` (orchestration + user-facing output) → `repository/` (local git introspection) + `provider/` (remote API calls) + `config/` (token storage).

- **`pro.go`** — urfave/cli app definition. Declares the `auth`, `open`, and `list` (alias `ls`) commands plus the `--print/-p` and `--copy/-c` flags. The default action (no subcommand) is `command.Open`. The hardcoded `Version:` string here is what `bin/bump_version` rewrites.

- **`command/`** — the actual behavior, and the only layer that prints to the user or calls `os.Exit`. Lower layers return typed errors; `command/` translates them into colored messages (via `fatih/color`) with remediation hints and picks the exit code.
  - `open.go` — resolves branch → PR/MR URL. Main-branch names (`main`, `master`, `trunk`, `develop`, `dev`) short-circuit to the repo home page. If no PR/MR exists but the branch is pushed to the remote, it builds a "create PR/MR" URL; if the branch isn't on the remote, it errors. Host is dispatched by `gitURL.Host` (`github.com` vs `gitlab.com`) — other hosts are unsupported. Also owns `openBrowser` (per-OS: `open`/`xdg-open`/`rundll32`) and `copyToClipboard`.
  - `list.go` — interactive PR/MR picker using `ktr0731/go-fuzzyfinder`.
  - `auth.go` — prompts for and saves a personal access token.

- **`repository/`** — wraps `go-git` to find the repo and read the origin URL. Two things are done manually rather than through go-git:
  - **Worktree support**: `CurrentBranchName` reads and parses the `.git/HEAD` file directly because go-git does not resolve branches correctly inside external worktrees. `makeRepository` detects the `.git/worktrees` path and tracks both the worktree git dir and the real git dir separately.
  - **Parent traversal**: `FindInParents` recurses up the directory tree until it finds a git repo or hits the filesystem root.
  - Typed errors live in `errors.go` (`ErrNoRemoteOrigin`, `ErrNoActiveBranch`) and are matched with `errors.Is` in `command/`.

- **`provider/github/` and `provider/gitlab/`** — one file each, self-contained REST clients built on `net/http` (no SDK). Each finds the PR/MR for a branch and lists remote branches, handles pagination, and returns typed sentinel errors (e.g. `ErrNotFound`/`ErrUnauthorized`, `ErrMergeRequestNotFound`/`ErrProjectNotFound`/`ErrTokenExpired`). The two providers do not share an interface — `command/` switches on host and calls each directly.

- **`config/`** — reads/writes `~/.config/pro/config.yml` (YAML), holding `github_token` and `gitlab_token`. Written with `0600`. A missing file yields an empty `Config{}` rather than an error.

## Conventions

- Errors flow up as typed sentinel values; only `command/` decides how to present them and when to exit. Keep new provider/repository errors as sentinels matched with `errors.Is`, don't print or exit below the `command/` layer.
- User-facing output goes to **stderr** with `fatih/color`; the resolved URL is the only thing intended for **stdout** (so `pro -p` is pipeable).
