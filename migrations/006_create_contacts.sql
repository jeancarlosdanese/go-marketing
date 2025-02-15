-- Criação da tabela de contatos
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    whatsapp VARCHAR(20) UNIQUE NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('masculino', 'feminino', 'outro')) DEFAULT NULL, -- Gênero
    birth_date DATE DEFAULT NULL, -- Data de nascimento (opcional)
    tags JSONB DEFAULT '[]',
    history TEXT DEFAULT NULL, -- Pequena história do contato
    opt_out_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
