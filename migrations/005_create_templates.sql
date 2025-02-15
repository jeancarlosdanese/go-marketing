-- Criação da tabela de templates
CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('email', 'whatsapp', 'both')),
    content TEXT NOT NULL,
    variables JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP DEFAULT now()
);
