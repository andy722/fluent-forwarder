package appenders

import (
	"context"
	"fc/wal"
)

type Appender interface {
	Run(ctx context.Context)
	Name() string
	// FlushedOffset is a position in WAL, data before which is actually *written*
	// into appender's target. May fall behind current WAL *read* offset.
	FlushedOffset() wal.Offset
	Close() error
}
