-- Criação da tabela de mensagens geradas
CREATE TABLE generated_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    email_text TEXT,
    whatsapp_text TEXT,
    ai_model VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
