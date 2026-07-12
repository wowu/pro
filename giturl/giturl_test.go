package giturl

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantHost string
		wantPath string
	}{
		{
			name:     "scp github",
			raw:      "git@github.com:owner/repo.git",
			wantHost: "github.com",
			wantPath: "/owner/repo.git",
		},
		{
			name:     "scp without .git suffix",
			raw:      "git@github.com:owner/repo",
			wantHost: "github.com",
			wantPath: "/owner/repo",
		},
		{
			name:     "scp gitlab subgroup",
			raw:      "git@gitlab.com:group/subgroup/repo.git",
			wantHost: "gitlab.com",
			wantPath: "/group/subgroup/repo.git",
		},
		{
			name:     "https github",
			raw:      "https://github.com/owner/repo.git",
			wantHost: "github.com",
			wantPath: "/owner/repo.git",
		},
		{
			name:     "https without .git suffix",
			raw:      "https://github.com/owner/repo",
			wantHost: "github.com",
			wantPath: "/owner/repo",
		},
		{
			name:     "http gitlab subgroup",
			raw:      "http://gitlab.com/group/subgroup/repo.git",
			wantHost: "gitlab.com",
			wantPath: "/group/subgroup/repo.git",
		},
		{
			name:     "ssh scheme",
			raw:      "ssh://git@gitlab.com/group/subgroup/repo.git",
			wantHost: "gitlab.com",
			wantPath: "/group/subgroup/repo.git",
		},
		{
			name:     "ssh scheme with port",
			raw:      "ssh://git@github.com:22/owner/repo.git",
			wantHost: "github.com",
			wantPath: "/owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.raw, err)
			}
			if got.Host != tt.wantHost {
				t.Errorf("Parse(%q).Host = %q, want %q", tt.raw, got.Host, tt.wantHost)
			}
			if got.Path != tt.wantPath {
				t.Errorf("Parse(%q).Path = %q, want %q", tt.raw, got.Path, tt.wantPath)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "whitespace only", raw: "   "},
		{name: "no host or colon", raw: "just-a-string"},
		{name: "scheme without host", raw: "ssh:///owner/repo.git"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(tt.raw); err == nil {
				t.Errorf("Parse(%q) = nil error, want error", tt.raw)
			}
		})
	}
}
