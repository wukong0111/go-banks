-- Drop JSONB indexes
DROP INDEX IF EXISTS idx_bank_env_configs_periodic_payment_codes;
DROP INDEX IF EXISTS idx_bank_env_configs_instant_payment_codes;
DROP INDEX IF EXISTS idx_bank_env_configs_simple_payment_codes;

-- Drop other indexes
DROP INDEX IF EXISTS idx_bank_env_configs_enabled;
DROP INDEX IF EXISTS idx_bank_env_configs_environment;

-- Drop table and type
DROP TABLE IF EXISTS bank_environment_configs;
DROP TYPE IF EXISTS environment_type;