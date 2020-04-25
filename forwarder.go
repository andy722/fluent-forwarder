package main

import (
	"context"
	"fc/appenders"
	"fc/inputs"
	"fc/wal"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
	"time"
)

type ForwarderOptions struct {
	InputFwd    string

	WalPath     string

	LogPath     string
	ForwardAddr string
}

type Forwarder struct {
	ForwarderOptions

	wal  *wal.Wal
	meta *wal.Meta

	inputs []inputs.Input
	appenders []appenders.Appender
}

func NewForwarder(opts ForwarderOptions) (rc *Forwarder, err error) {
	rc = &Forwarder{ForwarderOptions: opts}

	//
	//	Init  WAL
	//
	rc.wal, err = wal.NewWal(rc.WalPath)
	if err != nil {
		return nil, fmt.Errorf("WAL initialization failed: %w", err)
	}

	rc.meta, err = wal.NewWalMeta(filepath.Join(rc.WalPath, "wal-rd.bin"))
	if err != nil {
		return nil, fmt.Errorf("WAL meta initialization failed: %w", err)
	}

	//
	//	Init appenders
	//
	if rc.LogPath != "" {
		err = rc.initAppender("file",
			func(reader *wal.Reader) (appender appenders.Appender, err error) {
				return appenders.NewFileAppender(reader, rc.LogPath)
			})
		if err != nil {
			return nil, fmt.Errorf("appender [%v] initialization failed: %w", "file", err)
		}
	}

	if rc.ForwardAddr != "" {
		err = rc.initAppender("forward",
			func(reader *wal.Reader) (appender appenders.Appender, err error) {
				return appenders.NewForwardAppender(reader, rc.ForwardAddr)
			})
		if err != nil {
			return nil, fmt.Errorf("appender [%v] initialization failed: %w", "forward", err)
		}
	}

	//
	//	Init input
	//
	if rc.InputFwd != "" {
		err = rc.initInput(func() (inputs.Input, error) {
			return inputs.NewForwardInput(rc.InputFwd, rc.wal.Handle)
		})
		if err != nil {
			return nil, fmt.Errorf("input [%v] initialization failed: %w", "forward", err)
		}
	}

	return
}

func (forwarder *Forwarder) initAppender(
	id string,
	creator func(*wal.Reader) (appenders.Appender, error)) error {

	offset, err := forwarder.meta.GetOffset(id)
	if err != nil {
		return err
	}

	log.Debugf("Initial offset [%v]: %+v", id, offset)

	reader, err := forwarder.wal.NewReader(id, offset)
	if err != nil {
		return err
	}

	if appender, err := creator(reader); err != nil {
		return err

	} else {
		forwarder.appenders = append(forwarder.appenders, appender)
		log.Infof("Initialized appender %v", appender.Name())
		return nil
	}
}

func (forwarder *Forwarder) initInput(
	creator func() (inputs.Input, error)) error {

	if input, err := creator(); err != nil {
		return err

	} else {
		forwarder.inputs = append(forwarder.inputs, input)
		log.Infof("Initialized input %v", input.Name())
		return nil
	}
}

func (forwarder *Forwarder) Run(ctx context.Context) {

	wg := sync.WaitGroup{}

	//
	//	Start input
	//
	for _, input := range forwarder.inputs {
		wg.Add(1)
		go func(i inputs.Input) {
			defer wg.Done()
			defer log.Debugf("Stopped: input %v", i.Name())
			i.Run(ctx)
		}(input)
	}

	//
	//	Start appenders
	//
	for _, appender := range forwarder.appenders {
		wg.Add(1)

		go func(a appenders.Appender) {
			defer wg.Done()
			defer func() {
				if err := a.Close(); err != nil && err != context.Canceled {
					log.Warnf("Stopped: appender %v: ", a.Name(), err)

				} else {
					log.Debugf("Stopped: appender %v", a.Name())
				}
			}()

			a.Run(ctx)
		}(appender)
	}

	//
	//	Start WAL flush
	//
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Debug("Stopped: flush")

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if _, err := forwarder.flushReadIndex(); err != nil {
					log.Errorf("Flush index failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}

	}()

	//
	//	Start WAL cleanup
	//
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Debug("Stopped: clean")

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := forwarder.cleanBuffers(); err != nil {
					log.Errorf("Clean buffers failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()

	if err := forwarder.wal.Close(); err != nil {
		log.Warn("Failed to close WAL", err)
	}
}

func (forwarder *Forwarder) flushReadIndex() (*wal.Offset, error) {
	meta, err := forwarder.meta.Read()
	if err != nil {
		return nil, err
	}

	minOffset := wal.ZeroOffset
	for _, appender := range forwarder.appenders {
		offset := appender.FlushedOffset()
		meta.Readers[appender.Name()] = offset

		if minOffset == wal.ZeroOffset || minOffset.After(offset) {
			minOffset = offset
		}
	}

	return &minOffset, forwarder.meta.Write(meta)
}

func (forwarder *Forwarder) cleanBuffers() error {
	minOffset, err := forwarder.flushReadIndex()
	if err != nil {
		return err
	}

	return forwarder.wal.DropBefore(minOffset)
}
