CREATE TYPE environment_type AS ENUM ('sandbox', 'production', 'uat', 'test');

CREATE TABLE bank_environment_configs (
    bank_id VARCHAR(255) NOT NULL REFERENCES banks(bank_id) ON UPDATE CASCADE,
    environment environment_type NOT NULL,
    enabled SMALLINT NOT NULL DEFAULT 1,
    blocked BOOLEAN NOT NULL DEFAULT FALSE,
    blocked_text TEXT,
    risky BOOLEAN NOT NULL DEFAULT FALSE,
    risky_message TEXT,
    supports_instant_payments BOOLEAN DEFAULT FALSE,
    instant_payments_activated BOOLEAN DEFAULT FALSE,
    instant_payments_limit INTEGER DEFAULT 0,
    ok_status_codes_simple_payment JSONB,
    ok_status_codes_instant_payment JSONB,
    ok_status_codes_periodic_payment JSONB,
    enabled_periodic_payment BOOLEAN DEFAULT FALSE,
    frequency_periodic_payment VARCHAR(255),
    config_periodic_payment TEXT,
    app_auth_setup_required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (bank_id, environment)
);

CREATE INDEX idx_bank_env_configs_environment ON bank_environment_configs(environment);
CREATE INDEX idx_bank_env_configs_enabled ON bank_environment_configs(enabled);

-- JSONB indexes for better performance on status codes queries
CREATE INDEX idx_bank_env_configs_simple_payment_codes ON bank_environment_configs USING GIN(ok_status_codes_simple_payment);
CREATE INDEX idx_bank_env_configs_instant_payment_codes ON bank_environment_configs USING GIN(ok_status_codes_instant_payment);
CREATE INDEX idx_bank_env_configs_periodic_payment_codes ON bank_environment_configs USING GIN(ok_status_codes_periodic_payment);
