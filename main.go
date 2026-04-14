package main

import (
	"fmt"
	"os"
	"regexp"
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

	dateInput, repoDirs := parseArgs(os.Args[1:])

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

	fmt.Print(output)
}

func parseArgs(args []string) (string, []string) {
	var dateInput string
	var repoDirs []string

	for _, arg := range args {
		if dateInput == "" && isDateArg(arg) {
			dateInput = arg
		} else {
			repoDirs = append(repoDirs, arg)
		}
	}

	if dateInput == "" {
		dateInput = "today"
	}
	if len(repoDirs) == 0 {
		repoDirs = []string{"."}
	}

	return dateInput, repoDirs
}

func isDateArg(s string) bool {
	return s == "today" || s == "yesterday" || datePattern.MatchString(s)
}

func printHelp() {
	fmt.Print(`git-daily — Extract your git activity across repos for a specific day.

USAGE
  git-daily [OPTIONS] [DATE] [REPO_DIR...]

OPTIONS
  -h, --help       Show this help message and exit.
  -v, --version    Print version and exit.

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

BEHAVIOR
  - Matches commits by any of the configured author identities.
  - Searches all branches and tags (--all), reporting the source
    branch or tag for each commit.
  - Merge commits are excluded.
  - For each commit: reports short hash, commit message, time of day,
    insertions/deletions, and source ref (branch or tag).
  - If no commits are found, prints a message and exits with code 0.

OUTPUT
  Markdown printed to stdout:

    ## Git Activity — YYYY-MM-DD

    > N commits across repos · +INSERTIONS/-DELETIONS lines · FILES files changed

    ### repo-name (N commits)
    - **HH:MM** ` + "`" + `shorthash` + "`" + ` Commit message (+12/-3) · ` + "`" + `branch-name` + "`" + `

EXAMPLES
  git-daily                              # today, repos under cwd
  git-daily 2026-04-14                   # specific day, repos under cwd
  git-daily 2026-04-14 ~/projects        # specific day, specific directory
  git-daily yesterday ~/work ~/oss       # yesterday, multiple search roots
  ACTIVITY=$(git-daily 2026-04-14)       # capture output into a variable

EXIT CODES
  0    Success (including when no commits are found).
  1    Invalid date format or no git user configured.
`)
}
