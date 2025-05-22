// internal/service/evolution_service.go

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

//	type EvolutionService interface {
//		SendTextMessage(phone, message string) error
//	}
type EvolutionService interface {
	SendTextMessage(instanceID, phone, message string) error
}

type evolutionService struct {
	apiURL string
	apiKey string
}

func NewEvolutionService() EvolutionService {
	return &evolutionService{
		apiURL: os.Getenv("EVOLUTION_API_URL"),
		apiKey: os.Getenv("EVOLUTION_API_KEY"),
	}
}

func (s *evolutionService) SendTextMessage(instanceID, number, message string) error {
	url := fmt.Sprintf("%s/message/sendText/%s", s.apiURL, instanceID)

	payload := map[string]string{
		"number": number,  // <- campo obrigatório correto
		"text":   message, // <- campo obrigatório correto
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro Evolution API: %s", string(respBody))
	}

	return nil
}
