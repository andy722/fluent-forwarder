package inputs

import (
	"context"
	"fc/protocol"
	. "fc/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type ForwardInput struct {
	network  string
	listener net.Listener

	handler  func([]byte) error
}

type ForwardHandler func([]byte) error

func NewForwardInput(ListenOn string, handler ForwardHandler) (*ForwardInput, error) {

	network, address, err := SplitAddress(ListenOn)
	if err != nil {
		return nil, err
	}

	if network == "unix" {
		if err := os.RemoveAll(address); err != nil {
			return nil, fmt.Errorf("%v: %w", address, err)
		}
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	log.Infof("Bound to %v://%v", network, address)

	return &ForwardInput{
		network:  network,
		listener: listener,
		handler:  handler,
	}, nil
}

func (input *ForwardInput) Name() string {
	return fmt.Sprintf("fluentd forward: %v", input.listener.Addr().String())
}

func (input *ForwardInput) Run(ctx context.Context) {
	defer input.closeListener()

	var tcpListener *net.TCPListener
	var unixListener *net.UnixListener
	if l, ok := input.listener.(*net.TCPListener); ok {
		tcpListener = l
	} else if l, ok := input.listener.(*net.UnixListener); ok {
		unixListener = l
	} else {
		log.Panicf("Unexpected listener %v", input.listener)
	}

	var handlers sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			handlers.Wait()
			return
		default:

			if tcpListener != nil {
				if err := tcpListener.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
					log.Panicf("%v", err)
				}

			} else if unixListener != nil {
				if err := unixListener.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
					log.Panicf("%v", err)
				}
			}

			conn, err := input.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}

				log.Warningf("Failed to accept connection: %v", err)
			}

			handlers.Add(1)
			go func() {
				if err := input.handleConnection(conn, ctx); err != nil {
					log.Warningf("Connection error: %v", err)
				}
				handlers.Done()
			}()
		}
	}

}

func (input *ForwardInput) handleConnection(conn net.Conn, ctx context.Context) (err error) {
	var (
		_log = log.WithField("remote", conn.RemoteAddr())

		reader  = msgp.NewReaderSize(conn, 16 * 1024)
		msgpEOF = msgp.WrapError(io.EOF)

		readMsg  = protocol.FluentRawMsg{}
		writeBuf = make([]byte, 1024)
	)

	_log.Debug("Accepted connection")

	defer func() {
		_log.Trace("Closing connection")
		if err := conn.Close(); err != nil {
			_log.Warnf("Failed to close connection: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			log.Panicf("%v", err)
		}

		if err = readMsg.DecodeMsg(reader); err == msgpEOF {
			_log.Debug("Remote closed")
			return nil

		} else if netError, ok := msgp.Cause(err).(*net.OpError); ok && netError.Timeout() {
			continue

		} else if err != nil {
			return
		}

		//if log.IsLevelEnabled(log.TraceLevel) {
		//	_log.Tracef("IN: %+v", readMsg)
		//}

		writeBuf = writeBuf[:0]
		if writeBuf, err = readMsg.MarshalMsg(writeBuf); err != nil {
			return err
		}

		if err := input.handler(writeBuf); err != nil {
			return fmt.Errorf("failed to process: %v: %w", readMsg, err)
		}
	}
}

func (input *ForwardInput) closeListener() {
	if err := input.listener.Close(); err != nil {
		log.Warn("Failed to close listener: ", err)
	}
}
