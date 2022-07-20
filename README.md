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
