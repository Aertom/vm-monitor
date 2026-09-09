package collector

import (
	"regexp"
	"strings"

	"github.com/Aertom/vm-monitor/backend/internal/models"
)

var versionDirRegex = regexp.MustCompile(`^(.+)_(\d+\.\d+\.\d+)$`)

// versionPathByFamily définit où chercher selon la famille
func versionPathByFamily(family models.VMFamily) string {
	switch family {
	case models.FamilySM:
		return "/opt"
	default: // cm, ws, oa
		return "/appli"
	}
}

// parseVersionDirs prend le résultat de `ls -1 <path>` et extrait nom+version
func parseVersionDirs(lsOutput string) map[string]string {
	versions := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(lsOutput), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := versionDirRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			appName := matches[1]
			version := matches[2]
			versions[appName] = version
		}
	}
	return versions
}
