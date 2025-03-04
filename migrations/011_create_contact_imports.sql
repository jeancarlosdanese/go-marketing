-- File: /migrations/011_create_contact_imports.sql

CREATE TABLE contact_imports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    status VARCHAR(20) CHECK (status IN ('pendente', 'processando', 'concluído', 'erro')) DEFAULT 'pendente',
    config JSONB DEFAULT '{}'::jsonb,
    preview_data JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX ON contact_imports (account_id);
CREATE UNIQUE INDEX ON contact_imports (account_id, file_name);
