// File: /internal/server/handlers/template_handler.go

package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type TemplateHandle interface {
	CreateTemplateHandler() http.HandlerFunc
	GetAllTemplatesHandler() http.HandlerFunc
	GetTemplateHandler() http.HandlerFunc
	UpdateTemplateHandler() http.HandlerFunc
	DeleteTemplateHandler() http.HandlerFunc
	UploadTemplateFileHandler() http.HandlerFunc
}

type templateHandle struct {
	log          *slog.Logger
	templateRepo db.TemplateRepository
}

func NewTemplateHandle(templateRepo db.TemplateRepository) TemplateHandle {
	return &templateHandle{
		log:          logger.GetLogger(),
		templateRepo: templateRepo,
	}
}

// CreateTemplateHandler cria um novo template
func (h *templateHandle) CreateTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var templateDTO dto.TemplateCreateDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&templateDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// Validar DTO
		if err := templateDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Criar template no banco
		template := &models.Template{
			AccountID:   authAccount.ID,
			Name:        templateDTO.Name,
			Description: templateDTO.Description,
		}

		createdTemplate, err := h.templateRepo.Create(r.Context(), template)
		if err != nil {
			h.log.Error("Erro ao criar template", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao criar template")
			return
		}

		h.log.Info("Template criado com sucesso", "id", createdTemplate.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dto.NewTemplateResponseDTO(createdTemplate))
	}
}

// GetAllTemplatesHandler retorna todos os templates da conta autenticada
func (h *templateHandle) GetAllTemplatesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar filtros da query string (ex.: `?channel=email&name=promoção`)
		filters := map[string]string{}
		if channel := r.URL.Query().Get("channel"); channel != "" {
			filters["channel"] = channel
		}
		if name := r.URL.Query().Get("name"); name != "" {
			filters["name"] = name
		}
		if description := r.URL.Query().Get("description"); description != "" {
			filters["description"] = description
		}

		templates, err := h.templateRepo.GetByAccountID(r.Context(), authAccount.ID, filters)
		if err != nil {
			h.log.Error("Erro ao buscar templates", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar templates")
			return
		}

		var response []dto.TemplateResponseDTO
		for _, template := range templates {
			response = append(response, dto.NewTemplateResponseDTO(&template))
		}

		h.log.Info("Templates retornados com sucesso", "total", len(response), "filters", filters)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// GetTemplateHandler busca um template pelo ID
func (h *templateHandle) GetTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		templateID := r.PathValue("id")
		id, err := uuid.Parse(templateID)
		if err != nil {
			h.log.Warn("ID inválido", "id", templateID)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar template no banco
		template, err := h.templateRepo.GetByID(r.Context(), id)
		if err != nil {
			h.log.Error("Erro ao buscar template", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar template")
			return
		}
		if template == nil || template.AccountID != authAccount.ID {
			utils.SendError(w, http.StatusNotFound, "Template não encontrado")
			return
		}

		h.log.Info("Template recuperado com sucesso", "id", templateID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewTemplateResponseDTO(template))
	}
}

// UpdateTemplateHandler atualiza um template pelo ID
func (h *templateHandle) UpdateTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var templateDTO dto.TemplateUpdateDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&templateDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		templateID := r.PathValue("id")
		id, err := uuid.Parse(templateID)
		if err != nil {
			h.log.Warn("ID inválido", "id", templateID)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar template no banco
		existingTemplate, err := h.templateRepo.GetByID(r.Context(), id)
		if err != nil || existingTemplate == nil || existingTemplate.AccountID != authAccount.ID {
			utils.SendError(w, http.StatusNotFound, "Template não encontrado")
			return
		}

		// Validar DTO
		if err := templateDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Atualizar campos informados
		if templateDTO.Name != nil {
			existingTemplate.Name = *templateDTO.Name
		}
		if templateDTO.Description != nil {
			existingTemplate.Description = templateDTO.Description
		}

		// 🔄 Atualizar template no banco
		updatedTemplate, err := h.templateRepo.UpdateByID(r.Context(), id, existingTemplate)
		if err != nil {
			h.log.Error("Erro ao atualizar template", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar template")
			return
		}

		h.log.Info("Template atualizado com sucesso", "id", updatedTemplate.ID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewTemplateResponseDTO(updatedTemplate))
	}
}

// DeleteTemplateHandler deleta um template pelo ID
func (h *templateHandle) DeleteTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		id := r.PathValue("id")
		templateID, err := uuid.Parse(id)
		if err != nil {
			h.log.Warn("ID inválido", "id", templateID)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar template no banco
		existingTemplate, err := h.templateRepo.GetByID(r.Context(), templateID)
		if err != nil || existingTemplate == nil || existingTemplate.AccountID != authAccount.ID {
			utils.SendError(w, http.StatusNotFound, "Template não encontrado")
			return
		}

		// 🔄 Deletar template no banco
		if err := h.templateRepo.DeleteByID(r.Context(), templateID); err != nil {
			h.log.Error("Erro ao deletar template", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao deletar template")
			return
		}

		utils.DeleteTemplate(templateID.String(), "email")
		utils.DeleteTemplate(templateID.String(), "whatsapp")

		h.log.Info("Template deletado com sucesso", "id", templateID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *templateHandle) UploadTemplateFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		templateType := r.PathValue("type") // "email" ou "whatsapp"

		// Validar se o template existe
		templateID, err := uuid.Parse(id)
		if err != nil {
			h.log.Warn("ID inválido", "id", templateID)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Buscar template no banco
		existingTemplate, err := h.templateRepo.GetByID(r.Context(), templateID)
		if err != nil || existingTemplate == nil || existingTemplate.AccountID != authAccount.ID {
			utils.SendError(w, http.StatusNotFound, "Template não encontrado")
			return
		}

		// Validar tipo de template
		if templateType != "email" && templateType != "whatsapp" {
			utils.SendError(w, http.StatusBadRequest, "Tipo de template inválido")
			return
		}

		// Ler arquivo enviado
		file, _, err := r.FormFile("file")
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "Erro ao ler arquivo")
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao processar arquivo")
			return
		}

		// Salvar localmente e no S3
		if err := utils.SaveTemplate(templateID.String(), templateType, content); err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao salvar template")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Template salvo com sucesso"})
	}
}
