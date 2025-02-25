// File: /internal/server/handlers/contact_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type ContactHandle interface {
	CreateContactHandler() http.HandlerFunc
	GetAllContactsHandler() http.HandlerFunc
	GetContactHandler() http.HandlerFunc
	UpdateContactHandler() http.HandlerFunc
	DeleteContactHandler() http.HandlerFunc
}

type contactHandle struct {
	log  *slog.Logger
	repo db.ContactRepository
}

func NewContactHandle(repo db.ContactRepository) ContactHandle {
	return &contactHandle{
		log:  logger.GetLogger(),
		repo: repo,
	}
}

// 📌 Criar Contato
func (h *contactHandle) CreateContactHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var contactDTO dto.ContactCreateDTO

		// 📝 Decodificar JSON
		if err := json.NewDecoder(r.Body).Decode(&contactDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// Definir AccountID como a conta autenticada
		contactDTO.AccountID = authAccount.ID

		// 🔍 Validar DTO
		if err := contactDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Normalize DTO
		contactDTO.Normalize()

		// 📌 Criar o modelo de contato
		contact := &models.Contact{
			AccountID: contactDTO.AccountID,
			Name:      contactDTO.Name,
			Email:     contactDTO.Email,
			WhatsApp:  contactDTO.WhatsApp,
			Gender:    contactDTO.Gender,
			BirthDate: nil,
			Bairro:    contactDTO.Bairro,
			Cidade:    contactDTO.Cidade,
			Estado:    contactDTO.Estado,
			Tags:      contactDTO.Tags,
			History:   contactDTO.History,
		}

		if contactDTO.BirthDate != nil {
			birthDate, err := time.Parse(time.DateOnly, *contactDTO.BirthDate)
			if err != nil {
				h.log.Warn("Data de nascimento inválida", slog.String("error", err.Error()))
				utils.SendError(w, http.StatusBadRequest, "Data de nascimento inválida")
				return
			}
			contact.BirthDate = &birthDate
		}

		// 📌 Criar contato no banco de dados
		createdContact, err := h.repo.Create(r.Context(), contact)
		if err != nil {
			if utils.IsUniqueConstraintError(err) { // Capturar erro de chave única
				h.log.Warn("Tentativa de criar contato já existente",
					slog.String("email", utils.SafeString(contactDTO.Email)),
					slog.String("whatsapp", utils.SafeString(contactDTO.WhatsApp)),
				)
				utils.SendError(w, http.StatusConflict, "E-mail ou WhatsApp já cadastrado")
				return
			}
			h.log.Error("Erro ao inserir contato no banco",
				slog.String("error", err.Error()),
				slog.String("email", utils.SafeString(contactDTO.Email)),
				slog.String("whatsapp", utils.SafeString(contactDTO.WhatsApp)),
			)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao criar contato")
			return
		}

		h.log.Info("Contato criado",
			slog.String("contact_id", createdContact.ID.String()),
			slog.String("email", utils.SafeString(createdContact.Email)),
			slog.String("whatsapp", utils.SafeString(createdContact.WhatsApp)),
		)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dto.NewContactResponseDTO(createdContact))
	}
}

// 📌 Buscar Todos os Contatos
func (h *contactHandle) GetAllContactsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		filters := utils.ExtractQueryFilters(r.URL.Query(), []string{"name", "email", "whatsapp", "cidade", "estado", "bairro", "tags", "interesses", "perfil", "eventos"})
		contacts, err := h.repo.GetByAccountID(r.Context(), authAccount.ID, filters)
		if err != nil {
			h.log.Error("Erro ao buscar contatos", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar contatos")
			return
		}

		var response []dto.ContactResponseDTO
		for _, contact := range contacts {
			response = append(response, dto.NewContactResponseDTO(&contact))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// 📌 Buscar Contato Específico
func (h *contactHandle) GetContactHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		contactID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			h.log.Warn("ID inválido informado", "contact_id", r.PathValue("id"))
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		contact, err := h.repo.GetByID(r.Context(), contactID)
		if err != nil || contact == nil {
			h.log.Warn("Contato não encontrado", "contact_id", contactID)
			utils.SendError(w, http.StatusNotFound, "Contato não encontrado")
			return
		}

		if contact.AccountID != authAccount.ID {
			h.log.Warn("Usuário tentou acessar contato de outra conta", "user_id", authAccount.ID, "contact_id", contactID)
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewContactResponseDTO(contact))
	}
}

