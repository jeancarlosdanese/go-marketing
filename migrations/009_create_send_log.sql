-- Criação da tabela de histórico de envio
CREATE TABLE send_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('email', 'whatsapp')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('sent', 'failed', 'opened', 'clicked')),
    sent_at TIMESTAMP DEFAULT now()
);
