package main

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type EventKind int

const (
	KindCommit EventKind = iota
	KindMerge
	KindBranch
	KindTag
	numKinds
)

var kindLabels = [numKinds]struct {
	Singular, Plural string
}{
	KindCommit: {"commit", "commits"},
	KindMerge:  {"merge", "merges"},
	KindBranch: {"branch", "branches"},
	KindTag:    {"tag", "tags"},
}

type Event struct {
	Kind       EventKind
	When       time.Time
	Time       string
	Hash       string
	Subject    string
	Ref        string
	Name       string
	Files      int
	Insertions int
	Deletions  int
}

type RepoActivity struct {
	Name       string
	RemoteURL  string
	Events     []Event
	Counts     [numKinds]int
	TotalFiles int
	TotalIns   int
	TotalDel   int
}

func (ra *RepoActivity) addEvent(ev Event) {
	ra.Events = append(ra.Events, ev)
	ra.Counts[ev.Kind]++
}

func inDayBounds(t, start, end time.Time) bool {
	return !t.Before(start) && t.Before(end)
}

func trimEmailBrackets(s string) string {
	return strings.Trim(strings.TrimSpace(s), "<>")
}

var (
	reFiles = regexp.MustCompile(`(\d+) file`)
	reIns   = regexp.MustCompile(`(\d+) insertion`)
	reDel   = regexp.MustCompile(`(\d+) deletion`)
)

func extractActivity(repos []string, date string, authors []string) []RepoActivity {
	dayStart, dayEnd, err := dayBounds(date)
	if err != nil {
		return nil
	}

	var activities []RepoActivity

	for _, repo := range repos {
		ra := RepoActivity{Name: filepath.Base(repo), RemoteURL: remoteURL(repo)}

		extractCommitsAndMerges(repo, date, authors, &ra)
		extractBranches(repo, dayStart, dayEnd, authors, &ra)
		extractTags(repo, dayStart, dayEnd, authors, &ra)

		if len(ra.Events) == 0 {
			continue
		}

		sort.SliceStable(ra.Events, func(i, j int) bool {
			return ra.Events[i].When.After(ra.Events[j].When)
		})

		activities = append(activities, ra)
	}

	return activities
}

func dayBounds(date string) (time.Time, time.Time, error) {
	loc := time.Local
	start, err := time.ParseInLocation(time.DateOnly, date, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end := start.AddDate(0, 0, 1)
	return start, end, nil
}

func extractCommitsAndMerges(repo, date string, authors []string, ra *RepoActivity) {
	args := []string{"-C", repo, "log", "--all", "--shortstat",
		"--after=" + date + "T00:00:00",
		"--before=" + date + "T23:59:59",
		"--pretty=format:%h|%aI|%s|%S|%P",
	}
	for _, a := range authors {
		args = append(args, "--author="+a)
	}

	output := runGit(args...)
	if output == "" {
		return
	}

	lastIdx := -1

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "|") {
			parts := strings.SplitN(line, "|", 5)
			if len(parts) < 5 {
				continue
			}

			ev := Event{
				Hash:    parts[0],
				Subject: parts[2],
				Ref:     cleanRef(parts[3]),
			}

			if t, err := time.Parse(time.RFC3339, parts[1]); err == nil {
				ev.When = t
				ev.Time = t.Local().Format("15:04")
			}

			if strings.Contains(strings.TrimSpace(parts[4]), " ") {
				ev.Kind = KindMerge
			} else {
				ev.Kind = KindCommit
			}

			ra.addEvent(ev)
			lastIdx = len(ra.Events) - 1
		} else if lastIdx >= 0 && strings.HasPrefix(line, " ") {
			files, ins, del := parseDiffStat(line)
			ev := &ra.Events[lastIdx]
			ev.Files, ev.Insertions, ev.Deletions = files, ins, del
			ra.TotalFiles += files
			ra.TotalIns += ins
			ra.TotalDel += del
		}
	}
}

