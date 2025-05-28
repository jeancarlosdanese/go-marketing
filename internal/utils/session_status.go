// internal/utils/session_status.go

package utils

// ParseSessionStatusAPIResponse interpreta a resposta da API Baileys
// e retorna um dos valores padronizados do sistema.
func ParseSessionStatusAPIResponse(res map[string]interface{}) string {
	if qrCode, ok := res["qrCode"]; ok && qrCode != nil {
		return "aguardando_qr"
	}

	if status, ok := res["status"].(string); ok {
		switch status {
		case "connected":
			return "conectado"
		case "connecting":
			return "aguardando_qr" // ou "desconhecido" se preferir
		case "disconnected":
			return "desconectado"
		case "error":
			return "erro"
		case "timeout":
			return "qrcode_expirado"
		default:
			return "desconhecido"
		}
	}

	return "desconhecido"
}

// IsValidSessionStatus verifica se o status informado está entre os valores aceitos.
func IsValidSessionStatus(status string) bool {
	switch status {
	case "desconhecido", "aguardando_qr", "qrcode_expirado", "conectado", "desconectado", "erro":
		return true
	default:
		return false
	}
}
