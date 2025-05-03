-- /migrations/013_create_campaign_messages.sql
CREATE TABLE campaign_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  contact_id UUID REFERENCES contacts(id), -- opcional, p/ personalização futura
  channel TEXT NOT NULL CHECK (channel IN ('email', 'whatsapp')),
  saudacao TEXT,
  corpo TEXT,
  finalizacao TEXT,
  assinatura TEXT,
  prompt_usado TEXT,
  feedback JSONB, -- ex: ["Ficou muito genérico", "Adicionar nome da cidade"]
  version INT DEFAULT 1,
  is_approved BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);
