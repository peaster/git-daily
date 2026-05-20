# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build              # build ./git-daily for the current platform with version baked in
make install            # go install to $GOPATH/bin
make dist               # cross-compile darwin/linux/windows × amd64/arm64 into dist/

go test ./...           # run all tests
go test -run TestName   # run a single test (e.g. -run TestExtractCommitsAndMerges)
go vet ./...
```

The version string is injected at link time via `-X main.version=$(git describe --tags --always --dirty)`. A plain `go build` produces a binary that reports `version = "dev"`.

## Architecture

Single `main` package, flat layout, no internal subpackages. Each file owns one responsibility — touch the file whose name matches the concern.

**Data flow** (`main.go`):
1. `parseArgs` → date, repo dirs, render flags
2. `resolveAuthors` → identities from `GIT_DAILY_AUTHORS` env, falling back to `git config user.{email,name}`
3. `findRepos` → walks each search dir up to **depth 3**, dedupes, returns sorted absolute paths
4. `extractActivity` → per-repo `RepoActivity` with `Events []Event`, sorted newest-first
5. `formatMarkdown` → assembles the document
6. `renderForOutput` → either raw markdown (piped/redirected/`--plain`/`NO_COLOR`) or Glamour ANSI

**Event model** (`activity.go`): `EventKind` is an iota enum (`KindCommit | KindMerge | KindBranch | KindTag`) with `numKinds` as the sentinel terminator used to size fixed-length arrays (`Counts [numKinds]int`, `kindLabels [numKinds]struct{Singular,Plural string}`). Adding a kind = add iota entry, bump `numKinds`, add a `kindLabels` row, and add a case in `format.go:renderEvent`. There is no per-kind type switch outside `renderEvent`.

**Event extraction**: the tool shells out to `git` — there is no Go git library dependency. `runGit` (`git.go`) swallows all errors and returns `""`, by design (matches the bash `|| true` pattern). Each extractor uses a different git mechanism:
- Commits/merges: `git log --all --shortstat --after/--before --author=...`. Merge classification is based on multiple parents in `%P`.
- Branches: `git for-each-ref refs/heads`, then `git log -g --grep-reflog="^branch: Created from "` per branch. Pruned reflogs silently miss older creations.
- Tags: `git for-each-ref refs/tags` with a unit-separator (`\x1f`) format string. Annotated vs lightweight is distinguished via `%(objecttype)`; the email field differs (tagger vs author).

**Author matching** (`authorMatches`): case-insensitive **substring** match against any identity in the configured list. A user's name is a valid identity, not just email — `GIT_DAILY_AUTHORS="peaster"` matches both `peaster@example.com` and `Paul Easterbrooks <peaster@...>` because the candidate string is `git`'s `%ae`/`%ce`/`%(taggeremail)` value.

**Remote URL normalization** (`git.go`): `normalizeRemoteURL` accepts https/http/ssh/git/SCP-style URLs and returns a base usable as `<base>/commit/<sha>`. The `/commit/<sha>` suffix is honored by GitHub, GitLab, Gitea/Forgejo, Codeberg, Azure DevOps, and sourcehut — do not special-case providers.

**Rendering** (`render.go`): `shouldOutputPlain` short-circuits on `--plain`, `NO_COLOR`, or non-TTY stdout. Otherwise Glamour renders with a JSON overlay (`styleOverrides`) that strips `## `/`### ` heading prefixes and softens link/code styling, then `stripRedundantLinkURLs` collapses OSC 8 hyperlinks whose visible text equals their URL (Glamour's default duplicates the URL after the link text). Width is clamped to `[40, 120]`, default 80.

## Commit messages — **load-bearing for releases**

**Every commit that lands on `main` must follow [Conventional Commits](https://www.conventionalcommits.org/).** release-please parses commit messages to decide whether to bump the version and what to write in the changelog — non-conforming commits are silently ignored. A merge that contains only off-spec subjects produces no release PR update.

Required prefixes (anything else is dropped by release-please):

| Prefix | Triggers a release? | Changelog section |
|---|---|---|
| `feat:` | yes — minor bump (pre-1.0) | Features |
| `fix:` | **no** while pre-1.0 (see below) | Bug Fixes |
| `feat!:` / `fix!:` / `BREAKING CHANGE:` footer | yes — minor (pre-1.0), major (post-1.0) | flagged in changelog |
| `perf:` / `revert:` / `refactor:` / `docs:` / `build:` | no | own section |
| `ci:` / `chore:` / `test:` / `style:` | no | hidden |

Pre-1.0 nuance from `release-please-config.json`: `bump-minor-pre-major: true`, `bump-patch-for-minor-pre-major: false`. While we're below `1.0.0`:

- `feat:` → minor bump (e.g. `0.6.0` → `0.7.0`)
- `fix:` → **no bump at all** (the changelog entry still appears once a `feat:` triggers a release)
- breaking changes → minor bump, not major

When merging PRs via the GitHub UI, the **PR title** becomes the squash-merge subject — that subject must follow Conventional Commits. Off-spec PR titles are the most common cause of "release-please didn't open a release PR."

If you need to ship a fix without waiting for an unrelated `feat:`, bundle it with any trivially conventional `feat:` or land them together; do not retitle a `fix:` as `feat:` just to force a release.

## Release flow

Release-please + goreleaser, both wired in `.github/workflows/release-please.yml`:

1. Conventional-commit pushes to `main` keep an open "chore: release vX.Y.Z" PR (config: `release-please-config.json`, current version pinned in `.release-please-manifest.json`).
2. Merging that PR creates the `vX.Y.Z` tag and GitHub release.
3. A follow-up job in the same workflow (gated on `release_created == 'true'`) checks out the new tag and runs goreleaser, which publishes:
   - GitHub release archives (`.tar.gz` / `.zip`)
   - Homebrew formula (`peaster/homebrew-tap`, requires `HOMEBREW_TAP_GITHUB_TOKEN`)
   - `.deb` packages via nfpm
   - AUR `git-daily-bin` (requires `AUR_KEY`)

Tags follow `vX.Y.Z`.
