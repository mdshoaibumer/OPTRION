package routing

import (
	"github.com/google/uuid"
)

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
	ChannelIDs []uuid.UUID
}

// RoutingTable holds all routing rules.
type RoutingTable struct {
	Rules []RoutingRule
}

func (rt *RoutingTable) GetChannelsForSeverity(severity string) []uuid.UUID {
	for _, rule := range rt.Rules {
		if rule.Severity == severity {
			return rule.ChannelIDs
		}
	}
	return nil
}
