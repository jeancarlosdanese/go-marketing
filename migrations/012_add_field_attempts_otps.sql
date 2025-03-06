-- /migrations/011_create_contact_imports.sql

ALTER TABLE account_otps ADD COLUMN attempts INT DEFAULT 0;
