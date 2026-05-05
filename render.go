package main

import (
	"fmt"
	"os"
	"regexp"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

const (
	defaultWidth = 80
	minWidth     = 40
	maxWidth     = 120
)

// styleOverrides layers on top of whichever Glamour base style is in use to
// soften its defaults: drop the literal "## "/"### " heading prefixes, remove
// the padded-block treatment around inline code, and dim the URL portion of
// links so it does not visually compete with the linked text (terminals that
// support OSC 8 hyperlinks make the linked text itself clickable).
var styleOverrides = []byte(`{
  "h2": { "prefix": "" },
  "h3": { "prefix": "" },
  "code": { "prefix": "", "suffix": "", "background_color": null },
  "link": { "color": "240", "underline": false }
}`)

// hyperlinkPattern matches a single OSC 8 hyperlink: open + visible content
// (no inner ESC) + close. We use it to drop the visible-URL portion that
// Glamour appends after a styled link text when the visible content equals
// the URL itself — the hash text is already clickable via OSC 8.
var hyperlinkPattern = regexp.MustCompile("\x1b\\]8;[^;]*;([^\x07]+)\x07([^\x1b]+)\x1b\\]8;;\x07")

// renderForOutput writes the markdown to stdout, choosing between rich ANSI
// rendering and raw passthrough based on environment, flags, and TTY state.
func renderForOutput(markdown string, forcePlain bool, styleOverride string) {
	if shouldOutputPlain(forcePlain) {
		fmt.Print(markdown)
		return
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath(resolveStyle(styleOverride)),
		glamour.WithStylesFromJSONBytes(styleOverrides),
		glamour.WithWordWrap(terminalWidth()),
	)
	if err != nil {
		fmt.Print(markdown)
		return
	}

	out, err := r.Render(markdown)
	if err != nil {
		fmt.Print(markdown)
		return
	}

	lipgloss.Print(stripRedundantLinkURLs(out))
}

func stripRedundantLinkURLs(s string) string {
	return hyperlinkPattern.ReplaceAllStringFunc(s, func(match string) string {
		groups := hyperlinkPattern.FindStringSubmatch(match)
		if len(groups) == 3 && groups[1] == groups[2] {
			return ""
		}
		return match
	})
}

func shouldOutputPlain(forcePlain bool) bool {
	if forcePlain {
		return true
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

func resolveStyle(styleOverride string) string {
	if styleOverride != "" {
		return styleOverride
	}
	if s := os.Getenv("GLAMOUR_STYLE"); s != "" {
		return s
	}
	if lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
		return "dark"
	}
	return "light"
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return defaultWidth
	}
	if w < minWidth {
		return minWidth
	}
	if w > maxWidth {
		return maxWidth
	}
	return w
}
