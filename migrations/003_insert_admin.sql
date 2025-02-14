INSERT INTO accounts (id, name, email, whatsapp)
VALUES ('00000000-0000-0000-0000-000000000001', 'Admin', 'admin@hyberica.io', '5549999669869')
ON CONFLICT (email) DO NOTHING;