// File: /internal/server/handlers/campaign_audience_handler.go

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

type CampaignAudienceHandle interface {
	AddContactsToCampaignHandler() http.HandlerFunc
	GetCampaignAudienceHandler() http.HandlerFunc
}

type campaignAudienceHandle struct {
	log          *slog.Logger
	campaignRepo db.CampaignRepository
	contactRepo  db.ContactRepository
	audienceRepo db.CampaignAudienceRepository
}

func NewCampaignAudienceHandle(
	campaignRepo db.CampaignRepository,
	contactRepo db.ContactRepository,
	audienceRepo db.CampaignAudienceRepository,
) CampaignAudienceHandle {
	return &campaignAudienceHandle{
		log:          logger.GetLogger(),
		campaignRepo: campaignRepo,
		contactRepo:  contactRepo,
		audienceRepo: audienceRepo,
	}
}

// ✅ **Adicionar contatos a uma campanha**
func (h *campaignAudienceHandle) AddContactsToCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		var requestDTO dto.CampaignAudienceCreateDTO

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

		// 🔍 Buscar ID da campanha
		campaignID := h.getCampaignIDFromRequest(r, w)

		// 🔍 Buscar contatos e garantir que pertencem ao usuário autenticado
		var validContacts []uuid.UUID
		for _, contactID := range requestDTO.ContactIDs {
			contact, err := h.contactRepo.GetByID(contactID)
			if err != nil || contact == nil {
				h.log.Warn("Contato não encontrado", "contact_id", contactID)
				continue
			}
			if contact.AccountID != authAccount.ID {
				h.log.Warn("Usuário tentou adicionar contato de outra conta", "user_id", authAccount.ID, "contact_id", contactID)
				continue
			}
			validContacts = append(validContacts, contactID)
		}

		if len(validContacts) == 0 {
			h.log.Warn("Nenhum contato válido encontrado para adicionar à campanha")
			utils.SendError(w, http.StatusBadRequest, "Nenhum contato válido encontrado")
			return
		}

		// 🚀 Criar registros na tabela `campaign_audience`
		var audiences []models.CampaignAudience

		if requestDTO.Type == nil {
			for _, channelType := range models.CampaignChannelsTypes {
				for _, contactID := range validContacts {
					audiences = append(audiences, *models.NewCampaignAudience(campaignID, contactID, channelType))
				}
			}
		} else {
			for _, contactID := range validContacts {
				audiences = append(audiences, *models.NewCampaignAudience(campaignID, contactID, *requestDTO.Type))
			}
		}

		// 📦 Salvar registros
		audiencesSaved, err := h.audienceRepo.AddContactsToCampaign(campaignID, audiences)
		if err != nil {
			h.log.Error("Erro ao adicionar contatos à campanha", "campaign_id", campaignID, "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao adicionar contatos à campanha")
			return
		}

		h.log.Info("Contatos adicionados à campanha com sucesso", "campaign_id", campaignID, "total", len(validContacts))
		utils.SendSuccess(w, http.StatusCreated, audiencesSaved)
	}
}

// ✅ **Obter audiência de uma campanha**
func (h *campaignAudienceHandle) GetCampaignAudienceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar ID da campanha
		campaignID := h.getCampaignIDFromRequest(r, w)

		// 🔍 Buscar contatos da audiência da campanha
		var audienceType *string
		audience, err := h.audienceRepo.GetCampaignAudience(campaignID, audienceType)
		if err != nil {
			h.log.Error("Erro ao buscar audiência da campanha", "campaign_id", campaignID, "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar audiência da campanha")
			return
		}

		// 🔄 Converter para DTO
		var response []dto.CampaignAudienceResponseDTO
		for _, aud := range audience {
			response = append(response, dto.NewCampaignAudienceResponseDTO(&aud))
		}

		h.log.Info("Audiência recuperada com sucesso", "campaign_id", campaignID, "total", len(response))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func (h *campaignAudienceHandle) getCampaignIDFromRequest(r *http.Request, w http.ResponseWriter) uuid.UUID {
	// 🔍 Buscar conta autenticada
	authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

	// 🔍 Capturar `campaign_id` da URL
	campaignIDParam := r.PathValue("campaign_id")
	campaignID, err := uuid.Parse(campaignIDParam)
	if err != nil {
		h.log.Warn("ID inválido informado", "campaign_id", campaignIDParam)
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
	}

	// 🔍 Buscar a campanha para garantir que pertence à conta autenticada
	campaign, err := h.campaignRepo.GetByID(campaignID)
	if err != nil || campaign == nil {
		h.log.Warn("Campanha não encontrada", "campaign_id", campaignID)
		utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
	}
	if campaign.AccountID != authAccount.ID {
		h.log.Warn("Usuário tentou acessar audiência de campanha de outra conta", "user_id", authAccount.ID, "campaign_id", campaign.ID)
		utils.SendError(w, http.StatusForbidden, "Acesso negado")
	}

	return campaignID
}
