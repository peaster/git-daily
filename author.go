package main

import (
	"errors"
	"os"
	"strings"
)

func resolveAuthors() ([]string, error) {
	if env := os.Getenv("GIT_DAILY_AUTHORS"); env != "" {
		var authors []string
		for _, a := range strings.Split(env, ",") {
			if a = strings.TrimSpace(a); a != "" {
				authors = append(authors, a)
			}
		}
		if len(authors) > 0 {
			return authors, nil
		}
	}

	var authors []string
	if email := runGit("config", "user.email"); email != "" {
		authors = append(authors, email)
	}
	if name := runGit("config", "user.name"); name != "" {
		authors = append(authors, name)
	}
	if len(authors) == 0 {
		return nil, errors.New("cannot determine git user. Set GIT_DAILY_AUTHORS or git config user.name/user.email")
	}
	return authors, nil
}
