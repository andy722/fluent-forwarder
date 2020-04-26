package appenders

import (
	"bufio"
	"context"
	"fmt"
	"github.com/andy722/fluent-forwarder/util"
	"github.com/andy722/fluent-forwarder/wal"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type ForwardAppender struct {
	*baseAppender

	target string

	conn   net.Conn
	writer *bufio.Writer
}

func NewForwardAppender(reader *wal.Reader, target string) (appender *ForwardAppender, err error) {
	network, address, err := util.SplitAddress(target)
	if err != nil {
		return nil, err
	} else if network != "tcp" {
		return nil, fmt.Errorf("%v: invalid address", target)
	}

	appender = &ForwardAppender{
		baseAppender: newBaseAppender(reader),
		target:       address,
	}
	return appender, appender.reconnect()
}

func (appender *ForwardAppender) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if err := appender.reader.Close(); err != nil {
				appender.log().Warn("Failed to close WAL reader ", err)
			}
			return
		default:
		}

		err := appender.runLoopRaw(ctx, func(msg []byte) (err error) {
			if err = appender.tryWriteRaw(msg); err != nil {
				appender.log().Warn("Entry write failed, will retry: ", err)
			}
			return
		})

		if err == context.Canceled {
			return

		} else if err != nil {
			appender.log().Warn("Append failed: ", err)
		}
	}
}

func (appender *ForwardAppender) tryWriteRaw(msg []byte) (err error) {
	// TODO: gzip

	if _, err = appender.writer.Write(msg); err != nil {
		err = appender.reconnect()
		return
	}

	return
}

func (appender *ForwardAppender) reconnect() (err error) {

	_ = appender.closeConnection()

	appender.conn, err = net.DialTimeout("tcp", appender.target, 30*time.Second)

	if appender.writer != nil {
		err = appender.writer.Flush()
		appender.writer.Reset(appender.conn)
	} else {
		appender.writer = bufio.NewWriterSize(appender.conn, 16*1024)
	}

	return err
}

func (appender *ForwardAppender) log() *log.Entry {
	return log.WithFields(log.Fields{
		"offset": appender.reader.Offset().String(),
		"target": appender.target,
	})
}

func (appender *ForwardAppender) closeConnection() (err error) {
	if appender.conn != nil {
		if appender.writer != nil {
			if err = appender.writer.Flush(); err != nil {
				appender.log().Warn("Failed to flush: ", err)
			}
		}

		if err = appender.conn.Close(); err != nil {
			appender.log().Warn("Failed to close connection: ", err)
		}
	}
	return err
}

func (appender *ForwardAppender) Close() (err error) {
	if err = appender.closeConnection(); err != nil {
		log.Warnf("Connection close failed: %v", err)
	}

	appender.log().Debugf("Closing reader %v", appender.reader.Tag)
	return appender.reader.Close()
}
