package version

import (
	"os/exec"
	"strings"
)

var CommitHash = "unknown"

func init() {
	if CommitHash != "unknown" {
		return
	}
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		CommitHash = strings.TrimSpace(string(out))
	}
}
