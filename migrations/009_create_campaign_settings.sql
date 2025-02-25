-- File: /migrations/009_create_campaign_settings.sql

CREATE TABLE campaign_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID UNIQUE NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,

    -- Campos padronizados
    brand VARCHAR(100) NOT NULL,              -- Nome da marca (compatível com contacts.name)
    subject VARCHAR(150) NOT NULL,            -- Assunto do e-mail (compatível com contacts.email)
    tone VARCHAR(20) DEFAULT NULL,            -- Tom de voz da campanha (20 caracteres é suficiente)
    
    email_from VARCHAR(150) NOT NULL,         -- E-mail do remetente (compatível com contacts.email)
    email_reply VARCHAR(150) NOT NULL,        -- E-mail de resposta (compatível com contacts.email)
    email_footer TEXT DEFAULT NULL,           -- Rodapé do e-mail
    email_instructions TEXT NOT NULL,         -- Instruções para e-mail

    whatsapp_from VARCHAR(20) NOT NULL,       -- WhatsApp do remetente (compatível com contacts.whatsapp)
    whatsapp_reply VARCHAR(20) NOT NULL,      -- WhatsApp de resposta (compatível com contacts.whatsapp)
    whatsapp_footer TEXT DEFAULT NULL,        -- Rodapé do WhatsApp
    whatsapp_instructions TEXT NOT NULL,      -- Instruções para WhatsApp

    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
