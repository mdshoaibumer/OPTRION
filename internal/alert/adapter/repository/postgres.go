package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/alert/domain/alert"
	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
	"github.com/optrion/optrion/internal/alert/domain/alertdelivery"
	"github.com/optrion/optrion/internal/alert/domain/alertrule"
	"github.com/optrion/optrion/internal/alert/domain/escalationpolicy"
)

// AlertRulePostgresRepository implements AlertRuleRepository with PostgreSQL.
type AlertRulePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAlertRulePostgresRepository(pool *pgxpool.Pool) *AlertRulePostgresRepository {
	return &AlertRulePostgresRepository{pool: pool}
}

func (r *AlertRulePostgresRepository) Create(ctx context.Context, rule *alertrule.AlertRule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO alert_rules (id, tenant_id, name, description, severity, enabled, channels, escalation_policy_id, created_at, updated_at, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		rule.ID, rule.TenantID, rule.Name, rule.Description, rule.Severity,
		rule.Enabled, rule.Channels, rule.EscalationPolicyID,
		rule.CreatedAt, rule.UpdatedAt, rule.CreatedBy, rule.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting alert rule: %w", err)
	}
	return nil
}

func (r *AlertRulePostgresRepository) Update(ctx context.Context, rule *alertrule.AlertRule) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE alert_rules SET name=$1, description=$2, severity=$3, enabled=$4, channels=$5, escalation_policy_id=$6, updated_at=$7, updated_by=$8
		 WHERE id=$9 AND tenant_id=$10`,
		rule.Name, rule.Description, rule.Severity, rule.Enabled,
		rule.Channels, rule.EscalationPolicyID, time.Now().UTC(), rule.UpdatedBy,
		rule.ID, rule.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating alert rule: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert rule not found")
	}
	return nil
}

func (r *AlertRulePostgresRepository) FindByID(ctx context.Context, id string) (*alertrule.AlertRule, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, severity, enabled, channels, escalation_policy_id, created_at, updated_at, created_by, updated_by
		 FROM alert_rules WHERE id = $1`, id,
	)

	var rule alertrule.AlertRule
	err := row.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Description,
		&rule.Severity, &rule.Enabled, &rule.Channels, &rule.EscalationPolicyID,
		&rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy, &rule.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying alert rule: %w", err)
	}
	return &rule, nil
}

