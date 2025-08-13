-- Seed data for bank_groups table
-- This creates test data for the bank groups used in testing

INSERT INTO bank_groups (id, name, show_grouped) VALUES 
('550e8400-e29b-41d4-a716-446655440000', 'Cajas Rurales', 1)
ON CONFLICT (id) DO NOTHING;