// File: internal/dto/import_config_dto.go

package dto

import "github.com/jeancarlosdanese/go-marketing/internal/models"

// ConvertToConfigImportContactDTO converte o modelo salvo para o DTO usado na importação
func ConvertToConfigImportContactDTO(model models.ContactImportConfig) *ConfigImportContactDTO {
	return &ConfigImportContactDTO{
		AboutData:     model.AboutData,
		Name:          model.Name,
		Email:         model.Email,
		WhatsApp:      model.WhatsApp,
		Gender:        model.Gender,
		BirthDate:     model.BirthDate,
		Bairro:        model.Bairro,
		Cidade:        model.Cidade,
		Estado:        model.Estado,
		Eventos:       model.Eventos,
		Interesses:    model.Interesses,
		Perfil:        model.Perfil,
		History:       model.History,
		LastContactAt: model.LastContactAt,
	}
}
