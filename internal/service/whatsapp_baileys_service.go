// internal/service/whatsapp_baileys_service.go

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

type WhatsAppBaileysService interface {
	StartSession(sessionID, webhookURL string) (*StartSessionResponse, error)
	GetQRCode(sessionID string) (*QRCodeResponse, error)
	SendTextMessage(sessionID, number, message string) (*SendMessageResponse, error)
	GetSessionState(sessionID string) (*dto.SessionStatusDTO, error)
}

type whatsAppBaileysService struct {
	log    *slog.Logger
	apiURL string
	apiKey string
}

func NewWhatsAppBaileysService(apiURL, apiKey string) WhatsAppBaileysService {
	return &whatsAppBaileysService{
		log:    logger.GetLogger(),
		apiURL: apiURL,
		apiKey: apiKey,
	}
}

type StartSessionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type QRCodeResponse struct {
	QRCode string `json:"qrCode,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

type SendMessageRequest struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

type SendMessageResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StartSession inicia uma nova sessão do WhatsApp Baileys
func (s *whatsAppBaileysService) StartSession(sessionID, webhookURL string) (*StartSessionResponse, error) {
	s.log.Debug("Iniciando sessão do WhatsApp Baileys", slog.String("sessionID", sessionID), slog.String("webhookURL", webhookURL))
	url := fmt.Sprintf("%s/sessions/%s/start", s.apiURL, sessionID)

	s.log.Debug("URL da API", slog.String("url", url))

	body := map[string]string{"webhookUrl": webhookURL}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("erro ao montar payload JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	var res StartSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &res, fmt.Errorf("falha na API: %s", res.Error)
	}

	return &res, nil
}

// GetQRCode obtém o QR Code para autenticação da sessão do WhatsApp Baileys
func (s *whatsAppBaileysService) GetQRCode(sessionID string) (*QRCodeResponse, error) {
	url := fmt.Sprintf("%s/sessions/%s/qrcode", s.apiURL, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("X-API-Key", s.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	var res QRCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &res, fmt.Errorf("falha na API: %s", res.Error)
	}

	return &res, nil
}

// SendTextMessage envia uma mensagem de texto para um número específico
func (s *whatsAppBaileysService) SendTextMessage(sessionID, number, message string) (*SendMessageResponse, error) {
	url := fmt.Sprintf("%s/sessions/%s/send", s.apiURL, sessionID)

	payload := SendMessageRequest{
		Number: number,
		Text:   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	var res SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &res, fmt.Errorf("falha na API: %s", res.Error)
	}

	return &res, nil
}

// GetSessionState consulta o state da sessão e atualiza no banco
func (s *whatsAppBaileysService) GetSessionState(sessionID string) (*dto.SessionStatusDTO, error) {
	url := fmt.Sprintf("%s/sessions/%s/state", s.apiURL, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}
	req.Header.Set("X-API-Key", s.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("Erro ao ler o corpo da resposta", slog.Any("error", err))
		return nil, fmt.Errorf("erro ao ler corpo da resposta: %w", err)
	}
	s.log.Debug("BODY DA RESPOSTA", slog.Any("body", string(bodyBytes)))

	var sessionStatus *dto.SessionStatusDTO
	if err := json.Unmarshal(bodyBytes, &sessionStatus); err != nil {
		s.log.Error("Erro ao decodificar resposta JSON", slog.Any("error", err))
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %w", err)
	}

	return sessionStatus, nil
}
