package models

import "time"

type VMFamily string

const (
	FamilySM      VMFamily = "sm"
	FamilyCM      VMFamily = "cm"
	FamilyWS      VMFamily = "ws"
	FamilyOA      VMFamily = "oa"
	FamilyUnknown VMFamily = "unknown"
)

// DetectFamily déduit la famille depuis le hostname (mot-clé contenu dedans)
func DetectFamily(hostname string) VMFamily {
	lower := hostname
	for _, f := range []VMFamily{FamilySM, FamilyCM, FamilyWS, FamilyOA} {
		if contains(lower, string(f)) {
			return f
		}
	}
	return FamilyUnknown
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			ca, cb := s[i+j], substr[j]
			if 'A' <= ca && ca <= 'Z' {
				ca += 32
			}
			if 'A' <= cb && cb <= 'Z' {
				cb += 32
			}
			if ca != cb {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

type VM struct {
	ID          string            `json:"id"`
	IP          string            `json:"ip"`
	Hostname    string            `json:"hostname"`
	Family      VMFamily          `json:"family"`
	Hypervisor  string            `json:"hypervisor"`
	GroupID     string            `json:"groupId"`
	AppVersions map[string]string `json:"appVersions"`
	LastChecked time.Time         `json:"lastChecked"`
	Reachable   bool              `json:"reachable"`
	Error       string            `json:"error,omitempty"`
}

type Group struct {
	ID        string    `json:"id"`
	VMs       []VM      `json:"vms"`
	InUseBy   *string   `json:"inUseBy,omitempty"`
	CheckedAt *time.Time `json:"checkedAt,omitempty"`
}
