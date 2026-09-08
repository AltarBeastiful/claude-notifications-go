package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/777genius/agent-notifications/internal/analyzer"
)

// stubGitRoot replaces the git lookup for the duration of a test, so project
// naming is exercised without depending on a repository existing on disk.
func stubGitRoot(t *testing.T, roots map[string]string) {
	t.Helper()

	original := getGitRoot
	getGitRoot = func(cwd string) string { return roots[cwd] }
	t.Cleanup(func() { getGitRoot = original })
}

func TestTitleIdentity(t *testing.T) {
	stubGitRoot(t, map[string]string{
		"/home/remi/projects/DNP":                     "/home/remi/projects/DNP",
		"/home/remi/projects/claude-notifications-go": "/home/remi/projects/claude-notifications-go",
	})

	tests := []struct {
		name        string
		cwd         string
		sessionName string
		want        string
	}{
		{
			name:        "project folder replaces the session mnemonic",
			cwd:         "/home/remi/projects/DNP",
			sessionName: "swift 274b9ac1",
			want:        "DNP",
		},
		{
			name:        "longer project name",
			cwd:         "/home/remi/projects/claude-notifications-go",
			sessionName: "happy 06ddb8f7",
			want:        "claude-notifications-go",
		},
		{
			name:        "folder name containing spaces survives",
			cwd:         "/home/remi/my project",
			sessionName: "happy 06ddb8f7",
			want:        "my project",
		},
		{
			// Without a cwd there is no project to name, so the mnemonic is kept
			// rather than leaving the title with no identifier at all.
			name:        "unknown cwd falls back to the session mnemonic",
			cwd:         "",
			sessionName: "happy 06ddb8f7",
			want:        "happy 06ddb8f7",
		},
		{
			name:        "filesystem root falls back to the session mnemonic",
			cwd:         "/",
			sessionName: "happy 06ddb8f7",
			want:        "happy 06ddb8f7",
		},
		{
			name:        "neither available yields no identifier",
			cwd:         "",
			sessionName: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, titleIdentity(tt.cwd, tt.sessionName))
		})
	}
}

func TestProjectName(t *testing.T) {
	// The git root wins wherever one is known; the remaining cases fall back to
	// the cwd folder because the directory is not inside a working tree.
	stubGitRoot(t, map[string]string{
		"/home/remi/projects/DNP/harness/tests":       "/home/remi/projects/DNP",
		"/home/remi/projects/DNP":                     "/home/remi/projects/DNP",
		"/home/remi/projects/claude-notifications-go": "/home/remi/projects/claude-notifications-go",
		"/home/remi/worktrees/feat-dnd":               "/home/remi/worktrees/feat-dnd",
	})

	tests := []struct {
		name string
		cwd  string
		want string
	}{
		{
			name: "subdirectory of a repo is named after the repo root",
			cwd:  "/home/remi/projects/DNP/harness/tests",
			want: "DNP",
		},
		{
			name: "repo root names itself",
			cwd:  "/home/remi/projects/DNP",
			want: "DNP",
		},
		{
			name: "longer repo name",
			cwd:  "/home/remi/projects/claude-notifications-go",
			want: "claude-notifications-go",
		},
		{
			// A linked worktree reports its own root, so two worktrees of one
			// repository stay distinguishable.
			name: "linked worktree keeps its own name",
			cwd:  "/home/remi/worktrees/feat-dnd",
			want: "feat-dnd",
		},
		{
			name: "outside any repo falls back to the folder name",
			cwd:  "/home/remi/projects",
			want: "projects",
		},
		{
			name: "folder name with spaces",
			cwd:  "/home/remi/my project",
			want: "my project",
		},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"filesystem root", "/", ""},
		{"current directory", ".", ""},
		{"parent directory", "..", ""},
		{"relative path", "projects/dotfiles", "dotfiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, projectName(tt.cwd))
		})
	}
}

// TestTitlePipeline_FromHookMessage runs legacyPresentation, which SendDesktop
// uses to build the title, starting from the message format hooks.sendNotifications
// produces, and pins the rendered result.
func TestTitlePipeline_FromHookMessage(t *testing.T) {
	stubGitRoot(t, map[string]string{
		"/home/remi/projects/DNP": "/home/remi/projects/DNP",
	})

	tests := []struct {
		name    string
		message string
		cwd     string
		want    string
	}{
		{
			name:    "project name replaces the mnemonic",
			message: "[swift 274b9ac1|main DNP] Task complete",
			cwd:     "/home/remi/projects/DNP",
			want:    "✅ Task Completed [DNP]",
		},
		{
			// The folder is read from cwd, so a name containing spaces survives -
			// the message prefix could not encode it unambiguously.
			name:    "folder name containing spaces",
			message: "[happy 06ddb8f7|main my project] Task complete",
			cwd:     "/home/remi/my project",
			want:    "✅ Task Completed [my project]",
		},
		{
			name:    "no cwd keeps the previous mnemonic behaviour",
			message: "[happy 06ddb8f7|main DNP] Task complete",
			cwd:     "",
			want:    "✅ Task Completed [happy 06ddb8f7]",
		},
		{
			name:    "no prefix and no cwd leaves the status title alone",
			message: "Task complete",
			cwd:     "",
			want:    "✅ Task Completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := legacyPresentation(analyzer.StatusTaskComplete, tt.message, "✅ Task Completed", tt.cwd, true)
			assert.Equal(t, tt.want, got.Title)
		})
	}
}
