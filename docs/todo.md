# ✅ Checklist – GoMarketing & NextMarketing

## O QUE FALTA PARA CONCLUIR A META DE HOJE

### 🎯 Etapa 1: Gerar e revisar conteúdo da campanha com IA

- [ ] Criar endpoint `POST /campaigns/:id/generate-preview`
- [ ] Gerar preview de e-mail com `emailService.CreateEmailWithAI()`
- [ ] Gerar preview de WhatsApp (adaptar `whatsAppService.GenerateWhatsAppContent()`)
- [ ] Criar DTO e estrutura para resposta com ambos os previews
- [ ] Salvar preview em `campaign_settings` ou `campaign_message_samples`
- [ ] Exibir previews no frontend com botão "Aprovar conteúdo"

---

### 🎯 Etapa 2: Aprovar conteúdo e permitir envio

- [ ] Criar endpoint `POST /campaigns/:id/approve-preview`
- [ ] Marcar conteúdo como aprovado (flag em `campaign_settings` ou nova tabela)
- [ ] Bloquear botão "Iniciar Campanha" até aprovação

---

### 🎯 Etapa 3: Gerar e enviar mensagens personalizadas

- [ ] Atualizar workers de email e whatsapp para usar IA ao gerar conteúdo real
- [ ] Para cada contato:
  - [ ] Usar `emailService.CreateEmailWithAI()` e preencher o template
  - [ ] Implementar `whatsAppService.CreateMessageWithAI()` semelhante ao email
- [ ] Enviar via SES (email) ou Evolution API (whatsapp)
- [ ] Registrar falhas/sucessos no banco via `audienceRepo.UpdateStatus`

---

### 🧪 Testes finais

- [ ] Executar campanha de exemplo com 2 contatos e validar:
  - [ ] Geração do preview com IA
  - [ ] Aprovação no frontend
  - [ ] Envio real via SQS
  - [ ] Worker processando e enviando corretamente
