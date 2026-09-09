package collector

import (
	"regexp"
	"strings"
)

// EtcHostsEntry représente une ligne significative de /etc/hosts
type EtcHostsEntry struct {
	IP       string
	Hostname string
}

var hostsLineRegex = regexp.MustCompile(`^(\S+)\s+(\S+)`)

// parseEtcHosts extrait les entrées IP/hostname du contenu de /etc/hosts,
// en ignorant les commentaires et les lignes localhost.
func parseEtcHosts(content string) []EtcHostsEntry {
	var entries []EtcHostsEntry
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "localhost") {
			continue
		}
		matches := hostsLineRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			entries = append(entries, EtcHostsEntry{IP: matches[1], Hostname: matches[2]})
		}
	}
	return entries
}
