// File: /internal/service/ai_prompt.go

package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jeancarlosdanese/go-marketing/internal/dto"
)

// GeneratePromptForAI gera um prompt dinâmico baseado em um único registro do CSV e nas configurações definidas pelo usuário.
func GeneratePromptForAI(record []string, headers []string, config *dto.ConfigImportContactDTO) string {
	// 🔹 Criar um mapa associando cabeçalhos aos valores do CSV
	dataMap := make(map[string]string)
	for i, value := range record {
		dataMap[headers[i]] = value
	}
	dataJSON, _ := json.MarshalIndent(dataMap, "", "  ") // Garante um JSON formatado corretamente

	// 🔹 Gerar instruções personalizadas para a IA com base na configuração
	fieldInstructions := generateFieldInstructions(config)

	// 🔹 Esquema do banco de dados esperado pela IA
	dbSchema := `
	O banco de dados espera a seguinte estrutura:
	- name (string) -> Nome do contato.
	- email (string, único) -> Endereço de e-mail do contato.
	- whatsapp (string, único) -> Número de telefone formatado para WhatsApp.
	- gender (string) -> Gênero do contato.
	- birth_date (date) -> Data de nascimento do contato (YYYY-MM-DD).
	- bairro (string) -> Bairro onde reside.
	- cidade (string) -> Cidade onde reside.
	- estado (string) -> Sigla do estado (UF).
	- tags (JSONB) -> Informações categorizadas, podendo incluir interesses, perfil e eventos.
	- history (text) -> Notas sobre interações anteriores.
	- opt_out_at (timestamp) -> Caso o contato tenha solicitado exclusão.
	`

	// 🔹 Construção do prompt com as instruções específicas
	prompt := fmt.Sprintf(`
		📌 **Registro de Contato Recebido**:
		%s

		📌 **Esquema do Banco de Dados**:
		%s

		💡 **Tarefa**
		1. Transforme os dados do contato no formato esperado pelo banco de dados.
		2. Siga as instruções abaixo para processar corretamente cada campo.
		3. Garanta que a saída seja um JSON puro, sem formatação extra.

		📌 **Instruções Específicas**:
		%s
	`, string(dataJSON), dbSchema, fieldInstructions)

	return prompt
}

// generateFieldInstructions gera as instruções para a IA com base na configuração do usuário.
func generateFieldInstructions(config *dto.ConfigImportContactDTO) string {
	var instructions []string

	// 🔹 Percorrer cada campo do DTO de configuração
	fieldMappings := map[string]dto.FieldMapping{
		"name":       config.Name,
		"email":      config.Email,
		"whatsapp":   config.WhatsApp,
		"gender":     config.Gender,
		"birth_date": config.BirthDate,
		"bairro":     config.Bairro,
		"cidade":     config.Cidade,
		"estado":     config.Estado,
		"eventos":    config.Eventos,
		"interesses": config.Interesses,
		"perfil":     config.Perfil,
		"history":    config.History,
	}

	for field, mapping := range fieldMappings {
		if mapping.Source != "" {
			instruction := fmt.Sprintf("- **%s**: Utilize o campo `%s` do CSV", field, mapping.Source)

			// 🔹 Adicionar regras personalizadas se existirem
			if len(mapping.Rules) > 0 {
				rules := []string{}
				for ruleKey, ruleValue := range mapping.Rules {
					rules = append(rules, fmt.Sprintf("%s: %s", ruleKey, ruleValue))
				}
				instruction += " e siga as regras: " + fmt.Sprintf("[%s]", strings.Join(rules, "; "))
			}

			instructions = append(instructions, instruction)
		}
	}

	return strings.Join(instructions, "\n")
}
