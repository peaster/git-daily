package main

import (
	"fmt"
	"strings"
)

func formatMarkdown(date string, activities []RepoActivity) string {
	var totals [numKinds]int
	var totalFiles, totalIns, totalDel int
	for _, ra := range activities {
		for k, n := range ra.Counts {
			totals[k] += n
		}
		totalFiles += ra.TotalFiles
		totalIns += ra.TotalIns
		totalDel += ra.TotalDel
	}

	totalEvents := 0
	for _, n := range totals {
		totalEvents += n
	}
	if totalEvents == 0 {
		return ""
	}

	var b strings.Builder

	fmt.Fprintf(&b, "## Git Activity — %s\n\n", date)
	fmt.Fprintf(&b, "> %s across repos · +%d/-%d lines · %d files changed\n\n",
		summarizeCounts(totals),
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

func summarizeCounts(counts [numKinds]int) string {
	parts := make([]string, 0, numKinds)
	for k, n := range counts {
		if n > 0 {
			parts = append(parts, plural(n, kindLabels[k].Singular, kindLabels[k].Plural))
		}
	}
	return strings.Join(parts, ", ")
}

func repoHeading(ra RepoActivity) string {
	seen, only := 0, 0
	for k, n := range ra.Counts {
		if n > 0 {
			seen++
			only = k
		}
	}
	if seen == 1 {
		return plural(ra.Counts[only], kindLabels[only].Singular, kindLabels[only].Plural)
	}
	return plural(len(ra.Events), "event", "events")
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralForm)
}
