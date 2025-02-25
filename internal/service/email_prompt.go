// File: /internal/service/email_prompt.go

package service

import (
	"encoding/json"
	"fmt"

	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// GenerateEmailPromptForAI gera um prompt dinâmico para a IA criar um e-mail personalizado.
func GenerateEmailPromptForAI(contact models.Contact, campaign models.Campaign, campaignSettings models.CampaignSettings) string {
	// 🔹 Criar um mapa com as informações do contato
	contactData := map[string]interface{}{
		"Nome":           contact.Name,
		"E-mail":         contact.Email,
		"WhatsApp":       contact.WhatsApp,
		"Gênero":         contact.Gender,
		"Histórico":      contact.History,
		"Último Contato": contact.LastContactAt,
		"Localização":    fmt.Sprintf("%s, %s - %s", contact.Bairro, contact.Cidade, contact.Estado),
		"Interesses":     contact.Tags,
	}

	// 🔹 Criar um mapa com as informações da campanha
	campaignData := map[string]interface{}{
		"Nome":       campaign.Name,
		"Descrição":  campaign.Description,
		"Marca":      campaignSettings.Brand,
		"Assunto":    campaignSettings.Subject,
		"Tom de Voz": campaignSettings.Tone,
		"Instruções": campaignSettings.EmailInstructions,
		"Rodapé":     campaignSettings.EmailFooter,
	}

	// 🔹 Converter os mapas para JSON formatado
	contactJSON, _ := json.MarshalIndent(contactData, "", "  ")
	campaignJSON, _ := json.MarshalIndent(campaignData, "", "  ")

	// 🔹 Criar o esquema esperado para o e-mail
	emailSchema := `
	O e-mail gerado deve conter:
	1. **Saudação**: Cumprimentar o destinatário de forma personalizada.
	2. **Corpo**: Contextualizar a campanha e destacar a oferta ou mensagem principal.
	3. **Finalização**: Incentivar a ação desejada, como responder ao e-mail ou acessar um link.
	4. **Assinatura**: Finalizar com a marca da empresa e informações de contato.`

	// 🔹 Construção do prompt final
	prompt := fmt.Sprintf(`
	📩 **Geração de E-mail Personalizado**

	📌 **Detalhes do Contato**:
	%s

	📌 **Detalhes da Campanha**:
	%s

	📌 **Formato Esperado**:
	%s

	💡 **Tarefa**
	- Gere um e-mail personalizado com base nas informações acima.
	- Utilize um tom coerente com a campanha.
	- Retorne o e-mail formatado exclusivamente como um JSON válido, contendo os seguintes campos:
		{
			"saudacao": "string",
			"corpo": "string",
			"finalizacao": "string",
			"assinatura": "string"
		}
	`, string(contactJSON), string(campaignJSON), emailSchema)

	return prompt
}
