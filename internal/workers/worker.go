// File: /internal/workers/worker.go

package workers

import "context"

// Worker define um contrato comum para todos os workers
type Worker interface {
	Start(ctx context.Context)
}
