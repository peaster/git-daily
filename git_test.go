package main

import "testing"

func TestCommitURL(t *testing.T) {
	const sha = "abc1234"

	cases := []struct {
		name   string
		remote string
		want   string
	}{
		{"empty", "", ""},
		{"scp-style github", "git@github.com:peaster/git-daily.git", "https://github.com/peaster/git-daily/commit/abc1234"},
		{"scp-style no .git", "git@github.com:peaster/git-daily", "https://github.com/peaster/git-daily/commit/abc1234"},
		{"https github", "https://github.com/peaster/git-daily.git", "https://github.com/peaster/git-daily/commit/abc1234"},
		{"https with token", "https://token@github.com/owner/repo", "https://github.com/owner/repo/commit/abc1234"},
		{"https with user:pass", "https://user:pass@gitlab.com/group/proj.git", "https://gitlab.com/group/proj/commit/abc1234"},
		{"ssh:// with port", "ssh://git@gitlab.com:22/group/proj.git", "https://gitlab.com/group/proj/commit/abc1234"},
		{"ssh:// no port", "ssh://git@gitea.local/u/r.git", "https://gitea.local/u/r/commit/abc1234"},
		{"git:// upgraded", "git://github.com/owner/repo.git", "https://github.com/owner/repo/commit/abc1234"},
		{"http preserved", "http://gitea.local/u/r.git", "http://gitea.local/u/r/commit/abc1234"},
		{"trailing slash", "https://github.com/owner/repo/", "https://github.com/owner/repo/commit/abc1234"},
		{"unparseable", "not a url", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := commitURL(tc.remote, sha)
			if got != tc.want {
				t.Errorf("commitURL(%q, %q) = %q; want %q", tc.remote, sha, got, tc.want)
			}
		})
	}
}
