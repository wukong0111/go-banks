CREATE TABLE bank_groups (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    show_grouped SMALLINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_bank_groups_name ON bank_groups(name);