func (r *AlertRulePostgresRepository) ListByTenant(ctx context.Context, tenantID string) ([]*alertrule.AlertRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, severity, enabled, channels, escalation_policy_id, created_at, updated_at, created_by, updated_by
		 FROM alert_rules WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*alertrule.AlertRule
	for rows.Next() {
		var rule alertrule.AlertRule
		err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Description,
			&rule.Severity, &rule.Enabled, &rule.Channels, &rule.EscalationPolicyID,
			&rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy, &rule.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning alert rule: %w", err)
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func (r *AlertRulePostgresRepository) ListEnabledByTenant(ctx context.Context, tenantID string) ([]*alertrule.AlertRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, severity, enabled, channels, escalation_policy_id, created_at, updated_at, created_by, updated_by
		 FROM alert_rules WHERE tenant_id = $1 AND enabled = true ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing enabled alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*alertrule.AlertRule
	for rows.Next() {
		var rule alertrule.AlertRule
		err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Description,
			&rule.Severity, &rule.Enabled, &rule.Channels, &rule.EscalationPolicyID,
			&rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy, &rule.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning alert rule: %w", err)
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

// AlertPostgresRepository implements AlertRepository with PostgreSQL.
type AlertPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAlertPostgresRepository(pool *pgxpool.Pool) *AlertPostgresRepository {
	return &AlertPostgresRepository{pool: pool}
}

func (r *AlertPostgresRepository) Create(ctx context.Context, a *alert.Alert) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO alerts (id, tenant_id, rule_id, incident_id, severity, status, message, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		a.ID, a.TenantID, a.RuleID, a.IncidentID, a.Severity, a.Status, a.Message, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting alert: %w", err)
	}
	return nil
}

func (r *AlertPostgresRepository) Update(ctx context.Context, a *alert.Alert) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE alerts SET status=$1, updated_at=$2 WHERE id=$3 AND tenant_id=$4`,
		a.Status, time.Now().UTC(), a.ID, a.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating alert: %w", err)
	}
	return nil
}

func (r *AlertPostgresRepository) FindByID(ctx context.Context, id string) (*alert.Alert, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, rule_id, incident_id, severity, status, message, created_at, updated_at
		 FROM alerts WHERE id = $1`, id,
	)

	var a alert.Alert
	err := row.Scan(&a.ID, &a.TenantID, &a.RuleID, &a.IncidentID, &a.Severity, &a.Status, &a.Message, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying alert: %w", err)
	}
	return &a, nil
}

func (r *AlertPostgresRepository) ListByTenant(ctx context.Context, tenantID string) ([]*alert.Alert, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, rule_id, incident_id, severity, status, message, created_at, updated_at
		 FROM alerts WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*alert.Alert
	for rows.Next() {
		var a alert.Alert
		err := rows.Scan(&a.ID, &a.TenantID, &a.RuleID, &a.IncidentID, &a.Severity, &a.Status, &a.Message, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

// AlertChannelPostgresRepository implements AlertChannelRepository with PostgreSQL.
type AlertChannelPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAlertChannelPostgresRepository(pool *pgxpool.Pool) *AlertChannelPostgresRepository {
	return &AlertChannelPostgresRepository{pool: pool}
}

func (r *AlertChannelPostgresRepository) Create(ctx context.Context, c *alertchannel.AlertChannel) error {
	configJSON, err := encodeConfig(c.Config)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO alert_channels (id, tenant_id, type, name, config, enabled, created_at, updated_at, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.ID, c.TenantID, c.Type, c.Name, configJSON, c.Enabled, c.CreatedAt, c.UpdatedAt, c.CreatedBy, c.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting alert channel: %w", err)
	}
	return nil
}

func (r *AlertChannelPostgresRepository) Update(ctx context.Context, c *alertchannel.AlertChannel) error {
	configJSON, err := encodeConfig(c.Config)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE alert_channels SET name=$1, config=$2, enabled=$3, updated_at=$4, updated_by=$5
		 WHERE id=$6 AND tenant_id=$7`,
		c.Name, configJSON, c.Enabled, time.Now().UTC(), c.UpdatedBy, c.ID, c.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating alert channel: %w", err)
	}
	return nil
}

func (r *AlertChannelPostgresRepository) FindByID(ctx context.Context, id string) (*alertchannel.AlertChannel, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, type, name, config, enabled, created_at, updated_at, created_by, updated_by
		 FROM alert_channels WHERE id = $1`, id,
	)

	var c alertchannel.AlertChannel
	var configJSON []byte
	err := row.Scan(&c.ID, &c.TenantID, &c.Type, &c.Name, &configJSON, &c.Enabled, &c.CreatedAt, &c.UpdatedAt, &c.CreatedBy, &c.UpdatedBy)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying alert channel: %w", err)
	}
	c.Config = decodeConfig(configJSON)
	return &c, nil
}

func (r *AlertChannelPostgresRepository) ListByTenant(ctx context.Context, tenantID string) ([]*alertchannel.AlertChannel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, type, name, config, enabled, created_at, updated_at, created_by, updated_by
		 FROM alert_channels WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing alert channels: %w", err)
	}
	defer rows.Close()

	var channels []*alertchannel.AlertChannel
	for rows.Next() {
		var c alertchannel.AlertChannel
		var configJSON []byte
		err := rows.Scan(&c.ID, &c.TenantID, &c.Type, &c.Name, &configJSON, &c.Enabled, &c.CreatedAt, &c.UpdatedAt, &c.CreatedBy, &c.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("scanning alert channel: %w", err)
		}
		c.Config = decodeConfig(configJSON)
		channels = append(channels, &c)
	}
	return channels, nil
}

// AlertDeliveryPostgresRepository implements AlertDeliveryRepository with PostgreSQL.
type AlertDeliveryPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAlertDeliveryPostgresRepository(pool *pgxpool.Pool) *AlertDeliveryPostgresRepository {
	return &AlertDeliveryPostgresRepository{pool: pool}
}

func (r *AlertDeliveryPostgresRepository) Create(ctx context.Context, d *alertdelivery.AlertDelivery) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO alert_deliveries (id, tenant_id, alert_id, channel_id, status, attempts, last_error, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.ID, d.TenantID, d.AlertID, d.ChannelID, d.Status, d.Attempts, d.LastError, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting alert delivery: %w", err)
	}
	return nil
}

func (r *AlertDeliveryPostgresRepository) Update(ctx context.Context, d *alertdelivery.AlertDelivery) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE alert_deliveries SET status=$1, attempts=$2, last_error=$3, updated_at=$4
		 WHERE id=$5 AND tenant_id=$6`,
		d.Status, d.Attempts, d.LastError, time.Now().UTC(), d.ID, d.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating alert delivery: %w", err)
	}
	return nil
}

func (r *AlertDeliveryPostgresRepository) FindByID(ctx context.Context, id string) (*alertdelivery.AlertDelivery, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, alert_id, channel_id, status, attempts, last_error, created_at, updated_at
		 FROM alert_deliveries WHERE id = $1`, id,
	)

	var d alertdelivery.AlertDelivery
	err := row.Scan(&d.ID, &d.TenantID, &d.AlertID, &d.ChannelID, &d.Status, &d.Attempts, &d.LastError, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying alert delivery: %w", err)
	}
	return &d, nil
}

