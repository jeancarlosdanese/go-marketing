// File: internal/internal_error/internal_error.go

// Package internalerror é responsável por criar erros internos padronizados
// para serem utilizados em todo o sistema.
// O pacote possui três funções para criar erros padronizados:
// - NewNotFoundError: cria um erro de recurso não encontrado
// - NewInternalServerError: cria um erro interno
// - NewBadRequestError: cria um erro de requisição inválida
// Para utilizar o pacote, basta importar o pacote e chamar a função desejada.
// Exemplo:
//
//	err := internalerror.NewNotFoundError("Recurso não encontrado")
//	err := internalerror.NewInternalServerError("Erro interno")
//	err := internalerror.NewBadRequestError("Requisição inválida")
//	fmt.Println(err.Error())
//	fmt.Println(err.Err)
//	fmt.Println(err.Message)
//	fmt.Println(err)

package internalerror

// InternalError representa um erro interno padronizado
type InternalError struct {
	Message string
	Err     string
}

// Error retorna a mensagem do erro
func (ie *InternalError) Error() string {
	return ie.Message
}

// NewNotFoundError cria um erro de recurso não encontrado
func NewNotFoundError(message string) *InternalError {
	return &InternalError{
		Message: message,
		Err:     "not_found",
	}
}

// NewInternalServerError cria um erro interno
func NewInternalServerError(message string) *InternalError {
	return &InternalError{
		Message: message,
		Err:     "internal_server",
	}
}

// NewBadRequestError cria um erro de requisição inválida
func NewBadRequestError(message string) *InternalError {
	return &InternalError{
		Message: message,
		Err:     "bad_request",
	}
}

// // NewUnauthorizedError cria um erro de não autorizado
// func NewUnauthorizedError(message string) *InternalError {
// 	return &InternalError{
// 		Message: message,
// 		Err:     "unauthorized",
// 	}
// }

// // NewForbiddenError cria um erro de acesso proibido
// func NewForbiddenError(message string) *InternalError {
// 	return &InternalError{
// 		Message: message,
// 		Err:     "forbidden",
// 	}
// }

// // NewConflictError cria um erro de conflito
// func NewConflictError(message string) *InternalError {
// 	return &InternalError{
// 		Message: message,
// 		Err:     "conflict",
// 	}
// }

// // NewValidationError cria um erro de validação
// func NewValidationError(message string) *InternalError {
// 	return &InternalError{
// 		Message: message,
// 		Err:     "validation_error",
// 	}
// }
