// File: /internal/audit/audit.go

package audit

import (
	"fmt"
	"time"

	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

// AuditEvent representa um evento de auditoria
type AuditEvent struct {
	Timestamp time.Time
	User      string
	Action    string
	Details   string
}

// LogEvent registra um evento de auditoria
func LogEvent(user, action, details string) {
	log := logger.GetLogger()
	event := AuditEvent{
		Timestamp: time.Now(),
		User:      user,
		Action:    action,
		Details:   details,
	}

	log.Info(fmt.Sprintf("AUDIT: [%s] %s - %s - %s",
		event.Timestamp.Format(time.RFC3339),
		event.User,
		event.Action,
		event.Details,
	))
}
