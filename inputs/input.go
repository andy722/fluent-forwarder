package inputs

import (
	"context"
)

type Input interface {
	Run(ctx context.Context)
	Name() string
}
