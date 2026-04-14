package main

import (
	"fmt"
	"strings"
)

func formatMarkdown(date string, activities []RepoActivity) string {
	var totalCommits, totalFiles, totalIns, totalDel int
	for _, ra := range activities {
		totalCommits += len(ra.Commits)
		for _, c := range ra.Commits {
			totalFiles += c.Files
			totalIns += c.Insertions
			totalDel += c.Deletions
		}
	}

	if totalCommits == 0 {
		return ""
	}

	var b strings.Builder

	fmt.Fprintf(&b, "## Git Activity \u2014 %s\n\n", date)
	fmt.Fprintf(&b, "> %d commits across repos \u00b7 +%d/-%d lines \u00b7 %d files changed\n\n",
		totalCommits, totalIns, totalDel, totalFiles)

	for _, ra := range activities {
		fmt.Fprintf(&b, "### %s (%d commits)\n", ra.Name, len(ra.Commits))
		for _, c := range ra.Commits {
			b.WriteString("- ")
			if c.Time != "" {
				fmt.Fprintf(&b, "**%s** ", c.Time)
			}
			fmt.Fprintf(&b, "`%s` %s", c.Hash, c.Subject)
			if c.Insertions > 0 || c.Deletions > 0 {
				fmt.Fprintf(&b, " (+%d/-%d)", c.Insertions, c.Deletions)
			}
			if c.Ref != "" {
				fmt.Fprintf(&b, " \u00b7 `%s`", c.Ref)
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	return b.String()
}
