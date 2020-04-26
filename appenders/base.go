package appenders

import (
	"context"
	"fmt"
	"github.com/andy722/fluent-forwarder/metrics"
	"github.com/andy722/fluent-forwarder/protocol"
	"github.com/andy722/fluent-forwarder/wal"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type baseAppender struct {
	reader *wal.Reader

	labels prometheus.Labels
	msg    protocol.FluentMsg

	flushedOffset wal.Offset
}

func newBaseAppender(reader *wal.Reader) *baseAppender {
	return &baseAppender{
		reader: reader,
		labels: prometheus.Labels{"target": reader.Tag},
		flushedOffset:reader.Offset(),
	}
}

func (appender *baseAppender) runLoop(ctx context.Context, writer func(*protocol.FluentMsg) error) error {

	data, err := appender.reader.Read(ctx)
	if err == context.Canceled {
		return err

	} else if err != nil {
		return fmt.Errorf("entry read failed: %w", err)
	}

	if _, err := appender.msg.UnmarshalMsg(data); err != nil {
		return fmt.Errorf("entry unmarshal failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := writer(&appender.msg)
		if err == nil {
			metrics.ForwardPackets.With(appender.labels).Add(1)
			appender.flushedOffset = appender.reader.Offset()
			break

		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

func (appender *baseAppender) runLoopRaw(ctx context.Context, writer func([]byte) error) error {

	data, err := appender.reader.Read(ctx)
	if err == context.Canceled {
		return err

	} else if err != nil {
		return fmt.Errorf("entry read failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := writer(data)
		if err == nil {
			metrics.ForwardPackets.With(appender.labels).Add(1)
			appender.flushedOffset = appender.reader.Offset()
			break

		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

func (appender *baseAppender) Name() string {
	return appender.reader.Tag
}

func (appender *baseAppender) FlushedOffset() wal.Offset {
	return appender.flushedOffset
}