func (r *AlertDeliveryPostgresRepository) ListByAlert(ctx context.Context, alertID string) ([]*alertdelivery.AlertDelivery, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, alert_id, channel_id, status, attempts, last_error, created_at, updated_at
		 FROM alert_deliveries WHERE alert_id = $1 ORDER BY created_at DESC`, alertID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing alert deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*alertdelivery.AlertDelivery
	for rows.Next() {
		var d alertdelivery.AlertDelivery
		err := rows.Scan(&d.ID, &d.TenantID, &d.AlertID, &d.ChannelID, &d.Status, &d.Attempts, &d.LastError, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning alert delivery: %w", err)
		}
		deliveries = append(deliveries, &d)
	}
	return deliveries, nil
}

// Helper functions for config JSON encoding
func encodeConfig(config map[string]string) ([]byte, error) {
	if config == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(config)
}

func decodeConfig(data []byte) map[string]string {
	if len(data) == 0 {
		return make(map[string]string)
	}
	result := make(map[string]string)
	_ = json.Unmarshal(data, &result)
	return result
}

// EscalationPolicyPostgresRepository implements EscalationPolicyRepository with PostgreSQL.
type EscalationPolicyPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEscalationPolicyPostgresRepository(pool *pgxpool.Pool) *EscalationPolicyPostgresRepository {
	return &EscalationPolicyPostgresRepository{pool: pool}
}

func (r *EscalationPolicyPostgresRepository) Create(ctx context.Context, p *escalationpolicy.EscalationPolicy) error {
	stepsJSON, err := json.Marshal(p.Steps)
	if err != nil {
		return fmt.Errorf("marshaling escalation steps: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO escalation_policies (id, tenant_id, name, description, steps, created_at, updated_at, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.TenantID, p.Name, p.Description, stepsJSON, p.CreatedAt, p.UpdatedAt, p.CreatedBy, p.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting escalation policy: %w", err)
	}
	return nil
}

func (r *EscalationPolicyPostgresRepository) Update(ctx context.Context, p *escalationpolicy.EscalationPolicy) error {
	stepsJSON, err := json.Marshal(p.Steps)
	if err != nil {
		return fmt.Errorf("marshaling escalation steps: %w", err)
	}
	result, err := r.pool.Exec(ctx,
		`UPDATE escalation_policies SET name=$1, description=$2, steps=$3, updated_at=$4, updated_by=$5
		 WHERE id=$6 AND tenant_id=$7`,
		p.Name, p.Description, stepsJSON, time.Now().UTC(), p.UpdatedBy, p.ID, p.TenantID,
	)
	if err != nil {
		return fmt.Errorf("updating escalation policy: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("escalation policy not found")
	}
	return nil
}

func (r *EscalationPolicyPostgresRepository) FindByID(ctx context.Context, id string) (*escalationpolicy.EscalationPolicy, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, description, steps, created_at, updated_at, created_by, updated_by
		 FROM escalation_policies WHERE id = $1`, id,
	)

	var p escalationpolicy.EscalationPolicy
	var stepsJSON []byte
	err := row.Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &stepsJSON, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying escalation policy: %w", err)
	}
	if err := json.Unmarshal(stepsJSON, &p.Steps); err != nil {
		return nil, fmt.Errorf("unmarshaling escalation steps: %w", err)
	}
	return &p, nil
}

func (r *EscalationPolicyPostgresRepository) ListByTenant(ctx context.Context, tenantID string) ([]*escalationpolicy.EscalationPolicy, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, name, description, steps, created_at, updated_at, created_by, updated_by
		 FROM escalation_policies WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing escalation policies: %w", err)
	}
	defer rows.Close()

	var policies []*escalationpolicy.EscalationPolicy
	for rows.Next() {
		var p escalationpolicy.EscalationPolicy
		var stepsJSON []byte
		err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &stepsJSON, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy)
		if err != nil {
			return nil, fmt.Errorf("scanning escalation policy: %w", err)
		}
		if err := json.Unmarshal(stepsJSON, &p.Steps); err != nil {
			return nil, fmt.Errorf("unmarshaling escalation steps: %w", err)
		}
		policies = append(policies, &p)
	}
	return policies, nil
}
