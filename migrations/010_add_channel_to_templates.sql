-- File: /migrations/010_add_channel_to_templates.sql

ALTER TABLE templates ADD COLUMN channel VARCHAR(20) NOT NULL DEFAULT 'email';
ALTER TABLE templates ADD CONSTRAINT check_channel CHECK (channel IN ('email', 'whatsapp'));
