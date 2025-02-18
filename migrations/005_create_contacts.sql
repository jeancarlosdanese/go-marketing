CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE DEFAULT NULL,
    whatsapp VARCHAR(20) UNIQUE DEFAULT NULL,
    gender VARCHAR(10) CHECK (gender IN ('masculino', 'feminino', 'outro')) DEFAULT NULL,
    birth_date DATE DEFAULT NULL,
    bairro VARCHAR(100) DEFAULT NULL,
    cidade VARCHAR(100) DEFAULT NULL,
    estado VARCHAR(50) DEFAULT NULL,
    tags JSONB DEFAULT '{}'::jsonb, -- JSONB estruturado (Interesses, Perfil, Eventos)
    history TEXT DEFAULT NULL,
    opt_out_at TIMESTAMP NULL,
    last_contact_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
