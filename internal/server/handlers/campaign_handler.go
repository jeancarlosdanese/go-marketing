// File: /internal/server/handlers/campaign_handler.go

package handlers

import (
	"bytes"
	"context"
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
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type CampaignHandle interface {
	CreateCampaignHandler() http.HandlerFunc
	GetCampaignHandler() http.HandlerFunc
	GetAllCampaignsHandler() http.HandlerFunc
	UpdateCampaignHandler() http.HandlerFunc
	UpdateCampaignStatusHandler() http.HandlerFunc
	GetCampaignStatusHandler() http.HandlerFunc
	DeleteCampaignHandler() http.HandlerFunc
}

type campaignHandle struct {
	log               *slog.Logger
	campaignRepo      db.CampaignRepository
	audienceRepo      db.CampaignAudienceRepository
	campaignProcessor service.CampaignProcessorService
}

func NewCampaignHandle(
	campaignRepo db.CampaignRepository,
	audienceRepo db.CampaignAudienceRepository,
	campaignProcessor service.CampaignProcessorService,
) CampaignHandle {
	return &campaignHandle{
		log:               logger.GetLogger(),
		campaignRepo:      campaignRepo,
		audienceRepo:      audienceRepo,
		campaignProcessor: campaignProcessor,
	}
}

// CreateCampaignHandler cria uma nova campanha
func (h *campaignHandle) CreateCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		h.log.Debug("Criando nova campanha", "body", string(bodyBytes))
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var campaignDTO dto.CampaignCreateDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&campaignDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		for channel, config := range campaignDTO.Channels {
			if config.TemplateID == uuid.Nil {
				h.log.Warn("Removendo template vazio", "channel", channel)
				delete(campaignDTO.Channels, channel)
			}
		}

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
			// Filters:     campaignDTO.Filters,
			Status: models.StatusPendente,
		}

		createdCampaign, err := h.campaignRepo.Create(r.Context(), campaign)
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
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		// 🔍 Buscar campanha no banco
		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil {
			h.log.Error("Erro ao buscar campanha", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar campanha")
			return
		}
		if campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// Checar se é admin ou dono
		if !middleware.IsAdminOrOwner(authAccount, campaign.AccountID) {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
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

		// 🔍 Capturar filtros da URL (ex.: `?status=active&name=promo`)
		filters := map[string]string{}
		if status := r.URL.Query().Get("status"); status != "" {
			filters["status"] = status
		}
		if name := r.URL.Query().Get("name"); name != "" {
			filters["name"] = name
		}
		if createdAfter := r.URL.Query().Get("created_after"); createdAfter != "" {
			filters["created_after"] = createdAfter
		}
		if createdBefore := r.URL.Query().Get("created_before"); createdBefore != "" {
			filters["created_before"] = createdBefore
		}

		// 🔍 Buscar campanhas com os filtros aplicados
		campaigns, err := h.campaignRepo.GetAllByAccountID(r.Context(), authAccount.ID, &filters)
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

		h.log.Info("Campanhas recuperadas com sucesso", "total", len(response), "filters", filters)
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
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		// 🔍 Buscar campanha
		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// Checar se é admin ou dono
		if !middleware.IsAdminOrOwner(authAccount, campaign.AccountID) {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
			return
		}

		// Remover filtros vazios antes de atualizar
		if updateDTO.Filters != nil {
			// 🔥 Remover `gender` se estiver vazio
			if updateDTO.Filters.Gender != nil && *updateDTO.Filters.Gender == "" {
				h.log.Warn("Removendo filtro vazio", "campo", "gender")
				updateDTO.Filters.Gender = nil
			}

			// 🔥 Verificar se `birth_date_range` está vazio
			if updateDTO.Filters.BirthDateRange != nil {
				// 🔥 Verificar se `start` e `end` estão vazios
				if updateDTO.Filters.BirthDateRange.Start == nil || *updateDTO.Filters.BirthDateRange.Start == "" {
					h.log.Warn("Removendo filtro vazio", "campo", "birth_date_range.start")
					updateDTO.Filters.BirthDateRange.Start = nil // 🔥 Resetar para um struct vazio
				}
				// 🔥 Verificar se `end` está vazio
				if updateDTO.Filters.BirthDateRange.End == nil || *updateDTO.Filters.BirthDateRange.End == "" {
					h.log.Warn("Removendo filtro vazio", "campo", "birth_date_range.end")
					updateDTO.Filters.BirthDateRange.End = nil // 🔥 Resetar para um struct vazio
				}
				// 🔥 Remover `birth_date_range` se ambos estiverem vazios
				if updateDTO.Filters.BirthDateRange.Start == nil && updateDTO.Filters.BirthDateRange.End == nil {
					h.log.Warn("Removendo filtro vazio", "campo", "birth_date_range")
					updateDTO.Filters.BirthDateRange = nil // 🔥 Resetar para um struct vazio
				}
			}

			// 🔥 Remover `tags` se estiver vazio
			if updateDTO.Filters.Tags != nil && (len(updateDTO.Filters.Tags) == 0 || (updateDTO.Filters.Tags[0] != nil && *updateDTO.Filters.Tags[0] == "")) {
				h.log.Warn("Removendo filtro vazio", "campo", "tags")
				updateDTO.Filters.Tags = nil
			}

			// Remove filtros vazios
			if updateDTO.Filters.Tags == nil && updateDTO.Filters.BirthDateRange == nil && updateDTO.Filters.Gender == nil {
				h.log.Warn("Removendo filtros vazios", "campo", "filters")
				updateDTO.Filters = nil
			}
		}

		// // Atualizar filtros
		// campaign.Filters = updateDTO.Filters

		// Atualizar campanha
		if updateDTO.Name != nil {
			campaign.Name = *updateDTO.Name
		}
		campaign.Description = updateDTO.Description
		if updateDTO.Channels != nil {
			campaign.Channels = *updateDTO.Channels
		}
		campaign.Status = *updateDTO.Status

		// Validar updateDTO
		if err := updateDTO.Validate(); err != nil {
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Atualizar campanha no banco
		updatedCampaign, err := h.campaignRepo.UpdateByID(r.Context(), campaignID, campaign)
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
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		// 🔍 Buscar campanha
		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// Checar se é admin ou dono
		if !middleware.IsAdminOrOwner(authAccount, campaign.AccountID) {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
			return
		}

		// 🔍 Se for ativar a campanha, verificar se há contatos na audiência
		if statusDTO.Status == "processando" {
			audience, err := h.audienceRepo.GetCampaignAudienceToSQS(r.Context(), authAccount.ID, campaignID, nil)
			if err != nil {
				h.log.Error("Erro ao buscar audiência", "campaign_id", campaignID, "error", err)
				utils.SendError(w, http.StatusInternalServerError, "Erro ao verificar audiência")
				return
			}
			if len(audience) == 0 {
				h.log.Warn("Tentativa de ativar campanha sem audiência", "campaign_id", campaignID)
				utils.SendError(w, http.StatusBadRequest, "Não é possível ativar uma campanha sem audiência")
				return
			}

			// 🟡 Atualizar status da campanha para "processando"
			campaign.Status = "processando" // 🟡 Define status intermediário
			if err := h.campaignRepo.UpdateStatus(r.Context(), campaignID, "processando"); err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar status")
				return
			}

			// 🚀 Iniciar worker para processar campanha
			go func() {
				h.log.Info("Iniciando worker de envio de mensagens", "campaign_id", campaignID)
				err := h.campaignProcessor.ProcessCampaign(r.Context(), campaign, audience)
				if err != nil {
					h.log.Error("Erro no processamento da campanha", "campaign_id", campaignID, "error", err)
				} else {
					// ✅ Atualiza para "ativa" após processar
					h.campaignRepo.UpdateStatus(context.Background(), campaignID, "processando")
				}
			}()
		}

		h.log.Info("Status da campanha atualizado com sucesso", "campaign_id", campaignID, "status", campaign.Status)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": string(campaign.Status)})
	}
}

// GetCampaignStatusHandler retorna o status de uma campanha
func (h *campaignHandle) GetCampaignStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": string(campaign.Status)})
	}
}

// DeleteCampaignHandler remove uma campanha pelo ID
func (h *campaignHandle) DeleteCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar `campaign_id` da URL
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		// 🔍 Buscar campanha
		campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
		if err != nil || campaign == nil {
			utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
			return
		}

		// 🚨 Verificar se a campanha não está ativa ou concluída
		if campaign.Status != models.StatusPendente {
			h.log.Warn("Tentativa de exclusão de campanha não pendente", "campaign_id", campaignID, "status", campaign.Status)
			utils.SendError(w, http.StatusForbidden, "Apenas campanhas pendentes podem ser excluídas")
			return
		}

		// Checar se é admin ou dono
		if !middleware.IsAdminOrOwner(authAccount, campaign.AccountID) {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
			return
		}

		err = h.campaignRepo.DeleteByID(r.Context(), campaignID)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao deletar campanha")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Campanha removida com sucesso"})
	}
}
