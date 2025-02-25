// File: /internal/server/handlers/ses_feedback_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// SESFeedbackHandler define a interface para o handler
type SESFeedbackHandler interface {
	HandleSESFeedback() http.HandlerFunc
	HandleSNSEvent() http.HandlerFunc
}

// sesFeedbackHandler estrutura que lida com feedback do SES
type sesFeedbackHandler struct {
	log          *slog.Logger
	audienceRepo db.CampaignAudienceRepository
}

// NewSESFeedbackHandler cria um novo handler para processar feedback do SES
func NewSESFeedbackHandler(audienceRepo db.CampaignAudienceRepository) SESFeedbackHandler {
	log := logger.GetLogger()
	return &sesFeedbackHandler{log: log, audienceRepo: audienceRepo}
}

// SNSMessage estrutura para decodificar eventos do SNS
type SNSMessage struct {
	Message string `json:"Message"`
}

// SESNotification estrutura do evento SES
type SESNotification struct {
	EventType string `json:"eventType"`
	Mail      struct {
		MessageID string `json:"messageId"`
	} `json:"mail"`
	Bounce struct {
		BounceType string `json:"bounceType,omitempty"`
	} `json:"bounce,omitempty"`
	Complaint struct {
		ComplaintFeedbackType string `json:"complaintFeedbackType,omitempty"`
	} `json:"complaint,omitempty"`
}

// HandleSESFeedback processa eventos do SNS com notificações do SES
func (h *sesFeedbackHandler) HandleSESFeedback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("🔔 Recebendo evento SNS do SES...")

		// Decodifica o evento recebido do Amazon SNS
		var snsMessage SNSMessage
		if err := json.NewDecoder(r.Body).Decode(&snsMessage); err != nil {
			h.log.Error("❌ Erro ao decodificar mensagem do SNS", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Bad Request")
			return
		}

		// Decodifica a mensagem JSON do SES dentro do SNS
		var sesEvent SESNotification
		if err := json.Unmarshal([]byte(snsMessage.Message), &sesEvent); err != nil {
			h.log.Error("❌ Erro ao decodificar evento do SES", "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		h.log.Info("📩 Evento SES recebido", "event", sesEvent.EventType, "message_id", sesEvent.Mail.MessageID)

		// Mapeia os eventos SES para status do banco
		status := ""
		feedback := ""

		switch sesEvent.EventType {
		case "Send":
			status = "enviado"
		case "Rendering Failure":
			status = "falha_renderizacao"
		case "Reject":
			status = "rejeitado"
		case "Delivery":
			status = "entregue"
		case "Bounce":
			status = "devolvido"
			feedback = sesEvent.Bounce.BounceType
		case "Complaint":
			status = "reclamado"
			feedback = sesEvent.Complaint.ComplaintFeedbackType
		case "DeliveryDelay":
			status = "atrasado"
		case "SubscriptionUpdate":
			status = "atualizou_assinatura"
		default:
			h.log.Warn("⚠️ Evento SES não tratado", "event", sesEvent.EventType)
			http.Error(w, "Evento não tratado", http.StatusOK)
			return
		}

		// Atualiza o status no banco de dados
		var feedbackPtr *string
		if feedback != "" {
			feedbackPtr = &feedback
		}

		err := h.audienceRepo.UpdateStatusByMessageID(r.Context(), sesEvent.Mail.MessageID, status, feedbackPtr)
		if err != nil {
			h.log.Error("❌ Erro ao atualizar status no banco", "error", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		h.log.Info("✅ Status atualizado com sucesso!", "message_id", sesEvent.Mail.MessageID, "status", status)
		w.WriteHeader(http.StatusOK)
	}
}

// HandleSNSEvent processa eventos do SNS com notificações do SES
func (h *sesFeedbackHandler) HandleSNSEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decodifica a mensagem do SNS
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.log.Error("❌ Erro ao decodificar mensagem do SNS", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Bad Request")
			return
		}

		// Verificar se a mensagem é uma solicitação de assinatura
		if subscribeURL, ok := body["SubscribeURL"].(string); ok {
			h.log.Info("🔔 Recebido evento de inscrição SNS", "subscribe_url", subscribeURL)

			// Acesse a URL para confirmar a assinatura automaticamente
			resp, err := http.Get(subscribeURL)
			if err != nil {
				h.log.Error("❌ Erro ao confirmar inscrição SNS", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			h.log.Info("✅ Inscrição no SNS confirmada com sucesso!")
			w.WriteHeader(http.StatusOK)
			return
		}

		h.log.Info("📩 Evento SNS recebido sem SubscribeURL", "body", body)

		// Responder OK sem tentar nada mais
		w.WriteHeader(http.StatusOK)
	}
}
