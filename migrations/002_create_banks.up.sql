CREATE TABLE banks (
    bank_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    bank_codes JSONB NOT NULL,
    bic VARCHAR(255),
    real_name VARCHAR(255),
    api VARCHAR(255) NOT NULL,
    api_version VARCHAR(10) DEFAULT '-' NOT NULL,
    aspsp VARCHAR(255) NOT NULL,
    product_code VARCHAR(255),
    country CHAR(2) NOT NULL,
    bank_group_id UUID REFERENCES bank_groups(group_id),
    logo_url TEXT,
    documentation TEXT,
    keywords JSONB,
    attribute JSONB,
    auth_type_choice_required BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_banks_country ON banks(country);
CREATE INDEX idx_banks_bank_group_id ON banks(bank_group_id);
CREATE INDEX idx_banks_bank_codes ON banks USING GIN(bank_codes);
CREATE INDEX idx_banks_keywords ON banks USING GIN(keywords);
CREATE INDEX idx_banks_attribute ON banks USING GIN(attribute);