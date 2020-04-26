package inputs

import (
	"context"
)

// Input receives log entries and stores to WAL.
type Input interface {
	Run(ctx context.Context)
	Name() string
}
