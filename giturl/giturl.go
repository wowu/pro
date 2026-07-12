package giturl

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type GitURL struct {
	// Host is the bare hostname, without any port (e.g. "github.com").
	Host string

	// Path is the repository path with a leading slash (e.g. "/owner/repo.git").
	Path string
}

// scpSyntax matches SCP-like git URLs such as "git@github.com:owner/repo.git".
// The optional user is discarded; group 1 is the host, group 2 is the path.
var scpSyntax = regexp.MustCompile(`^(?:[^@/]+@)?([^:/]+):(.*)$`)

// Parse parses a git remote URL into its host and path.
func Parse(raw string) (*GitURL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty git URL")
	}

	// SCP-like syntax (host:path) has no scheme, so no "://".
	if !strings.Contains(raw, "://") {
		m := scpSyntax.FindStringSubmatch(raw)
		if m == nil {
			return nil, fmt.Errorf("unable to parse git URL: %q", raw)
		}
		return &GitURL{Host: m[1], Path: normalizePath(m[2])}, nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("unable to parse git URL %q: %w", raw, err)
	}

	host := u.Hostname() // strips any port
	if host == "" {
		return nil, fmt.Errorf("git URL %q has no host", raw)
	}

	return &GitURL{Host: host, Path: normalizePath(u.Path)}, nil
}

// normalizePath ensures the path starts with a single leading slash.
func normalizePath(p string) string {
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}
