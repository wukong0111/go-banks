CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_bank_groups_updated_at 
    BEFORE UPDATE ON bank_groups 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_banks_updated_at 
    BEFORE UPDATE ON banks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bank_environment_configs_updated_at 
    BEFORE UPDATE ON bank_environment_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();