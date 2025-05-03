// internal/server/handlers/campaign_message_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type CampaignMessageHandler interface {
	GenerateCampaignMessageHandler() http.HandlerFunc
}

type campaignMessageHandler struct {
	log              *slog.Logger
	campaignRepo     db.CampaignRepository
	settingsRepo     db.CampaignSettingsRepository
	contactRepo      db.ContactRepository
	audienceRepo     db.CampaignAudienceRepository
	processorService service.CampaignProcessorService
	messageRepo      db.CampaignMessageRepository
}

func NewCampaignMessageHandler(
	campaignRepo db.CampaignRepository,
	settingsRepo db.CampaignSettingsRepository,
	contactRepo db.ContactRepository,
	audienceRepo db.CampaignAudienceRepository,
	messageRepo db.CampaignMessageRepository,
	processor service.CampaignProcessorService,
) CampaignMessageHandler {
	return &campaignMessageHandler{
		log:              logger.GetLogger(),
		campaignRepo:     campaignRepo,
		settingsRepo:     settingsRepo,
		contactRepo:      contactRepo,
		audienceRepo:     audienceRepo,
		processorService: processor,
		messageRepo:      messageRepo,
	}
}

// POST /campaigns/{campaign_id}/generate-message
func (h *campaignMessageHandler) GenerateCampaignMessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		var payload struct {
			ContactID uuid.UUID `json:"contact_id"`
			Channel   string    `json:"channel"` // "email" ou "whatsapp"
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			utils.SendError(w, http.StatusBadRequest, "JSON inválido")
			return
		}

		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil || campaign == nil {
			h.log.Error("Erro ao buscar campanha", "error", err)
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		settings, err := h.settingsRepo.GetSettingsByCampaignID(r.Context(), campaign.ID)
		if err != nil || settings == nil {
			h.log.Error("Erro nas configurações da campanha", "error", err)
			utils.SendError(w, http.StatusNotFound, "Configurações não encontradas")
			return
		}

		var contact *models.Contact
		if payload.ContactID != uuid.Nil {
			contact, err = h.contactRepo.GetByID(r.Context(), payload.ContactID)
			if err != nil || contact == nil {
				h.log.Error("Erro ao buscar contato", "error", err)
				utils.SendError(w, http.StatusNotFound, "Contato não encontrado")
				return
			}
		} else {
			contact, err = h.audienceRepo.GetRandomContact(r.Context(), campaign.ID, payload.Channel)
			if err != nil || contact == nil {
				h.log.Error("Erro ao buscar contato", "error", err)
				utils.SendError(w, http.StatusNotFound, "Contato não encontrado")
				return
			}
		}

		fullDTO := dto.ToCampaignMessageFullDTO(*account, *campaign, *settings, *contact, payload.Channel)

		result, prompt, err := h.processorService.GenerateCampaignContent(r.Context(), fullDTO)
		if err != nil {
			h.log.Error("Erro ao gerar conteúdo da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao gerar mensagem")
			return
		}

		// ✅ Renderização final
		rendered, err := service.RenderWithTemplateFile(result, campaign.Channels[payload.Channel].TemplateID, payload.Channel)
		if err != nil {
			h.log.Error("Erro ao renderizar template", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao renderizar preview")
			return
		}

		// 💾 Salvar rascunho
		dtoToSave := &dto.CampaignMessageCreateDTO{
			CampaignID:  campaignID,
			ContactID:   &fullDTO.ContactID,
			Channel:     payload.Channel,
			Saudacao:    result.Saudacao,
			Corpo:       result.Corpo,
			Finalizacao: result.Finalizacao,
			Assinatura:  result.Assinatura,
			PromptUsado: prompt,
			Version:     1,
			IsApproved:  false,
		}
		modelToSave := dtoToSave.ToModel()
		saved, err := h.messageRepo.Create(r.Context(), modelToSave)
		if err != nil {
			h.log.Error("Erro ao salvar mensagem", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao salvar mensagem")
			return
		}

		// 📨 Resposta
		response := map[string]interface{}{
			"message": map[string]interface{}{
				"id":           saved.ID,
				"channel":      payload.Channel,
				"saudacao":     result.Saudacao,
				"corpo":        result.Corpo,
				"finalizacao":  result.Finalizacao,
				"assinatura":   result.Assinatura,
				"prompt_usado": prompt,
			},
			"preview":      rendered,
			"preview_type": map[string]string{"email": "html", "whatsapp": "markdown"}[payload.Channel],
		}

		utils.SendSuccess(w, http.StatusCreated, response)
	}
}
