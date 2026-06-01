-- Alert Intelligence Platform Migrations (UP)

-- 1. escalation_policies
CREATE TABLE IF NOT EXISTS escalation_policies (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    steps JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- 2. alert_rules
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(32) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    conditions JSONB NOT NULL,
    channels UUID[] NOT NULL,
    escalation_policy_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,
    CONSTRAINT fk_escalation_policy FOREIGN KEY (escalation_policy_id) REFERENCES escalation_policies(id)
);

-- 3. alert_channels
CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    type VARCHAR(32) NOT NULL,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- 4. alerts
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    rule_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    severity VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    message TEXT,
    deliveries UUID[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_rule FOREIGN KEY (rule_id) REFERENCES alert_rules(id)
);

-- 5. alert_deliveries
CREATE TABLE IF NOT EXISTS alert_deliveries (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    alert_id UUID NOT NULL,
    channel_id UUID NOT NULL,
    status VARCHAR(32) NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    history JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_alert FOREIGN KEY (alert_id) REFERENCES alerts(id),
    CONSTRAINT fk_channel FOREIGN KEY (channel_id) REFERENCES alert_channels(id)
);

-- 6. notification_templates
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template TEXT NOT NULL,
    variables TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Indexes for scalability
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_id ON alert_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_id ON alerts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alert_channels_tenant_id ON alert_channels(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alert_deliveries_tenant_id ON alert_deliveries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_escalation_policies_tenant_id ON escalation_policies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notification_templates_tenant_id ON notification_templates(tenant_id);
