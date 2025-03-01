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
	GetAvailableContactsHandler() http.HandlerFunc
	GetPaginatedCampaignAudienceHandler() http.HandlerFunc
	AddContactsToCampaignHandler() http.HandlerFunc
	GetCampaignAudienceHandler() http.HandlerFunc
	RemoveContactFromCampaignHandler() http.HandlerFunc
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

// ✅ **Obter contatos disponíveis para uma campanha**
func (h *campaignAudienceHandle) GetAvailableContactsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := r.Context().Value(middleware.AuthAccountKey).(*models.Account)

		// 🔍 Buscar ID da campanha
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		// Validar se a campanha pertence ao usuário autenticado
		h.validateOwnerCampaign(r, w, campaignID)

		// 🔍 Capturar filtros da query string
		filters := utils.ExtractQueryFilters(r.URL.Query(), []string{"name", "email", "whatsapp", "cidade", "estado", "tags"})
		page, perPage, sort := utils.ExtractPaginationParams(r)

		// 🔍 Buscar contatos disponíveis para a campanha
		paginator, err := h.contactRepo.GetAvailableContactsForCampaign(r.Context(), authAccount.ID, campaignID, filters, sort, page, perPage)
		if err != nil {
			h.log.Error("Erro ao buscar contatos disponíveis", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar contatos disponíveis")
			return
		}

		// ✅ Responder com JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(paginator)
	}
}

// ✅ **Adicionar contatos a uma campanha**
func (h *campaignAudienceHandle) AddContactsToCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := r.Context().Value(middleware.AuthAccountKey).(*models.Account)

		// 🔍 Buscar ID da campanha
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

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

		// 🔍 Buscar contatos e garantir que pertencem ao usuário autenticado
		var validContacts []models.Contact
		for _, contactID := range requestDTO.ContactIDs {
			contact, err := h.contactRepo.GetByID(r.Context(), contactID)
			if err != nil || contact == nil {
				h.log.Warn("Contato não encontrado", "contact_id", contactID)
				continue
			}
			if contact.AccountID != authAccount.ID {
				h.log.Warn("Usuário tentou adicionar contato de outra conta", "user_id", authAccount.ID, "contact_id", contactID)
				continue
			}
			validContacts = append(validContacts, *contact)
		}

		if len(validContacts) == 0 {
			h.log.Warn("Nenhum contato válido encontrado para adicionar à campanha")
			utils.SendError(w, http.StatusBadRequest, "Nenhum contato válido encontrado")
			return
		}

		// 🚀 Criar registros na tabela `campaign_audience`
		var audiences []models.CampaignAudience

		// 🔄 Adicionar contatos para todos os tipos de canais
		if requestDTO.Type == nil {
			for _, channelType := range models.AllowedChannels {
				for _, contact := range validContacts {
					// 🛑 Ignorar contatos sem o canal especificado
					if channelType == models.WhatsappChannel && contact.WhatsApp == nil || channelType == models.EmailChannel && contact.Email == nil {
						continue
					}

					audiences = append(audiences, *models.NewCampaignAudience(campaignID, contact.ID, channelType))
				}
			}
		} else { // 🔄 Adicionar contatos para um tipo específico de canal
			for _, contact := range validContacts {
				// 🛑 Ignorar contatos sem o canal especificado
				if *requestDTO.Type == models.WhatsappChannel && contact.WhatsApp == nil || *requestDTO.Type == models.EmailChannel && contact.Email == nil {
					continue
				}

				audiences = append(audiences, *models.NewCampaignAudience(campaignID, contact.ID, *requestDTO.Type))
			}
		}

		// 📦 Salvar registros
		audiencesSaved, err := h.audienceRepo.AddContactsToCampaign(r.Context(), campaignID, audiences)
		if err != nil {
			h.log.Error("Erro ao adicionar contatos à campanha", "campaign_id", campaignID, "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao adicionar contatos à campanha")
			return
		}

		h.log.Info("Contatos adicionados à campanha com sucesso", "campaign_id", campaignID, "total", len(validContacts))
		utils.SendSuccess(w, http.StatusCreated, audiencesSaved)
	}
}

// ✅ **Obter audiência da campanha com paginação**
func (h *campaignAudienceHandle) GetPaginatedCampaignAudienceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar ID da campanha
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		h.validateOwnerCampaign(r, w, campaignID)

		// 🔍 Capturar parâmetros de paginação
		page, perPage, _ := utils.ExtractPaginationParams(r)

		// 🔍 Capturar filtro opcional `type`
		var contactType *string
		if typeParam := r.URL.Query().Get("type"); typeParam != "" {
			contactType = &typeParam
		}

		// 🔍 Buscar audiência paginada
		paginator, err := h.audienceRepo.GetPaginatedCampaignAudience(r.Context(), campaignID, contactType, page, perPage)
		if err != nil {
			h.log.Error("Erro ao buscar audiência paginada", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar audiência da campanha")
			return
		}

		// ✅ Responder com JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(paginator)
	}
}

// ✅ **Obter audiência de uma campanha**
func (h *campaignAudienceHandle) GetCampaignAudienceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar ID da campanha
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		h.validateOwnerCampaign(r, w, campaignID)

		// 🔍 Buscar contatos da audiência da campanha
		var audienceType *string
		audience, err := h.audienceRepo.GetCampaignAudience(r.Context(), campaignID, audienceType)
		if err != nil {
			h.log.Error("Erro ao buscar audiência da campanha", "campaign_id", campaignID, "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar audiência da campanha")
			return
		}

		h.log.Info("Audiência recuperada com sucesso", "campaign_id", campaignID, "total", len(audience))
		// Responder com JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(audience)
	}
}

// ✅ **Remover contato de uma campanha**
func (h *campaignAudienceHandle) RemoveContactFromCampaignHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar ID da campanha
		campaignID := utils.GetUUIDFromRequestPath(r, w, "campaign_id")

		h.validateOwnerCampaign(r, w, campaignID)

		// 🔍 Buscar ID do contato
		audienceID := utils.GetUUIDFromRequestPath(r, w, "audience_id")
		if audienceID == uuid.Nil {
			return
		}

		// 🚀 Remover contato da campanha
		if err := h.audienceRepo.RemoveContactFromCampaign(r.Context(), campaignID, audienceID); err != nil {
			h.log.Error("Erro ao remover contato da campanha", "campaign_id", campaignID, "contact_id", audienceID, "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao remover contato da campanha")
			return
		}

		h.log.Info("Contato removido da campanha com sucesso", "campaign_id", campaignID, "contact_id", audienceID)
		utils.SendSuccess(w, http.StatusNoContent, nil)
	}
}

func (h *campaignAudienceHandle) validateOwnerCampaign(r *http.Request, w http.ResponseWriter, campaignID uuid.UUID) {
	// 🔍 Buscar conta autenticada
	authAccount := r.Context().Value(middleware.AuthAccountKey).(*models.Account)

	// 🔍 Buscar a campanha para garantir que pertence à conta autenticada
	campaign, err := h.campaignRepo.GetByID(r.Context(), campaignID)
	if err != nil || campaign == nil {
		h.log.Warn("Campanha não encontrada", "campaign_id", campaignID)
		utils.SendError(w, http.StatusNotFound, "Campanha não encontrada")
	}

	// 🔍 Garantir que a campanha pertence à conta autenticada
	if campaign.AccountID != authAccount.ID {
		h.log.Warn("Usuário tentou acessar audiência de campanha de outra conta", "user_id", authAccount.ID, "campaign_id", campaign.ID)
		utils.SendError(w, http.StatusForbidden, "Acesso negado")
	}
}