func extractBranches(repo string, dayStart, dayEnd time.Time, authors []string, ra *RepoActivity) {
	branchList := runGit("-C", repo, "for-each-ref", "refs/heads", "--format=%(refname:short)")
	if branchList == "" {
		return
	}

	for _, branch := range strings.Split(branchList, "\n") {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}

		out := runGit("-C", repo, "log", "-g", "--max-count=1",
			"--grep-reflog=^branch: Created from ",
			"--date=iso-strict", "--pretty=format:%h|%gd|%ce|%gs", branch)
		if out == "" {
			continue
		}

		oldest := strings.SplitN(out, "\n", 2)[0]
		ev, ok := parseBranchReflogLine(oldest)
		if !ok {
			continue
		}
		if !inDayBounds(ev.When, dayStart, dayEnd) {
			continue
		}
		if !authorMatches(ev.subjectAuthor, authors) {
			continue
		}

		ra.addEvent(Event{
			Kind: KindBranch,
			When: ev.When,
			Time: ev.When.Local().Format("15:04"),
			Name: branch,
			Hash: ev.Hash,
			Ref:  ev.Source,
		})
	}
}

type branchReflogEntry struct {
	Hash          string
	When          time.Time
	subjectAuthor string
	Source        string
}

func parseBranchReflogLine(line string) (branchReflogEntry, bool) {
	parts := strings.SplitN(line, "|", 4)
	if len(parts) < 4 {
		return branchReflogEntry{}, false
	}

	gd := parts[1]
	open := strings.Index(gd, "{")
	closeIdx := strings.LastIndex(gd, "}")
	if open == -1 || closeIdx == -1 || closeIdx <= open {
		return branchReflogEntry{}, false
	}
	t, err := time.Parse(time.RFC3339, gd[open+1:closeIdx])
	if err != nil {
		return branchReflogEntry{}, false
	}

	rest, ok := strings.CutPrefix(parts[3], "branch: Created from ")
	if !ok {
		return branchReflogEntry{}, false
	}

	return branchReflogEntry{
		Hash:          parts[0],
		When:          t,
		subjectAuthor: parts[2],
		Source:        cleanRef(strings.TrimSpace(rest)),
	}, true
}

func extractTags(repo string, dayStart, dayEnd time.Time, authors []string, ra *RepoActivity) {
	const sep = "\x1f"
	format := strings.Join([]string{
		"%(refname:short)",
		"%(objecttype)",
		"%(creatordate:iso-strict)",
		"%(taggeremail)",
		"%(authoremail)",
		"%(subject)",
		"%(*objectname:short)",
		"%(objectname:short)",
	}, sep)

	out := runGit("-C", repo, "for-each-ref", "refs/tags", "--format="+format)
	if out == "" {
		return
	}

	for _, line := range strings.Split(out, "\n") {
		ev, ok := parseTagRefLine(line, sep)
		if !ok {
			continue
		}
		if !inDayBounds(ev.When, dayStart, dayEnd) {
			continue
		}

		var emailToCheck string
		if ev.IsAnnotated {
			emailToCheck = ev.TaggerEmail
		} else {
			emailToCheck = ev.AuthorEmail
		}
		if !authorMatches(emailToCheck, authors) {
			continue
		}

		ra.addEvent(Event{
			Kind:    KindTag,
			When:    ev.When,
			Time:    ev.When.Local().Format("15:04"),
			Name:    ev.Name,
			Hash:    ev.CommitShort,
			Subject: ev.Subject,
		})
	}
}

type tagRefEntry struct {
	Name        string
	IsAnnotated bool
	When        time.Time
	TaggerEmail string
	AuthorEmail string
	Subject     string
	CommitShort string
}

func parseTagRefLine(line, sep string) (tagRefEntry, bool) {
	parts := strings.Split(line, sep)
	if len(parts) < 8 {
		return tagRefEntry{}, false
	}
	t, err := time.Parse(time.RFC3339, parts[2])
	if err != nil {
		return tagRefEntry{}, false
	}
	annotated := parts[1] == "tag"
	commitShort := parts[6]
	if commitShort == "" {
		commitShort = parts[7]
	}
	return tagRefEntry{
		Name:        parts[0],
		IsAnnotated: annotated,
		When:        t,
		TaggerEmail: trimEmailBrackets(parts[3]),
		AuthorEmail: trimEmailBrackets(parts[4]),
		Subject:     parts[5],
		CommitShort: commitShort,
	}, true
}

func authorMatches(candidate string, authors []string) bool {
	candidate = strings.ToLower(trimEmailBrackets(candidate))
	if candidate == "" {
		return false
	}
	for _, a := range authors {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" {
			continue
		}
		if strings.Contains(candidate, a) {
			return true
		}
	}
	return false
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
