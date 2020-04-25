package appenders

import (
	"bufio"
	"context"
	"fc/protocol"
	. "fc/util"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type tagTarget struct {
	sync.Mutex

	currentPath string

	currentFile       *os.File
	currentFileWriter *bufio.Writer

	rotationCheckTimer *time.Ticker
	flushTicker        *time.Ticker
}

func newTagTarget(ctx context.Context, currentPath string) (target *tagTarget, err error) {

	target = &tagTarget{
		rotationCheckTimer: time.NewTicker(1 * time.Minute),
		flushTicker:        time.NewTicker(1 * time.Second),
	}

	if err = target.ResetPath(currentPath); err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-target.flushTicker.C:
				target.Lock()
				if err := target.currentFileWriter.Flush(); err != nil {
					log.Warnf("%v: flush failed: %v", target.currentPath, err)
				}
				target.Unlock()
			case <-ctx.Done():
				return
			}
		}

	}()

	return target, nil
}

func (target *tagTarget) Write(tag string, msg *protocol.FluentMsg) (err error) {
	target.Lock()
	defer target.Unlock()

	writer := target.currentFileWriter

	if len(msg.Record.ContainerName) > 0 {
		tag = protocol.ReadRaw(msg.Record.ContainerName)
	}

	_, err = writer.WriteString(tag)
	if err != nil {
		return
	}

	_, err = writer.WriteString("\t")
	if err != nil {
		return
	}

	_, err = writer.WriteString(protocol.ReadRaw(msg.Record.Log))
	if err != nil {
		return
	}

	_, err = writer.WriteString("\n")
	if err != nil {
		return
	}

	err = writer.Flush()
	if err != nil {
		return
	}

	return
}

func (target *tagTarget) ResetPath(newPath string) error {
	target.currentPath = newPath
	return target.open()
}

func (target *tagTarget) open() (err error) {
	target.Lock()
	defer target.Unlock()

	if err = target.Close(); err != nil {
		return
	}

	dir, _ := filepath.Split(target.currentPath)
	if err := EnsureDirExists(dir); err != nil {
		return err
	}

	target.currentFile, err =
		os.OpenFile(target.currentPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if target.currentFileWriter != nil {
		target.currentFileWriter.Reset(target.currentFile)
	} else {
		target.currentFileWriter = bufio.NewWriterSize(target.currentFile, 4 * 1024)
	}

	return err
}

func (target *tagTarget) Close() (err error) {

	if target.currentFile != nil {
		if err = target.currentFileWriter.Flush(); err != nil {
			return
		}
		if err = target.currentFile.Close(); err != nil {
			return
		}
		target.currentFile = nil
	}

	return
}