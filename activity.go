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
	Name       string
	Commits    []Commit
	TotalFiles int
	TotalIns   int
	TotalDel   int
}

var (
	reFiles = regexp.MustCompile(`(\d+) file`)
	reIns   = regexp.MustCompile(`(\d+) insertion`)
	reDel   = regexp.MustCompile(`(\d+) deletion`)
)

func extractActivity(repos []string, date string, authors []string) []RepoActivity {
	var activities []RepoActivity

	for _, repo := range repos {
		args := []string{"-C", repo, "log", "--all", "--shortstat",
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
		lastIdx := -1

		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "|") {
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

				ra.Commits = append(ra.Commits, c)
				lastIdx = len(ra.Commits) - 1
			} else if lastIdx >= 0 && strings.HasPrefix(line, " ") {
				files, ins, del := parseDiffStat(line)
				c := &ra.Commits[lastIdx]
				c.Files, c.Insertions, c.Deletions = files, ins, del
				ra.TotalFiles += files
				ra.TotalIns += ins
				ra.TotalDel += del
			}
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
		// Strip refs/remotes/<remote>/
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
