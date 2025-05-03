# ✅ Checklist – GoMarketing & NextMarketing (Atualizado em 03/05/2025)

## 🎯 Etapa 1: Gerar e revisar conteúdo da campanha com IA

- [x] Criar endpoint `POST /campaigns/:id/generate-message`
- [x] Gerar preview de e-mail com `emailService.CreateEmailWithAI()` + renderização no template HTML
- [ ] Gerar preview de WhatsApp (adaptar `whatsAppService.GenerateWhatsAppContent()`)
- [x] Criar DTO e estrutura para resposta com preview e mensagem bruta
- [ ] Salvar preview em `campaign_settings` ou `campaign_message_samples`
- [x] Exibir preview no frontend (aba “✨ Mensagens com IA”)

---

## 🎯 Etapa 2: Aprovar conteúdo e permitir envio

- [ ] Criar endpoint `POST /campaigns/:id/approve-preview`
- [ ] Marcar conteúdo como aprovado (flag em `campaign_settings` ou nova tabela)
- [ ] Bloquear botão "Iniciar Campanha" até aprovação

---

## 🎯 Etapa 3: Gerar e enviar mensagens personalizadas

- [ ] Atualizar workers de email e WhatsApp para gerar conteúdo com IA
- [ ] Para cada contato:
  - [ ] Usar `emailService.CreateEmailWithAI()` e preencher o template
  - [ ] Implementar `whatsAppService.CreateMessageWithAI()` semelhante ao email
- [ ] Enviar via SES (email) ou Evolution API (whatsapp)
- [ ] Registrar falhas/sucessos via `audienceRepo.UpdateStatus`

---

## 🧪 Testes finais

- [ ] Executar campanha de exemplo com 2 contatos e validar:
  - [x] Geração do preview com IA
  - [ ] Aprovação no frontend
  - [ ] Envio real via SQS
  - [ ] Worker processando e enviando corretamente
