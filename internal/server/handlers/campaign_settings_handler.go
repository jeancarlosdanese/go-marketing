// File: /internal/server/handlers/campaign_settings_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// CampaignSettingsHandler define as operações do handler
type CampaignSettingsHandler interface {
	CreateSettingsHandler() http.HandlerFunc
	GetSettingsHandler() http.HandlerFunc
	UpdateSettingsHandler() http.HandlerFunc
	DeleteSettingsHandler() http.HandlerFunc
	GetLastSettingsHandler() http.HandlerFunc
}

type campaignSettingsHandler struct {
	log          *slog.Logger
	campaignRepo db.CampaignRepository
	settingsRepo db.CampaignSettingsRepository
}

// NewCampaignSettingsHandler cria um novo handler
func NewCampaignSettingsHandler(settingsRepo db.CampaignSettingsRepository, campaignRepo db.CampaignRepository) CampaignSettingsHandler {
	return &campaignSettingsHandler{
		log:          logger.GetLogger(),
		campaignRepo: campaignRepo,
		settingsRepo: settingsRepo,
	}
}

// ✅ Criar configurações de uma campanha
func (h *campaignSettingsHandler) CreateSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		var requestDTO dto.CampaignSettingsDTO

		// 📝 Decodificar JSON
		if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Verificar ID da campanha e
		if requestDTO.CampaignID != campaignID {
			h.log.Warn("ID da campanha inválido", "campaign_id", requestDTO.CampaignID)
			utils.SendError(w, http.StatusBadRequest, "ID da campanha inválido")
			return
		}

		// 🔍 Validar DTO
		if err := requestDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil {
			h.log.Error("Erro ao buscar campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar campanha")
			return
		}

		// 🔍 Verificar se a campanha pertence à conta autenticada
		if campaign == nil || campaign.AccountID != authAccount.ID {
			h.log.Warn("Campanha não encontrada", "campaign_id", campaignID)
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// 🚀 Criar configurações
		settings, err := h.settingsRepo.CreateSettings(r.Context(), requestDTO.ToModel())
		if err != nil {
			h.log.Error("Erro ao criar configurações da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao criar configurações")
			return
		}

		h.log.Info("Configurações criadas com sucesso", "campaign_id", settings.CampaignID)
		utils.SendSuccess(w, http.StatusCreated, settings)
	}
}

// ✅ Buscar configurações de uma campanha
func (h *campaignSettingsHandler) GetSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := r.Context().Value(middleware.AuthAccountKey).(*models.Account)

		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil {
			h.log.Error("Erro ao buscar campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar campanha")
			return
		}

		// 🔍 Verificar se a campanha pertence à conta autenticada
		if campaign == nil || campaign.AccountID != authAccount.ID {
			h.log.Warn("Campanha não encontrada", "campaign_id", campaignID)
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		settings, err := h.settingsRepo.GetSettingsByCampaignID(r.Context(), campaignID)
		if err != nil {
			h.log.Error("Erro ao buscar configurações da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar configurações")
			return
		}

		if settings == nil {
			utils.SendError(w, http.StatusNoContent, "Nenhuma configuração encontrada")
			return
		}

		h.log.Info("Configurações da campanha encontradas", "campaign_id", campaignID)
		utils.SendSuccess(w, http.StatusOK, settings)
	}
}

// ✅ Atualizar configurações de uma campanha
func (h *campaignSettingsHandler) UpdateSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		var requestDTO dto.CampaignSettingsDTO

		// 📝 Decodificar JSON
		if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Validar DTO
		if err := requestDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Get settings by campaign ID
		settings, err := h.settingsRepo.GetSettingsByCampaignID(r.Context(), campaignID)
		if err != nil {
			h.log.Error("Erro ao buscar configurações da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar configurações")
			return
		}

		if settings == nil {
			utils.SendError(w, http.StatusNotFound, "Nenhuma configuração encontrada")
			return
		}

		// 🔍 Verificar ID da campanha
		if requestDTO.CampaignID != campaignID {
			h.log.Warn("ID da campanha inválido", "campaign_id", requestDTO.CampaignID)
			utils.SendError(w, http.StatusBadRequest, "ID da campanha inválido")
			return
		}

		// 🚀 Atualizar configurações
		settingsUpdated, err := h.settingsRepo.UpdateSettings(r.Context(), requestDTO.ToModel())
		if err != nil {
			h.log.Error("Erro ao atualizar configurações da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar configurações")
			return
		}

		h.log.Info("Configurações da campanha atualizadas", "campaign_id", campaignID)
		utils.SendSuccess(w, http.StatusOK, settingsUpdated)
	}
}

// ✅ Excluir configurações de uma campanha
func (h *campaignSettingsHandler) DeleteSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		// 🚀 Excluir configurações
		err := h.settingsRepo.DeleteSettings(r.Context(), campaignID)
		if err != nil {
			h.log.Error("Erro ao excluir configurações da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao excluir configurações")
			return
		}

		h.log.Info("Configurações da campanha excluídas", "campaign_id", campaignID)
		utils.SendSuccess(w, http.StatusOK, "Configurações removidas com sucesso")
	}
}

// ✅ Buscar última configuração usada por uma conta
func (h *campaignSettingsHandler) GetLastSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		authAccount := r.Context().Value(middleware.AuthAccountKey).(*models.Account)

		h.log.Debug("Buscando última configuração", "account_id", authAccount.ID)

		settings, err := h.settingsRepo.GetLastSettings(r.Context(), authAccount.ID)
		if err != nil {
			h.log.Error("Erro ao buscar última configuração", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar última configuração")
			return
		}

		if settings == nil {
			utils.SendError(w, http.StatusNotFound, "Nenhuma configuração recente encontrada")
			return
		}

		// Clone settings to DTO
		settingsDTO := dto.CampaignSettingsDTO{
			CampaignID:           campaignID,
			Brand:                settings.Brand,
			Subject:              settings.Subject,
			Tone:                 settings.Tone,
			EmailFrom:            settings.EmailFrom,
			EmailReply:           settings.EmailReply,
			EmailFooter:          settings.EmailFooter,
			EmailInstructions:    settings.EmailInstructions,
			WhatsAppFrom:         settings.WhatsAppFrom,
			WhatsAppReply:        settings.WhatsAppReply,
			WhatsAppFooter:       settings.WhatsAppFooter,
			WhatsAppInstructions: settings.WhatsAppInstructions,
		}

		settingsCreated, err := h.settingsRepo.CreateSettings(r.Context(), settingsDTO.ToModel())
		if err != nil {
			h.log.Error("Erro ao criar configurações da campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao criar configurações")
			return
		}

		h.log.Info("Última configuração recuperada", "account_id", campaignID)
		utils.SendSuccess(w, http.StatusOK, settingsCreated)
	}
}