// / 📌 Atualizar contato
func (h *contactHandle) UpdateContactHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var contactDTO dto.ContactUpdateDTO

		// 📝 Decodificar JSON
		if err := json.NewDecoder(r.Body).Decode(&contactDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 📌 Capturar `contact_id` da URL
		contactID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			h.log.Warn("ID inválido informado", "contact_id", r.PathValue("id"))
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 📌 Buscar contato no banco
		contact, err := h.repo.GetByID(r.Context(), contactID)
		if err != nil || contact == nil {
			h.log.Warn("Contato não encontrado", "contact_id", contactID)
			utils.SendError(w, http.StatusNotFound, "Contato não encontrado")
			return
		}

		// 📌 Verificar se pertence ao usuário autenticado
		if contact.AccountID != authAccount.ID {
			h.log.Warn("Usuário tentou atualizar contato de outra conta", "user_id", authAccount.ID, "contact_id", contactID)
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		// 📌 Atualizar os campos informados
		if contactDTO.Name != nil {
			contact.Name = *contactDTO.Name
		}
		if contactDTO.Email != nil {
			contact.Email = utils.NormalizeEmail(contact.Email)
		}
		if contactDTO.WhatsApp != nil {
			contact.WhatsApp = utils.FormatWhatsAppOnlyNumbers(contact.WhatsApp)
		}
		if contactDTO.Gender != nil {
			contact.Gender = contactDTO.Gender
		}
		if contactDTO.Bairro != nil {
			contact.Bairro = contactDTO.Bairro
		}
		if contactDTO.Cidade != nil {
			contact.Cidade = contactDTO.Cidade
		}
		if contactDTO.Estado != nil {
			contact.Estado = contactDTO.Estado
		}
		if contactDTO.BirthDate != nil {
			birthDate, err := time.Parse(time.DateOnly, *contactDTO.BirthDate)
			if err != nil {
				h.log.Warn("Data de nascimento inválida", "error", err)
				utils.SendError(w, http.StatusBadRequest, "Data de nascimento inválida")
				return
			}
			contact.BirthDate = &birthDate
		}

		if contactDTO.Tags != nil {
			contact.Tags = *contactDTO.Tags
		}
		if contactDTO.History != nil {
			contact.History = contactDTO.History
		}

		// 📌 Salvar atualização
		updatedContact, err := h.repo.UpdateByID(r.Context(), contactID, contact)
		if err != nil {
			h.log.Error("Erro ao atualizar contato", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar contato")
			return
		}

		h.log.Info("Contato atualizado com sucesso", "contact_id", contactID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewContactResponseDTO(updatedContact))
	}
}

// 📌 Deletar Contato
func (h *contactHandle) DeleteContactHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		contactID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			h.log.Warn("ID inválido informado", "contact_id", r.PathValue("id"))
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		contact, err := h.repo.GetByID(r.Context(), contactID)
		if err != nil || contact == nil {
			h.log.Warn("Contato não encontrado", "contact_id", contactID)
			utils.SendError(w, http.StatusNotFound, "Contato não encontrado")
			return
		}

		if contact.AccountID != authAccount.ID {
			h.log.Warn("Usuário tentou deletar contato de outra conta", "user_id", authAccount.ID, "contact_id", contactID)
			utils.SendError(w, http.StatusForbidden, "Acesso negado")
			return
		}

		if err := h.repo.DeleteByID(r.Context(), contactID); err != nil {
			h.log.Error("Erro ao deletar contato", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao deletar contato")
			return
		}

		h.log.Info("Contato deletado com sucesso", "contact_id", contactID)
		w.WriteHeader(http.StatusNoContent)
	}
}
