// File: internal/service/message_renderer.go

package service

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/config"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
)

// RenderWithTemplateFile aplica os dados ao template salvo no disco
func RenderWithTemplateFile(content *dto.CampaignContentResult, templateID uuid.UUID, channel string) (string, error) {
	templateBasePath := config.GetEnvVar("TEMPLATE_STORAGE_PATH")
	var filename string
	switch channel {
	case "email":
		filename = filepath.Join(templateBasePath, "email", templateID.String()+".html")
	case "whatsapp":
		filename = filepath.Join(templateBasePath, "whatsapp", templateID.String()+".md")
	default:
		return "", fmt.Errorf("canal inválido: %s", channel)
	}

	templateBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("erro ao ler template: %w", err)
	}

	tpl, err := template.New("msg").Parse(string(templateBytes))
	if err != nil {
		return "", fmt.Errorf("erro ao parsear template: %w", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, content); err != nil {
		return "", fmt.Errorf("erro ao aplicar template: %w", err)
	}

	return buf.String(), nil
}

// RenderMessagePreview gera a mensagem final formatada com base no canal (email ou whatsapp)
func RenderMessagePreview(content *dto.CampaignContentResult, channel string) (string, error) {
	switch strings.ToLower(channel) {
	case "email":
		return renderHTMLEmail(content)
	case "whatsapp":
		return renderWhatsApp(content), nil
	default:
		return "", fmt.Errorf("canal não suportado: %s", channel)
	}
}

// renderHTMLEmail aplica HTML nos campos da mensagem
func renderHTMLEmail(content *dto.CampaignContentResult) (string, error) {
	tpl := `<p>{{.Saudacao}}</p><p>{{.Corpo}}</p><p>{{.Finalizacao}}</p><p>{{.Assinatura}}</p>`
	t, err := template.New("email").Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, content)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderWhatsApp aplica formatação simples estilo markdown para WhatsApp
func renderWhatsApp(content *dto.CampaignContentResult) string {
	return fmt.Sprintf("*%s*\n\n%s\n\n%s\n\n%s",
		content.Saudacao,
		content.Corpo,
		content.Finalizacao,
		content.Assinatura,
	)
}
