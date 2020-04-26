package appenders

import (
	"github.com/andy722/fluent-forwarder/protocol"
	"io"
)

// FileFormatter appends single message into writable file.
type FileFormatter func (tag string, msg *protocol.FluentMsg, out io.StringWriter) error

// SimpleFileFormatter outputs tag (or container name in case of Swarm deployment) and log message,
// separated by TABs.
func SimpleFileFormatter(tag string, msg *protocol.FluentMsg, out io.StringWriter) (err error) {

	if len(msg.Record.ContainerName) > 0 {
		tag = protocol.ReadRaw(msg.Record.ContainerName)
	}

	_, err = out.WriteString(tag)
	if err != nil {
		return
	}

	_, err = out.WriteString("\t")
	if err != nil {
		return
	}

	_, err = out.WriteString(protocol.ReadRaw(msg.Record.Log))
	if err != nil {
		return
	}

	_, err = out.WriteString("\n")
	if err != nil {
		return
	}

	return nil
}