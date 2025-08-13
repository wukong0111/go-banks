-- Seed data for banks table
-- This creates test banks from Spain, Italy, and Portugal

INSERT INTO banks (
    bank_id, 
    name, 
    bank_codes, 
    bic, 
    real_name, 
    api, 
    api_version, 
    aspsp, 
    product_code, 
    country, 
    bank_group_id, 
    logo_url, 
    documentation, 
    keywords, 
    attribute, 
    auth_type_choice_required
) VALUES 
-- Spanish banks
('BES2100', 'CaixaBank', '["2100", "ES-2100"]', 'CAIXESBBXXX', 'CaixaBank S.A.', 'caixabank-api', '1.0', 'CaixaBank', 'CAIXA-001', 'ES', '550e8400-e29b-41d4-a716-446655440000', 'https://example.com/caixabank-logo.png', 'CaixaBank API documentation', '{"type": "retail", "services": ["payments", "accounts"]}', '{"supports_psd2": true, "instant_payments": true}', false),

('BES3059', 'Caja Rural Central', '["3059", "ES-3059", "RURAL-3059"]', 'BCOEESMM059', 'Caja Rural Central S.C.C.', 'caja-rural-api', '1.0', 'Caja Rural Central', 'CRC-001', 'ES', '550e8400-e29b-41d4-a716-446655440000', 'https://example.com/caja-rural-logo.png', 'Caja Rural Central API documentation', '{"type": "rural", "services": ["payments", "accounts", "agricultural"]}', '{"supports_psd2": true, "instant_payments": false}', false),

-- Italian banks
('BIT0300', 'Intesa Sanpaolo', '["0300", "IT-0300", "ISP-0300"]', 'BCITITMM', 'Intesa Sanpaolo S.p.A.', 'intesa-api', '1.0', 'Intesa Sanpaolo', 'ISP-001', 'IT', '550e8400-e29b-41d4-a716-446655440000', 'https://example.com/intesa-logo.png', 'Intesa Sanpaolo API documentation', '{"type": "retail", "services": ["payments", "accounts", "corporate"]}', '{"supports_psd2": true, "instant_payments": true}', true),

('BIT0200', 'UniCredit', '["0200", "IT-0200", "UCR-0200"]', 'UNCRITMM', 'UniCredit S.p.A.', 'unicredit-api', '1.0', 'UniCredit', 'UCR-001', 'IT', '550e8400-e29b-41d4-a716-446655440000', 'https://example.com/unicredit-logo.png', 'UniCredit API documentation', '{"type": "retail", "services": ["payments", "accounts", "investments"]}', '{"supports_psd2": true, "instant_payments": true}', false),

-- Portuguese banks
('BPT0033', 'Millennium BCP', '["0033", "PT-0033", "BCP-0033"]', 'BMILPTPL', 'Banco Comercial Português S.A.', 'millennium-api', '1.0', 'Millennium BCP', 'BCP-001', 'PT', '550e8400-e29b-41d4-a716-446655440000', 'https://example.com/millennium-logo.png', 'Millennium BCP API documentation', '{"type": "retail", "services": ["payments", "accounts", "corporate"]}', '{"supports_psd2": true, "instant_payments": false}', false),

('BPT0010', 'Banco Santander Totta', '["0010", "PT-0010", "STT-0010"]', 'TOTAPTPL', 'Banco Santander Totta S.A.', 'santander-pt-api', '1.0', 'Santander Totta', 'STT-001', 'PT', '550e8400-e29b-41d4-a716-446655440000', 'https://example.com/santander-pt-logo.png', 'Santander Totta API documentation', '{"type": "retail", "services": ["payments", "accounts", "mortgages"]}', '{"supports_psd2": true, "instant_payments": true}', true)

ON CONFLICT (bank_id) DO NOTHING;