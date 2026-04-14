# git-daily

Extract your git activity across repos for a specific day. Outputs clean Markdown, designed as a helper for daily notes and standups.

```
$ git-daily 2026-04-14 ~/projects

## Git Activity — 2026-04-14

> 3 commits across repos · +89/-12 lines · 5 files changed

### my-app (2 commits)
- **14:32** `a1b2c3d` add user profile endpoint (+65/-8) · `feature/profiles`
- **10:15** `e4f5a6b` fix auth middleware token refresh (+12/-4) · `main`

### dotfiles (1 commits)
- **09:03** `c7d8e9f` update neovim config (+12/-0) · `main`
```

## Install

**Homebrew:**

```bash
brew tap peaster/tap
brew install git-daily
```

**Go:**

```bash
go install github.com/peaster/git-daily@latest
```

**Binary:** download from [GitHub Releases](https://github.com/peaster/git-daily/releases).

## Usage

```
git-daily [OPTIONS] [DATE] [REPO_DIR...]
```

| Argument | Description |
|---|---|
| `DATE` | `today` (default), `yesterday`, or `YYYY-MM-DD` |
| `REPO_DIR` | Directories to search for git repos (default: `.`). Searched recursively up to 3 levels deep. |

**Options:**

- `-h`, `--help` — show help
- `-v`, `--version` — print version

**Examples:**

```bash
git-daily                              # today, repos under cwd
git-daily yesterday                    # yesterday
git-daily 2026-04-14 ~/projects        # specific day, specific directory
git-daily today ~/work ~/oss           # multiple search roots
```

## Configuration

Set `GIT_DAILY_AUTHORS` to match commits across multiple identities:

```bash
export GIT_DAILY_AUTHORS="peaster,paul@work.com,paul@home.com"
```

If unset, falls back to `git config user.name` and `user.email`.

## Behavior

- Searches all branches and tags (`--all`)
- Excludes merge commits
- Reports source branch or tag for each commit
- Reports per-commit insertions/deletions and files changed
- Exits 0 on success (including no activity found), 1 on errors

## Building from source

```bash
make build        # build for current platform
make dist         # cross-compile for all platforms
make install      # install to $GOPATH/bin
```

## License

MIT
