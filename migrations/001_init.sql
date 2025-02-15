CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL, -- 🔥 Limitado a 100 caracteres
    email VARCHAR(150) UNIQUE NOT NULL, -- 🔥 Limitado a 150 caracteres, pois alguns e-mails são longos
    whatsapp VARCHAR(20) UNIQUE -- 🔥 Limitado a 20 caracteres, pois o número de telefone é curto
);
