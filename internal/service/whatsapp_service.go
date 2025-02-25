// File: /internal/service/whatsapp_service.go

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// WhatsAppService gerencia o envio de mensagens via Evolution API
type WhatsAppService struct {
	log        *slog.Logger
	apiURL     string
	apiKey     string
	instanceID string
}

// NewWhatsAppService inicializa o serviço de WhatsApp
func NewWhatsAppService(apiURL, apiKey, instanceID string) *WhatsAppService {
	return &WhatsAppService{
		log:        slog.Default(),
		apiURL:     apiURL,
		apiKey:     apiKey,
		instanceID: instanceID,
	}
}

// SendWhatsApp envia uma mensagem via Evolution API
func (w *WhatsAppService) SendWhatsApp(whatsappRequest models.WhatsAppRequest) error {
	w.log.Info("📨 Enviando mensagem de WhatsApp via Evolution API", "to", whatsappRequest.To)

	// Criar payload JSON
	payload := map[string]interface{}{
		"to":         whatsappRequest.To,
		"templateId": whatsappRequest.TemplateID,
		"variables":  whatsappRequest.Variables,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		w.log.Error("❌ Erro ao serializar payload de WhatsApp", "error", err)
		return err
	}

	// Criar requisição HTTP
	url := fmt.Sprintf("%s/sendMessage?instance=%s", w.apiURL, w.instanceID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		w.log.Error("❌ Erro ao criar requisição de WhatsApp", "error", err)
		return err
	}

	// Adicionar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.apiKey)

	// 🚀 Enviar requisição para Evolution API
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.log.Error("❌ Erro ao enviar mensagem de WhatsApp", "error", err)
		return err
	}
	defer resp.Body.Close()

	// Verificar resposta da API
	if resp.StatusCode != http.StatusOK {
		w.log.Error("❌ Erro na resposta da Evolution API", "status", resp.StatusCode)
		return fmt.Errorf("erro na resposta da API: %d", resp.StatusCode)
	}

	w.log.Info("✅ Mensagem de WhatsApp enviada com sucesso!", "to", whatsappRequest.To)
	return nil
}

// GenerateWhatsAppContent cria variáveis para a mensagem do WhatsApp
func (w *WhatsAppService) GenerateWhatsAppContent(msg dto.CampaignMessageDTO) (map[string]string, error) {
	w.log.Info("Gerando conteúdo de WhatsApp", "contact_id", msg.ContactID)

	// // Personalizar nome do contato
	// name := msg.Name
	// if name == "" {
	// 	name = "Caro cliente"
	// }

	// Criar variáveis para o template do WhatsApp
	variables := map[string]string{
		// "name":   name,
		"offer":  "50% de desconto para você!", // Exemplo de variável
		"action": "Clique aqui para saber mais",
	}

	return variables, nil
}
