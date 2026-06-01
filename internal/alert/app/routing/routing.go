package routing

// Severity levels
const (
	SeverityInfo     = "Info"
	SeverityWarning  = "Warning"
	SeverityMinor    = "Minor"
	SeverityMajor    = "Major"
	SeverityCritical = "Critical"
)

// RoutingRule defines how to route alerts based on severity.
type RoutingRule struct {
	Severity   string
	ChannelIDs []string
}

// RoutingTable holds all routing rules.
type RoutingTable struct {
	Rules []RoutingRule
}

func (rt *RoutingTable) GetChannelsForSeverity(severity string) []string {
	for _, rule := range rt.Rules {
		if rule.Severity == severity {
			return rule.ChannelIDs
		}
	}
	return nil
}
