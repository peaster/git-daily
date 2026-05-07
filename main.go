package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	version     = "dev"
	datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func main() {
	// Check for help flag before any other processing
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		}
		if arg == "-v" || arg == "--version" {
			fmt.Println("git-daily " + version)
			return
		}
	}

	dateInput, repoDirs, plain, style, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	date, err := resolveDate(dateInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	authors, err := resolveAuthors()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	repos := findRepos(repoDirs)
	activities := extractActivity(repos, date, authors)
	output := formatMarkdown(date, activities)

	if output == "" {
		fmt.Printf("No git activity found for %s.\n", date)
		return
	}

	renderForOutput(output, plain, style)
}

func parseArgs(args []string) (dateInput string, repoDirs []string, plain bool, style string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--plain" || arg == "--no-color":
			plain = true
		case arg == "--style":
			if i+1 >= len(args) {
				return "", nil, false, "", fmt.Errorf("--style requires a value")
			}
			i++
			style = args[i]
		case strings.HasPrefix(arg, "--style="):
			style = strings.TrimPrefix(arg, "--style=")
		case dateInput == "" && isDateArg(arg):
			dateInput = arg
		default:
			repoDirs = append(repoDirs, arg)
		}
	}

	if dateInput == "" {
		dateInput = "today"
	}
	if len(repoDirs) == 0 {
		repoDirs = []string{"."}
	}

	return dateInput, repoDirs, plain, style, nil
}

func isDateArg(s string) bool {
	return s == "today" || s == "yesterday" || datePattern.MatchString(s)
}

func printHelp() {
	fmt.Print(`git-daily — Extract your git activity across repos for a specific day.

USAGE
  git-daily [OPTIONS] [DATE] [REPO_DIR...]

OPTIONS
  -h, --help          Show this help message and exit.
  -v, --version       Print version and exit.
      --plain         Output raw markdown even when stdout is a TTY.
      --no-color      Alias for --plain.
      --style NAME    Glamour style for rendered output. One of:
                        dark, light, dracula, tokyo-night, pink, ascii, notty
                      Or a path to a Glamour JSON style file.
                      Overrides $GLAMOUR_STYLE.

ARGUMENTS
  DATE          The target date for activity lookup.
                Accepts:
                  YYYY-MM-DD   A specific calendar date (e.g. 2026-04-14)
                  today        Today's date (default if omitted)
                  yesterday    Yesterday's date
                Only one date may be specified. Activity is extracted for
                that single calendar day (midnight to midnight, local time).

  REPO_DIR      One or more directories to search for git repositories.
                Each directory is searched recursively up to 3 levels deep
                for .git folders. If a directory is itself a git repo, it
                is included directly. Multiple directories can be specified
                as separate arguments. Defaults to the current directory
                if none are provided.

ENVIRONMENT
  GIT_DAILY_AUTHORS   Comma-separated list of author identities to match.
                      Useful when you commit under multiple emails or names.
                      Example:
                        export GIT_DAILY_AUTHORS="peaster,me@work.com,me@home.com"
                      If not set, falls back to git config user.email and
                      user.name from the current environment.

  NO_COLOR            If set (any value), output raw markdown instead of
                      styled ANSI. See https://no-color.org.

  GLAMOUR_STYLE       Default rendering style when --style is not given.
                      Same accepted values as --style.

BEHAVIOR
  - Surfaces commits, merges, branches created, and tags created.
  - Matches events by any of the configured author identities. Events
    whose author/committer/tagger email cannot be matched are dropped.
  - Searches all branches and tags (--all) for commits and merges,
    reporting the source branch or tag for each.
  - Branch creations are detected via the branch's reflog (oldest
    entry); pruned reflogs may miss older creations.
  - Tag creations use the tagger date for annotated tags and the
    underlying commit's date for lightweight tags.
  - For each commit/merge: reports short hash, message, time of day,
    insertions/deletions, and source ref. Merges are prefixed [merge].
  - The short hash links to the remote's commit URL when a remote is
    configured (origin preferred, else the first remote).
  - If no events are found, prints a message and exits with code 0.
  - When stdout is a terminal, output is rendered with ANSI styling.
    When piped or redirected (e.g. ` + "`" + `git-daily > today.md` + "`" + `,
    ` + "`" + `ACTIVITY=$(git-daily)` + "`" + `), raw markdown is emitted unchanged so
    capture and redirect flows are unaffected.

OUTPUT
  Markdown printed to stdout (rendered to ANSI when on a TTY):

    ## Git Activity — YYYY-MM-DD

    > N commits, M merges, B branches, T tags across repos · +INS/-DEL lines · FILES files changed

    ### repo-name (N events)
    - **HH:MM** [` + "`" + `shorthash` + "`" + `](REMOTE_URL/commit/SHA) Commit message (+12/-3) · ` + "`" + `branch-name` + "`" + `
    - **HH:MM** [merge] [` + "`" + `shorthash` + "`" + `](REMOTE_URL/commit/SHA) Merge subject · ` + "`" + `main` + "`" + `
    - **HH:MM** [branch] Created ` + "`" + `branch-name` + "`" + ` from ` + "`" + `source-ref` + "`" + `
    - **HH:MM** [tag] Created ` + "`" + `tag-name` + "`" + ` ([` + "`" + `shorthash` + "`" + `](REMOTE_URL/commit/SHA)) — Tag message

EXAMPLES
  git-daily                              # today, repos under cwd
  git-daily 2026-04-14                   # specific day, repos under cwd
  git-daily 2026-04-14 ~/projects        # specific day, specific directory
  git-daily yesterday ~/work ~/oss       # yesterday, multiple search roots
  ACTIVITY=$(git-daily 2026-04-14)       # capture output into a variable
  git-daily --plain                      # force raw markdown on a TTY
  git-daily --style dracula              # use the Dracula theme

EXIT CODES
  0    Success (including when no commits are found).
  1    Invalid date format or no git user configured.
`)
}
