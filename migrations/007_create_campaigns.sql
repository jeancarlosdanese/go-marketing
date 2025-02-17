CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT NULL,
    channels JSONB NOT NULL, -- Define os canais e templates usados (email, whatsapp, etc.)
    filters JSONB NOT NULL, -- Filtros da audiência (tags, gênero, etc.)
    status VARCHAR(20) NOT NULL CHECK (status IN ('pendente', 'ativa', 'concluida')),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
