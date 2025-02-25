-- File: /migrations/004_account_settings.sql

CREATE TABLE account_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
    openai_api_key VARCHAR(64), -- 🔥 Chaves de API geralmente têm tamanho fixo
    evolution_instance VARCHAR(50),
    aws_access_key_id VARCHAR(20), -- 🔥 AWS Keys têm tamanho fixo
    aws_secret_access_key VARCHAR(40), -- 🔥 AWS Secret tem um tamanho conhecido
    aws_region VARCHAR(20), -- 🔥 Usamos `us-east-1`, `eu-west-1`, etc.
    mail_from VARCHAR(150),
    mail_admin_to VARCHAR(150),
    UNIQUE(account_id)
);
