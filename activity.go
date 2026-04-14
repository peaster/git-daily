package main

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Commit struct {
	Hash       string
	Time       string
	Subject    string
	Ref        string
	Files      int
	Insertions int
	Deletions  int
}

type RepoActivity struct {
	Name    string
	Commits []Commit
}

var (
	reFiles = regexp.MustCompile(`(\d+) file`)
	reIns   = regexp.MustCompile(`(\d+) insertion`)
	reDel   = regexp.MustCompile(`(\d+) deletion`)
)

func extractActivity(repos []string, date string, authors []string) []RepoActivity {
	var activities []RepoActivity

	for _, repo := range repos {
		args := []string{"-C", repo, "log", "--all",
			"--after=" + date + "T00:00:00",
			"--before=" + date + "T23:59:59",
			"--pretty=format:%h|%aI|%s|%S",
			"--no-merges",
		}
		for _, a := range authors {
			args = append(args, "--author="+a)
		}

		output := runGit(args...)
		if output == "" {
			continue
		}

		ra := RepoActivity{Name: filepath.Base(repo)}

		for _, line := range strings.Split(output, "\n") {
			parts := strings.SplitN(line, "|", 4)
			if len(parts) < 4 {
				continue
			}

			c := Commit{
				Hash:    parts[0],
				Subject: parts[2],
				Ref:     cleanRef(parts[3]),
			}

			if t, err := time.Parse(time.RFC3339, parts[1]); err == nil {
				c.Time = t.Format("15:04")
			}

			c.Files, c.Insertions, c.Deletions = parseDiffStat(
				runGit("-C", repo, "diff", "--shortstat", parts[0]+"^.."+parts[0]),
			)

			ra.Commits = append(ra.Commits, c)
		}

		if len(ra.Commits) > 0 {
			activities = append(activities, ra)
		}
	}

	return activities
}

func cleanRef(ref string) string {
	ref = strings.TrimPrefix(ref, "refs/original/refs/heads/")
	ref = strings.TrimPrefix(ref, "refs/heads/")
	if strings.HasPrefix(ref, "refs/remotes/") {
		if _, after, ok := strings.Cut(ref[len("refs/remotes/"):], "/"); ok {
			ref = after
		}
	}
	ref = strings.Replace(ref, "refs/tags/", "tag:", 1)
	return ref
}

func parseDiffStat(output string) (files, ins, del int) {
	if m := reFiles.FindStringSubmatch(output); len(m) > 1 {
		files, _ = strconv.Atoi(m[1])
	}
	if m := reIns.FindStringSubmatch(output); len(m) > 1 {
		ins, _ = strconv.Atoi(m[1])
	}
	if m := reDel.FindStringSubmatch(output); len(m) > 1 {
		del, _ = strconv.Atoi(m[1])
	}
	return
}
