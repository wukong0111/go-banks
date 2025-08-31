-- Seed data for bank_groups table
-- This creates test data for the bank groups used in testing

INSERT INTO bank_groups (group_id, name, description) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'Cajas Rurales', 'Grupo de cajas rurales españolas')
ON CONFLICT (group_id) DO NOTHING;