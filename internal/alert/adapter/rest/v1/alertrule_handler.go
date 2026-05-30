package v1

import (
	"net/http"
)

// AlertRuleHandler handles HTTP requests for alert rules.
type AlertRuleHandler struct{}

func (h *AlertRuleHandler) GetAlertRules(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GET /api/v1/alert-rules
}

func (h *AlertRuleHandler) PostAlertRule(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement POST /api/v1/alert-rules
}

func (h *AlertRuleHandler) PatchAlertRule(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement PATCH /api/v1/alert-rules/{id}
}
