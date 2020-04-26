package wal

import (
	"bufio"
	"context"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Reader struct {
	Tag string

	segmentPtr
	wal *Wal

	// Temporary buffers
	header []byte
	body   []byte

	reader *bufio.Reader
}

func (wal *Wal) NewReader(tag string, offset Offset) (r *Reader, err error) {
	r = &Reader{
		segmentPtr: segmentPtr{
			path: wal.path,
		},
		wal: wal,
		Tag: tag,

		header: make([]byte, 4),
		body:   make([]byte, 1024),
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		defer r.log().Tracef("Opening WAL reader [%v] starting from %v", tag, offset)
	}

	segments, err := wal.segments(1)
	if err != nil {
		return nil, err
	}

	for _, id := range segments {
		if id < offset.Segment {
			continue
		}

		r.offset.Segment = id
		if r.offset.Segment == offset.Segment {
			r.offset.Position = offset.Position

		} else {
			r.offset.Position = 0
		}

		if err = r.open(); err != nil {
			return
		}

		break
	}

	return
}

func (r *Reader) Read(ctx context.Context) (data []byte, err error) {
	if r.offset == ZeroOffset {
		if err = r.nextSegment(ctx); err != nil {
			return
		}
	}

	length, err := r.readHeader(ctx)
	if err != nil {
		return nil, err
	}

	data, err = r.readData(ctx, length)
	return
}

func (r *Reader) readHeader(ctx context.Context) (int, error) {
top:
	for {
		length := 0
		read := 0

		for {
			select {
			case <-ctx.Done():
				return 0, context.Canceled
			default:
			}

			n, err := r.reader.Read(r.header[read:])
			read += n
			r.offset.Position += int64(n)
			if err != nil && err.Error() == "EOF" && read < 4 {
				if r.wal.isAdvancedAfter(r.offset.Segment) {
					advanceErr := r.nextSegment(ctx)
					if advanceErr != nil {
						return 0, advanceErr
					}
					continue top
				}
				// No newer log files, continue trying to read from this one
				time.Sleep(2 * time.Second)
				continue
			}
			if err != nil {
				r.log().Warnf("Header read failed %v: %v", r.fullName(), err)
				break
			}
			if read == 4 {
				length = int(binary.BigEndian.Uint32(r.header))
				break
			}
		}

		if length > 0 {
			return length, nil
		}

		err := r.nextSegment(ctx)
		if err != nil {
			return 0, err
		}
	}
}

func (r *Reader) readData(ctx context.Context, length int) ([]byte, error) {
	//if log.IsLevelEnabled(log.TraceLevel) {
	//	r.log().Trace("readData")
	//}

	if length >= cap(r.body) {
		r.body = make([]byte, length)
	}
	buf := r.body[:length]

	// Read into buffer
	read := 0
	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
		}

		n, err := r.reader.Read(buf[read:])
		read += n
		r.offset.Position += int64(n)

		if err != nil && err.Error() == "EOF" && read < length {
			// Not flushed? Continue trying to read from this one
			time.Sleep(2 * time.Second)
			continue

		} else if err != nil {
			return nil, err
		}

		if read == length {
			return buf, nil
		}
	}
}

func (r *Reader) Offset() Offset {
	return r.offset
}

func (r *Reader) Close() (err error) {
	if err = r.currentFile.Close(); err == os.ErrClosed {
		err = nil
	}
	return
}

func (r *Reader) open() error {
	if log.IsLevelEnabled(log.TraceLevel) {
		r.log().Trace("open")
	}

	err := r.openFile(os.O_RDONLY)
	if err != nil {
		return err
	}

	if r.reader != nil {
		r.reader.Reset(r.currentFile)

	} else {
		r.reader = bufio.NewReaderSize(r.currentFile, 16*1024)
	}

	if r.offset.Position > 0 {
		if _, err := r.reader.Discard(int(r.offset.Position)); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) nextSegment(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		segments, err := r.wal.segments(1)
		if err != nil {
			return err
		}

		for _, id := range segments {
			if id <= r.offset.Segment {
				continue
			}

			r.offset = Offset{Segment: id}
			return r.open()
		}

		if log.IsLevelEnabled(log.TraceLevel) {
			r.log().Trace("No next offset, wait and retry")
		}
		time.Sleep(1 * time.Second)
	}
}

func (r *Reader) log() *log.Entry {
	return log.WithFields(log.Fields{
		"offset": r.offset.String(),
		"wal":    r.path,
		"tag":    r.Tag,
	})
}
