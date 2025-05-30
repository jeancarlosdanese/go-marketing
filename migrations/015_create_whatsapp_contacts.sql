-- File: migrations/015_create_whatsapp_contacts.sql

CREATE TABLE whatsapp_contacts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  account_id         UUID           NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  contact_id         UUID           NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  name               VARCHAR(140)   NOT NULL,           -- Nome visível no WhatsApp
  phone              VARCHAR(20)    NOT NULL,           -- Número no formato E.164 (ex: 554999912345)
  jid                VARCHAR(100)   NOT NULL,           -- JID completo do contato (ex: 554999912345@s.whatsapp.net)
  is_business        BOOLEAN        DEFAULT FALSE,      -- true se for uma conta comercial
  business_profile   JSONB,                             -- Dados opcionais (nome comercial, descrição, etc.)
  created_at         TIMESTAMPTZ    DEFAULT now(),
  updated_at         TIMESTAMPTZ    DEFAULT now(),

  CONSTRAINT unique_account_phone UNIQUE (account_id, phone),
  CONSTRAINT unique_account_jid   UNIQUE (account_id, jid)
);
