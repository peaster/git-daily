package main

import (
	"os/exec"
	"strings"
)

// runGit executes a git command and returns its stdout.
// Returns "" on any error (matching the bash || true pattern).
func runGit(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\n")
}

// remoteURL returns a remote URL for repo, preferring "origin" and falling
// back to the first configured remote. Returns "" if none are configured.
func remoteURL(repo string) string {
	if url := runGit("-C", repo, "remote", "get-url", "origin"); url != "" {
		return url
	}
	remotes := runGit("-C", repo, "remote")
	if remotes == "" {
		return ""
	}
	first := strings.SplitN(remotes, "\n", 2)[0]
	return runGit("-C", repo, "remote", "get-url", first)
}

// commitURL converts a git remote URL into a browseable commit URL of the
// form <base>/commit/<sha>. The /commit/<sha> suffix is honored by GitHub,
// GitLab, Gitea/Forgejo, Codeberg, Azure DevOps, and sourcehut.
// Returns "" if remote is empty or cannot be normalized.
func commitURL(remote, sha string) string {
	base := normalizeRemoteURL(remote)
	if base == "" {
		return ""
	}
	return base + "/commit/" + sha
}

// normalizeRemoteURL turns any common git URL form into an https:// (or
// http:// for plain self-hosted) base URL with no trailing .git.
func normalizeRemoteURL(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}

	var base string
	switch {
	case strings.HasPrefix(remote, "https://"), strings.HasPrefix(remote, "http://"):
		base = stripUserInfo(remote)
	case strings.HasPrefix(remote, "ssh://"):
		base = "https://" + dropPort(stripScheme(remote, "ssh://"))
	case strings.HasPrefix(remote, "git://"):
		base = "https://" + stripScheme(remote, "git://")
	default:
		// SCP-style: [user@]host:path
		if at := strings.Index(remote, "@"); at != -1 {
			remote = remote[at+1:]
		}
		colon := strings.Index(remote, ":")
		if colon == -1 {
			return ""
		}
		base = "https://" + remote[:colon] + "/" + remote[colon+1:]
	}

	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, ".git")
	return base
}

// stripScheme removes a leading scheme and any user@ prefix.
func stripScheme(url, scheme string) string {
	url = strings.TrimPrefix(url, scheme)
	if at := strings.Index(url, "@"); at != -1 {
		if slash := strings.Index(url, "/"); slash == -1 || at < slash {
			url = url[at+1:]
		}
	}
	return url
}

// stripUserInfo removes user[:pass]@ from an http(s) URL.
func stripUserInfo(url string) string {
	scheme := "https://"
	if strings.HasPrefix(url, "http://") {
		scheme = "http://"
	}
	rest := url[len(scheme):]
	if at := strings.Index(rest, "@"); at != -1 {
		if slash := strings.Index(rest, "/"); slash == -1 || at < slash {
			rest = rest[at+1:]
		}
	}
	return scheme + rest
}

// dropPort removes a :port segment from host[:port]/path.
func dropPort(s string) string {
	slash := strings.Index(s, "/")
	host := s
	rest := ""
	if slash != -1 {
		host = s[:slash]
		rest = s[slash:]
	}
	if colon := strings.Index(host, ":"); colon != -1 {
		host = host[:colon]
	}
	return host + rest
}
