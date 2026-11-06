package notifier

import (
	"path/filepath"
	"strings"

	"github.com/777genius/agent-notifications/internal/platform"
)

// getGitRoot is a seam so tests can exercise projectName without a real
// repository on disk.
var getGitRoot = platform.GetGitRoot

// titleIdentity returns the label shown in brackets after the status in the
// notification title, e.g. the "DNP" in "✅ Task Completed [DNP]".
//
// It names the project, which is what tells the reader which session a banner
// came from. It replaces the session mnemonic ("swift 274b9ac1"), which was
// derived from the session UUID and so named nothing recognisable. The mnemonic
// stays as the fallback for when cwd is unknown, so the title never loses its
// identifier entirely.
func titleIdentity(cwd, sessionName string) string {
	if project := projectName(cwd); project != "" {
		return project
	}
	return sessionName
}

// projectName returns the name identifying the project: the git working tree
// root when cwd is inside one, and the cwd folder itself otherwise.
//
// The repository root is preferred because a session started in a subdirectory
// would otherwise be labelled with a generic leaf name - a session in
// .../DNP/harness/tests reads as "DNP" rather than "tests".
func projectName(cwd string) string {
	if strings.TrimSpace(cwd) == "" {
		return ""
	}

	if name := baseName(getGitRoot(cwd)); name != "" {
		return name
	}
	return baseName(cwd)
}

// baseName returns the final element of path, or "" when the path names no
// folder - it is empty, a filesystem root, "." or "..".
func baseName(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}

	// filepath.Clean normalises separators first, so comparing against the
	// platform separator also covers "/" and "C:\\" on Windows.
	base := filepath.Base(filepath.Clean(path))
	switch base {
	case ".", "..", string(filepath.Separator):
		return ""
	}
	return base
}
