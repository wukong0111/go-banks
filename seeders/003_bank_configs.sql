-- Seed data for bank_environment_configs table
-- This creates sandbox and production configurations for each test bank

INSERT INTO bank_environment_configs (
    bank_id,
    environment,
    enabled,
    blocked,
    blocked_text,
    risky,
    risky_message,
    supports_instant_payments,
    instant_payments_activated,
    instant_payments_limit,
    ok_status_codes_simple_payment,
    ok_status_codes_instant_payment,
    ok_status_codes_periodic_payment,
    enabled_periodic_payment,
    frequency_periodic_payment,
    config_periodic_payment
) VALUES 
-- CaixaBank configurations
('BES2100', 'sandbox', true, false, null, false, null, true, true, 15000, '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly', '{"max_amount": 5000}'),
('BES2100', 'production', true, false, null, false, null, true, true, 15000, '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly', '{"max_amount": 5000}'),
('BES2100', 'test', true, false, null, false, null, true, false, 5000, '["ACSC", "ACCC"]', '["ACSC"]', '["ACSC", "ACCC"]', true, 'monthly', '{"max_amount": 1000}'),

-- Caja Rural Central configurations
('BES3059', 'sandbox', true, false, null, true, 'Limited testing environment', false, false, 0, '["ACSC", "ACCC"]', null, '["ACSC", "ACCC"]', true, 'monthly', '{"max_amount": 2000}'),
('BES3059', 'production', true, false, null, false, null, false, false, 0, '["ACSC", "ACCC"]', null, '["ACSC", "ACCC"]', true, 'monthly', '{"max_amount": 2000}'),
('BES3059', 'test', true, false, null, true, 'Test environment', false, false, 0, '["ACSC", "ACCC"]', null, '["ACSC"]', true, 'monthly', '{"max_amount": 500}'),

-- Intesa Sanpaolo configurations
('BIT0300', 'sandbox', true, false, null, false, null, true, true, 25000, '["ACSC", "ACCC", "ACTC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly,yearly', '{"max_amount": 10000}'),
('BIT0300', 'production', true, false, null, false, null, true, true, 25000, '["ACSC", "ACCC", "ACTC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly,yearly', '{"max_amount": 10000}'),
('BIT0300', 'test', true, false, null, false, null, true, false, 10000, '["ACSC", "ACCC"]', '["ACSC"]', '["ACSC"]', true, 'monthly', '{"max_amount": 2500}'),

-- UniCredit configurations
('BIT0200', 'sandbox', true, false, null, false, null, true, true, 20000, '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly', '{"max_amount": 7500}'),
('BIT0200', 'production', true, false, null, false, null, true, true, 20000, '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly', '{"max_amount": 7500}'),
('BIT0200', 'test', true, false, null, false, null, true, false, 8000, '["ACSC", "ACCC"]', '["ACSC"]', '["ACSC"]', true, 'monthly', '{"max_amount": 2000}'),

-- Millennium BCP configurations
('BPT0033', 'sandbox', true, false, null, true, 'Testing environment with limitations', false, false, 0, '["ACSC", "ACCC"]', null, '["ACSC", "ACCC"]', true, 'monthly', '{"max_amount": 3000}'),
('BPT0033', 'production', true, false, null, false, null, false, false, 0, '["ACSC", "ACCC"]', null, '["ACSC", "ACCC"]', true, 'monthly', '{"max_amount": 3000}'),
('BPT0033', 'test', true, false, null, true, 'Test environment', false, false, 0, '["ACSC", "ACCC"]', null, '["ACSC"]', true, 'monthly', '{"max_amount": 800}'),

-- Santander Totta configurations
('BPT0010', 'sandbox', true, false, null, false, null, true, true, 12000, '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly', '{"max_amount": 4000}'),
('BPT0010', 'production', true, false, null, false, null, true, true, 12000, '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', '["ACSC", "ACCC"]', true, 'monthly,quarterly', '{"max_amount": 4000}'),
('BPT0010', 'test', true, false, null, false, null, true, false, 6000, '["ACSC", "ACCC"]', '["ACSC"]', '["ACSC"]', true, 'monthly', '{"max_amount": 1200}')

ON CONFLICT (bank_id, environment) DO NOTHING;