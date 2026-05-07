package main

import (
	"fmt"
	"strings"
)

func formatMarkdown(date string, activities []RepoActivity) string {
	var totalCommits, totalMerges, totalBranches, totalTags int
	var totalFiles, totalIns, totalDel int
	for _, ra := range activities {
		totalCommits += ra.Commits
		totalMerges += ra.Merges
		totalBranches += ra.Branches
		totalTags += ra.Tags
		totalFiles += ra.TotalFiles
		totalIns += ra.TotalIns
		totalDel += ra.TotalDel
	}

	totalEvents := totalCommits + totalMerges + totalBranches + totalTags
	if totalEvents == 0 {
		return ""
	}

	var b strings.Builder

	fmt.Fprintf(&b, "## Git Activity — %s\n\n", date)
	fmt.Fprintf(&b, "> %s across repos · +%d/-%d lines · %d files changed\n\n",
		summarizeCounts(totalCommits, totalMerges, totalBranches, totalTags),
		totalIns, totalDel, totalFiles)

	for _, ra := range activities {
		fmt.Fprintf(&b, "### %s (%s)\n", ra.Name, repoHeading(ra))
		for _, ev := range ra.Events {
			b.WriteString("- ")
			renderEvent(&b, ra.RemoteURL, ev)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func renderEvent(b *strings.Builder, remoteURL string, ev Event) {
	if ev.Time != "" {
		fmt.Fprintf(b, "**%s** ", ev.Time)
	}

	switch ev.Kind {
	case KindCommit:
		writeCommitBody(b, remoteURL, ev)
	case KindMerge:
		b.WriteString("[merge] ")
		writeCommitBody(b, remoteURL, ev)
	case KindBranch:
		b.WriteString("[branch] Created ")
		fmt.Fprintf(b, "`%s`", ev.Name)
		if ev.Ref != "" {
			fmt.Fprintf(b, " from `%s`", ev.Ref)
		}
	case KindTag:
		b.WriteString("[tag] Created ")
		fmt.Fprintf(b, "`%s`", ev.Name)
		if ev.Hash != "" {
			if url := commitURL(remoteURL, ev.Hash); url != "" {
				fmt.Fprintf(b, " ([`%s`](%s))", ev.Hash, url)
			} else {
				fmt.Fprintf(b, " (`%s`)", ev.Hash)
			}
		}
		if ev.Subject != "" {
			fmt.Fprintf(b, " — %s", ev.Subject)
		}
	}
}

func writeCommitBody(b *strings.Builder, remoteURL string, ev Event) {
	if url := commitURL(remoteURL, ev.Hash); url != "" {
		fmt.Fprintf(b, "[`%s`](%s) %s", ev.Hash, url, ev.Subject)
	} else {
		fmt.Fprintf(b, "`%s` %s", ev.Hash, ev.Subject)
	}
	if ev.Insertions > 0 || ev.Deletions > 0 {
		fmt.Fprintf(b, " (+%d/-%d)", ev.Insertions, ev.Deletions)
	}
	if ev.Ref != "" {
		fmt.Fprintf(b, " · `%s`", ev.Ref)
	}
}

func summarizeCounts(commits, merges, branches, tags int) string {
	parts := make([]string, 0, 4)
	if commits > 0 {
		parts = append(parts, plural(commits, "commit", "commits"))
	}
	if merges > 0 {
		parts = append(parts, plural(merges, "merge", "merges"))
	}
	if branches > 0 {
		parts = append(parts, plural(branches, "branch", "branches"))
	}
	if tags > 0 {
		parts = append(parts, plural(tags, "tag", "tags"))
	}
	return strings.Join(parts, ", ")
}

func repoHeading(ra RepoActivity) string {
	kindsPresent := 0
	if ra.Commits > 0 {
		kindsPresent++
	}
	if ra.Merges > 0 {
		kindsPresent++
	}
	if ra.Branches > 0 {
		kindsPresent++
	}
	if ra.Tags > 0 {
		kindsPresent++
	}

	if kindsPresent <= 1 {
		switch {
		case ra.Commits > 0:
			return plural(ra.Commits, "commit", "commits")
		case ra.Merges > 0:
			return plural(ra.Merges, "merge", "merges")
		case ra.Branches > 0:
			return plural(ra.Branches, "branch", "branches")
		case ra.Tags > 0:
			return plural(ra.Tags, "tag", "tags")
		}
	}
	return plural(len(ra.Events), "event", "events")
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralForm)
}
