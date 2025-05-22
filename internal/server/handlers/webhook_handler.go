// internal/server/handlers/webhook_handler.go

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type WebhookHandler interface {
	Handle() http.HandlerFunc
}

type webhookHandler struct {
	log     *slog.Logger
	chatSvc service.ChatWhatsAppService
}

func NewWebhookHandler(chatSvc service.ChatWhatsAppService) WebhookHandler {
	return &webhookHandler{
		log:     logger.GetLogger(),
		chatSvc: chatSvc,
	}
}

func (h *webhookHandler) Handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rawBody bytes.Buffer

		// TeeReader permite ler e salvar simultaneamente
		tee := io.TeeReader(r.Body, &rawBody)
		defer r.Body.Close()

		var payload dto.WebhookPayload
		if err := json.NewDecoder(tee).Decode(&payload); err != nil {
			h.log.Error("❌ Payload inválido no webhook", slog.Any("erro", err))
			utils.SendError(w, 400, "Payload inválido")
			return
		}

		// Log do JSON cru
		h.log.Debug("🔍 Payload cru recebido", slog.String("body", rawBody.String()))

		numero := payload.Data.Key.RemoteJID
		msg := payload.Data.Message.Conversation
		msgType := payload.Data.MessageType
		instancia := payload.Instance

		h.log.Info("📩 Webhook recebido",
			slog.String("instancia", instancia),
			slog.String("numero", numero),
			slog.String("tipo", msgType),
			slog.String("mensagem", msg),
		)

		if msgType != "conversation" || msg == "" {
			h.log.Warn("Mensagem ignorada: tipo inválido ou vazia")
			utils.SendSuccess(w, 200, map[string]string{"status": "ignorada"})
			return
		}

		// 🔧 Processamento principal
		go func() {
			if err := h.chatSvc.ProcessarMensagemRecebida(context.Background(), instancia, numero, msg); err != nil {
				h.log.Error("Erro ao processar mensagem recebida", slog.Any("err", err))
			}
		}()

		utils.SendSuccess(w, 200, map[string]string{"status": "ok"})
	}
}
