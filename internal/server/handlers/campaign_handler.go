// File: /internal/server/handlers/campaign_handler.go

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
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type CampaignHandle interface {
	CreateCampaignHandler() http.HandlerFunc
	GetCampaignHandler() http.HandlerFunc
	GetAllCampaignsHandler() http.HandlerFunc
	UpdateCampaignHandler() http.HandlerFunc
	UpdateCampaignStatusHandler() http.HandlerFunc
	DeleteCampaignHandler() http.HandlerFunc
}

type campaignHandle struct {
	log  *slog.Logger
	repo db.CampaignRepository
}

func NewCampaignHandle(repo db.CampaignRepository) CampaignHandle {
	return &campaignHandle{
		log:  logger.GetLogger(),
		repo: repo,
	}
}

// CreateCampaignHandler cria uma nova campanha
func (h *campaignHandle) CreateCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var campaignDTO dto.CampaignCreateDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&campaignDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// Validar DTO
		if err := campaignDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Criar campanha no banco
		campaign := &models.Campaign{
			AccountID:   authAccount.ID,
			Name:        campaignDTO.Name,
			Description: campaignDTO.Description,
			Channels:    campaignDTO.Channels,
			Filters:     campaignDTO.Filters,
			Status:      "pendente",
		}

		createdCampaign, err := h.repo.Create(campaign)
		if err != nil {
			h.log.Error("Erro ao criar campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao criar campanha")
			return
		}

		h.log.Info("Campanha criada com sucesso", "campaign_id", createdCampaign.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dto.NewCampaignResponseDTO(createdCampaign))
	}
}

// GetCampaignHandler busca uma campanha específica
func (h *campaignHandle) GetCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar `campaign_id` da URL
		campaignIDParam := r.PathValue("campaign_id")
		campaignID, err := uuid.Parse(campaignIDParam)
		if err != nil {
			h.log.Warn("ID inválido informado", "campaign_id", campaignIDParam)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar campanha no banco
		campaign, err := h.repo.GetByID(campaignID)
		if err != nil {
			h.log.Error("Erro ao buscar campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar campanha")
			return
		}
		if campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// ⚠️ Usuário só pode acessar suas próprias campanhas
		if !authAccount.IsAdmin() && authAccount.ID != campaign.AccountID {
			h.log.Warn("Usuário tentou acessar campanha de outra conta", "user_id", authAccount.ID, "requested_id", campaign.AccountID)
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		h.log.Info("Campanha encontrada com sucesso", "campaign_id", campaign.ID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewCampaignResponseDTO(campaign))
	}
}

// GetAllCampaignsHandler lista todas as campanhas da conta autenticada
func (h *campaignHandle) GetAllCampaignsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Buscar todas as campanhas da conta autenticada
		campaigns, err := h.repo.GetAllByAccountID(authAccount.ID)
		if err != nil {
			h.log.Error("Erro ao buscar campanhas", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar campanhas")
			return
		}

		// Montar resposta
		var response []dto.CampaignResponseDTO
		for _, campaign := range campaigns {
			response = append(response, dto.NewCampaignResponseDTO(&campaign))
		}

		h.log.Info("Campanhas recuperadas com sucesso", "total", len(response))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// UpdateCampaignHandler atualiza uma campanha
func (h *campaignHandle) UpdateCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updateDTO dto.CampaignUpdateDTO

		// Decodificar JSON
		if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar `campaign_id` da URL
		campaignIDParam := r.PathValue("campaign_id")
		campaignID, err := uuid.Parse(campaignIDParam)
		if err != nil {
			h.log.Warn("ID inválido informado", "campaign_id", campaignIDParam)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar campanha
		campaign, err := h.repo.GetByID(campaignID)
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// ⚠️ Usuário só pode modificar suas próprias campanhas
		if !authAccount.IsAdmin() && authAccount.ID != campaign.AccountID {
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		// Atualizar campanha
		if updateDTO.Name != nil {
			campaign.Name = *updateDTO.Name
		}
		if updateDTO.Description != nil {
			campaign.Description = updateDTO.Description
		}
		if updateDTO.Channels != nil {
			campaign.Channels = *updateDTO.Channels
		}
		if updateDTO.Filters != nil {
			campaign.Filters = *updateDTO.Filters
		}

		// Validar updateDTO
		if err := updateDTO.Validate(); err != nil {
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Atualizar campanha no banco
		updatedCampaign, err := h.repo.UpdateByID(campaignID, campaign)
		if err != nil {
			h.log.Error("Erro ao atualizar campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar campanha")
			return
		}

		h.log.Info("Campanha atualizada com sucesso", "campaign_id", updatedCampaign.ID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewCampaignResponseDTO(updatedCampaign))
	}
}

// UpdateCampaignStatusHandler atualiza o status de uma campanha
func (h *campaignHandle) UpdateCampaignStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var statusDTO dto.CampaignUpdateStatusDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&statusDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// Validar DTO
		if err := statusDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar `campaign_id` da URL
		campaignIDParam := r.PathValue("campaign_id")
		campaignID, err := uuid.Parse(campaignIDParam)
		if err != nil {
			h.log.Warn("ID inválido informado", "campaign_id", campaignIDParam)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔍 Buscar campanha
		campaign, err := h.repo.GetByID(campaignID)
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// ⚠️ Usuário só pode modificar suas próprias campanhas
		if !authAccount.IsAdmin() && authAccount.ID != campaign.AccountID {
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		// Validar novo status
		if err := statusDTO.Validate(); err != nil {
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Atualizar status
		campaign.Status = statusDTO.Status
		if err := h.repo.UpdateStatus(campaignID, campaign.Status); err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar status")
			return
		}

		updatedCampaign := dto.NewCampaignResponseDTO(campaign)

		h.log.Info("Status da campanha atualizado com sucesso", "campaign_id", campaignID, "status", campaign.Status)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedCampaign)
	}
}

// DeleteCampaignHandler remove uma campanha pelo ID
func (h *campaignHandle) DeleteCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		campaignID := r.PathValue("campaign_id")

		// 🔍 Buscar campanha
		campaign, err := h.repo.GetByID(uuid.MustParse(campaignID))
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// ⚠️ Usuário só pode deletar suas próprias campanhas
		if !authAccount.IsAdmin() && authAccount.ID != campaign.AccountID {
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		err = h.repo.DeleteByID(uuid.MustParse(campaignID))
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao deletar campanha")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Campanha removida com sucesso"})
	}
}
