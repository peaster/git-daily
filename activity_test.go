package main

import (
	"testing"
	"time"
)

func TestParseBranchReflogLine(t *testing.T) {
	cases := []struct {
		name        string
		line        string
		ok          bool
		wantSource  string
		wantAuthor  string
		wantHash    string
		wantTimeStr string
	}{
		{
			name:        "creation entry",
			line:        "abc1234|feature/foo@{2026-05-07T10:23:45-05:00}|user@example.com|branch: Created from main",
			ok:          true,
			wantSource:  "main",
			wantAuthor:  "user@example.com",
			wantHash:    "abc1234",
			wantTimeStr: "2026-05-07T10:23:45-05:00",
		},
		{
			name:       "creation from refs/heads/",
			line:       "deadbeef|topic@{2026-05-07T08:00:00+00:00}|x@y.io|branch: Created from refs/heads/main",
			ok:         true,
			wantSource: "main",
			wantAuthor: "x@y.io",
			wantHash:   "deadbeef",
		},
		{
			name:       "non-creation reflog entry has empty source",
			line:       "abc1234|main@{2026-05-07T10:00:00-05:00}|user@example.com|commit: fix something",
			ok:         true,
			wantSource: "",
			wantAuthor: "user@example.com",
			wantHash:   "abc1234",
		},
		{
			name: "missing fields fails",
			line: "abc1234|main@{2026-05-07T10:00:00-05:00}",
			ok:   false,
		},
		{
			name: "malformed gd selector fails",
			line: "abc1234|nobraces|user@example.com|branch: Created from main",
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseBranchReflogLine(tc.line)
			if ok != tc.ok {
				t.Fatalf("parseBranchReflogLine ok = %v; want %v", ok, tc.ok)
			}
			if !ok {
				return
			}
			if got.Source != tc.wantSource {
				t.Errorf("Source = %q; want %q", got.Source, tc.wantSource)
			}
			if got.subjectAuthor != tc.wantAuthor {
				t.Errorf("subjectAuthor = %q; want %q", got.subjectAuthor, tc.wantAuthor)
			}
			if got.Hash != tc.wantHash {
				t.Errorf("Hash = %q; want %q", got.Hash, tc.wantHash)
			}
			if tc.wantTimeStr != "" {
				want, _ := time.Parse(time.RFC3339, tc.wantTimeStr)
				if !got.When.Equal(want) {
					t.Errorf("When = %v; want %v", got.When, want)
				}
			}
		})
	}
}

func TestParseTagRefLine(t *testing.T) {
	const sep = "\x1f"
	join := func(fields ...string) string {
		out := ""
		for i, f := range fields {
			if i > 0 {
				out += sep
			}
			out += f
		}
		return out
	}

	t.Run("annotated tag", func(t *testing.T) {
		line := join(
			"v1.2.0",
			"tag",
			"2026-05-07T10:05:00-05:00",
			"<tagger@example.com>",
			"<author@example.com>",
			"Release 1.2.0",
			"abc1234",
			"deadbee",
		)
		got, ok := parseTagRefLine(line, sep)
		if !ok {
			t.Fatalf("parseTagRefLine ok = false")
		}
		if got.Name != "v1.2.0" {
			t.Errorf("Name = %q", got.Name)
		}
		if !got.IsAnnotated {
			t.Errorf("IsAnnotated = false; want true")
		}
		if got.TaggerEmail != "tagger@example.com" {
			t.Errorf("TaggerEmail = %q", got.TaggerEmail)
		}
		if got.AuthorEmail != "author@example.com" {
			t.Errorf("AuthorEmail = %q", got.AuthorEmail)
		}
		if got.Subject != "Release 1.2.0" {
			t.Errorf("Subject = %q", got.Subject)
		}
		if got.CommitShort != "abc1234" {
			t.Errorf("CommitShort = %q; want abc1234", got.CommitShort)
		}
	})

	t.Run("lightweight tag falls back to objectname", func(t *testing.T) {
		line := join(
			"v1.0",
			"commit",
			"2026-05-07T09:00:00-05:00",
			"",
			"<author@example.com>",
			"initial commit",
			"",
			"feedface",
		)
		got, ok := parseTagRefLine(line, sep)
		if !ok {
			t.Fatalf("parseTagRefLine ok = false")
		}
		if got.IsAnnotated {
			t.Errorf("IsAnnotated = true; want false")
		}
		if got.AuthorEmail != "author@example.com" {
			t.Errorf("AuthorEmail = %q", got.AuthorEmail)
		}
		if got.CommitShort != "feedface" {
			t.Errorf("CommitShort = %q; want feedface", got.CommitShort)
		}
	})

	t.Run("unparseable date fails", func(t *testing.T) {
		line := join("v1", "tag", "not-a-date", "", "", "", "", "abc1234")
		_, ok := parseTagRefLine(line, sep)
		if ok {
			t.Errorf("expected ok=false for bad date")
		}
	})
}

func TestAuthorMatches(t *testing.T) {
	authors := []string{"peaster", "paul@work.com", "Paul Easterbrooks"}

	cases := []struct {
		candidate string
		want      bool
	}{
		{"paul@work.com", true},
		{"<paul@work.com>", true},
		{"PAUL@WORK.COM", true},
		{"peaster", true},
		{"someone-peaster-else@x.com", true},
		{"paul easterbrooks", true},
		{"other@x.com", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.candidate, func(t *testing.T) {
			if got := authorMatches(tc.candidate, authors); got != tc.want {
				t.Errorf("authorMatches(%q) = %v; want %v", tc.candidate, got, tc.want)
			}
		})
	}
}
