CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL, -- 🔥 Limitado a 100 caracteres
    email VARCHAR(150) UNIQUE NOT NULL, -- 🔥 Limitado a 150 caracteres, pois alguns e-mails são longos
    whatsapp VARCHAR(20) UNIQUE -- 🔥 Limitado a 20 caracteres, pois o número de telefone é curto
);

CREATE TABLE account_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    openai_api_key VARCHAR(64), -- 🔥 Chaves de API geralmente têm tamanho fixo
    evolution_api_url VARCHAR(255), -- 🔥 URLs podem ser longas, mas 255 é um bom limite
    evolution_api_key VARCHAR(64),
    evolution_instance VARCHAR(50),
    aws_access_key_id VARCHAR(20), -- 🔥 AWS Keys têm tamanho fixo
    aws_secret_access_key VARCHAR(40), -- 🔥 AWS Secret tem um tamanho conhecido
    aws_region VARCHAR(20), -- 🔥 Usamos `us-east-1`, `eu-west-1`, etc.
    mail_from VARCHAR(150),
    mail_admin_to VARCHAR(150),
    x_api_key VARCHAR(64),
    limit_day INT DEFAULT 100,
    UNIQUE(account_id)
);
