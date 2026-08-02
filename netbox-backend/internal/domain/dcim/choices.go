package dcim

// SiteStatus is the exact value stored and exposed by the pinned NetBox Site
// contract.
type SiteStatus string

const (
	SiteStatusPlanned         SiteStatus = "planned"
	SiteStatusStaging         SiteStatus = "staging"
	SiteStatusActive          SiteStatus = "active"
	SiteStatusDecommissioning SiteStatus = "decommissioning"
	SiteStatusRetired         SiteStatus = "retired"
)

func ParseSiteStatus(value string) (SiteStatus, bool) {
	status := SiteStatus(value)
	switch status {
	case SiteStatusPlanned,
		SiteStatusStaging,
		SiteStatusActive,
		SiteStatusDecommissioning,
		SiteStatusRetired:
		return status, true
	default:
		return "", false
	}
}

func (status SiteStatus) String() string {
	return string(status)
}
