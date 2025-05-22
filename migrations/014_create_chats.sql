-- /migrations/014_create_chats.sql

-- 🔹 Chat definition per department (e.g. financial, support, sales)
CREATE TABLE public.chats (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  department VARCHAR(50) NOT NULL, -- financeiro, comercial, suporte
  title VARCHAR(150),
  instructions TEXT,
  phone_number VARCHAR(20),
  evolution_instance VARCHAR(50),
  webhook_url VARCHAR(255),
  status VARCHAR(20) DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo')),
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);


-- 🔹 Relationship between a contact and a specific chat
CREATE TABLE public.chat_contacts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
  status VARCHAR(20) DEFAULT 'aberto' CHECK (status IN ('aberto', 'pendente', 'fechado')),
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

-- 🔹 All messages exchanged in a conversation
CREATE TABLE public.chat_messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  chat_contact_id UUID NOT NULL REFERENCES chat_contacts(id) ON DELETE CASCADE,
  actor VARCHAR(20) NOT NULL CHECK (actor IN ('cliente', 'atendente', 'ai')),
  type VARCHAR(20) NOT NULL CHECK (type IN ('texto', 'audio', 'imagem', 'video', 'documento')),
  content TEXT,              -- AI-processed content (transcription, OCR, etc.)
  file_url TEXT,             -- Original file URL (if applicable)
  source_processed BOOLEAN DEFAULT FALSE, -- True if processed by AI
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now(),
  deleted_at TIMESTAMP
);
