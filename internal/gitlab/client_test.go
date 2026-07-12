package gitlab

import (
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name        string
		remoteURL   string
		wantHost    string
		wantPath    string
		wantErr     bool
	}{
		{
			name:      "SSH URL",
			remoteURL: "git@gitlab.com:user/project.git",
			wantHost:  "gitlab.com",
			wantPath:  "user%2Fproject",
			wantErr:   false,
		},
		{
			name:      "SSH URL with subgroups",
			remoteURL: "git@gitlab.example.com:group/subgroup/project.git",
			wantHost:  "gitlab.example.com",
			wantPath:  "group%2Fsubgroup%2Fproject",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL",
			remoteURL: "https://gitlab.com/user/project.git",
			wantHost:  "gitlab.com",
			wantPath:  "user%2Fproject",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with subgroups",
			remoteURL: "https://gitlab.example.com/group/subgroup/project",
			wantHost:  "gitlab.example.com",
			wantPath:  "group%2Fsubgroup%2Fproject",
			wantErr:   false,
		},
		{
			name:      "HTTP URL",
			remoteURL: "http://gitlab.local/user/project",
			wantHost:  "gitlab.local",
			wantPath:  "user%2Fproject",
			wantErr:   false,
		},
		{
			name:      "Invalid URL",
			remoteURL: "invalid-url",
			wantHost:  "",
			wantPath:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, path, err := ParseRemoteURL(tt.remoteURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if host != tt.wantHost {
				t.Errorf("ParseRemoteURL() host = %v, wantHost %v", host, tt.wantHost)
			}
			if path != tt.wantPath {
				t.Errorf("ParseRemoteURL() path = %v, wantPath %v", path, tt.wantPath)
			}
		})
	}
}
