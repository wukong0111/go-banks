DROP TRIGGER IF EXISTS update_bank_environment_configs_updated_at ON bank_environment_configs;
DROP TRIGGER IF EXISTS update_banks_updated_at ON banks;
DROP TRIGGER IF EXISTS update_bank_groups_updated_at ON bank_groups;
DROP FUNCTION IF EXISTS update_updated_at_column();