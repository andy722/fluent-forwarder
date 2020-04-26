package wal

import (
	"fmt"
	"github.com/andy722/fluent-forwarder/util"
	log "github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//
// SegmentID is a WAL segment sequence.
//
// Denotes a single file-based segment of a write-ahead log. File name pattern:
//	wal-<version><ID>.bin
//
// Here,
// <version> is a single digit WAL version ID stored for backwards compatibility
// <ID> is a 16-digit creation timestamp, up to microseconds
//
type SegmentID int64

func NewSegmentID() SegmentID {
	return SegmentID(time.Now().UnixNano() / 1000)
}

func NewSegmentIDFromFile(filename string) SegmentID {
	_, filePart := filepath.Split(filename)
	filePart = strings.ReplaceAll(filePart, "wal-", "")
	filePart = strings.ReplaceAll(filePart, ".bin", "")

	seq, err := strconv.ParseInt(filePart, 10, 64)
	if err != nil {
		log.Errorf("Invalid segment name: %v: %v", filename, err)
		return 0
	}
	return SegmentID(seq)
}

// FileName of WAL buffer which this segment references.
//noinspection GoReceiverNames
func (segmentId SegmentID) FileName() string {
	return fmt.Sprintf("wal-0%016d.bin", segmentId)
}

// Offset denotes a position in WAL.
//
//msgp:decode ignore Offset
//msgp:encode ignore Offset
type Offset struct {
	Segment  SegmentID
	Position int64
}

var ZeroOffset = Offset{Segment: 0, Position: 0}

//noinspection GoReceiverNames
func (offset *Offset) After(other Offset) bool {
	switch {
	case offset.Segment > other.Segment:
		return true
	case offset.Segment < other.Segment:
		return false
	default:
		return offset.Position > other.Position
	}
}

//noinspection GoReceiverNames
func (offset Offset) String() string {
	return fmt.Sprintf("%016d:%012d", offset.Segment, offset.Position)
}

//noinspection GoReceiverNames
func (offset *Offset) nextSegment() {
	offset.Segment = NewSegmentID()
	offset.Position = 0
}

//
// Pointer to a specific position inside WAL
//
type segmentPtr struct {
	path string

	currentFile *os.File
	offset      Offset
}

func (ptr *segmentPtr) openFile(flags int) (err error) {
	if ptr.currentFile != nil {
		if log.IsLevelEnabled(log.TraceLevel) {
			log.Trace("Closing ", ptr.currentFile.Name())
		}
		err = ptr.currentFile.Close()
		if err != nil {
			return
		}
	}

	ptr.currentFile, err = os.OpenFile(ptr.fullName(), flags, 0644)
	if err != nil {
		return
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		log.Trace("Opened ", ptr.fullName())
	}

	return
}

func (ptr *segmentPtr) fullName() string {
	return ptr.resolve(ptr.offset.Segment)
}

func (ptr *segmentPtr) resolve(id SegmentID) string {
	return filepath.Join(ptr.path, id.FileName())
}

//
// WAL metadata storage
//

//go:generate msgp -tests=false
//msgp:marshal ignore MetaContent
//msgp:unmarshal ignore MetaContent
type MetaContent struct {
	Readers map[string]Offset `msg:"Readers"`
}

//msgp:ignore Meta
type Meta struct {
	sync.Mutex
	file string
}

func NewWalMeta(file string) (*Meta, error) {
	dir, _ := filepath.Split(file)
	if err := util.EnsureDirExists(dir); err != nil {
		return nil, err
	}

	meta := &Meta{file: file}

	// If meta-file is missing, write empty structure to fail fast in case of an error
	if _, err := os.Stat(file); os.IsNotExist(err) {
		stub := &MetaContent{Readers: make(map[string]Offset)}
		if err := meta.Write(stub); err != nil {
			return nil, fmt.Errorf("failed to write %v: %w", file, err)
		}
	}

	return meta, nil
}

func (walMeta *Meta) Read() (*MetaContent, error) {
	walMeta.Lock()
	defer walMeta.Unlock()

	walMeta.log().Trace("Reading meta")

	file, err := os.OpenFile(walMeta.file, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("%v: file open failed: %w", walMeta.file, err)
	}

	rc := &MetaContent{}
	if err := rc.DecodeMsg(msgp.NewReader(file)); err != nil {
		return nil, fmt.Errorf("%v: WAL meta decode failed: %w", walMeta.file, err)
	}

	if rc.Readers == nil {
		rc.Readers = make(map[string]Offset)
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		walMeta.log().Tracef("Meta: %+v", rc.Readers)
	}

	return rc, nil
}

func (walMeta *Meta) Write(content *MetaContent) error {
	walMeta.Lock()
	defer walMeta.Unlock()

	if log.IsLevelEnabled(log.TraceLevel) {
		walMeta.log().Tracef("Writing meta: %v", content.Readers)
		defer walMeta.log().Trace("Flushed")
	}

	file, err := os.OpenFile(walMeta.file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("%v: file open failed: %w", walMeta.file, err)
	}

	writer := msgp.NewWriter(file)

	if err := content.EncodeMsg(writer); err != nil {
		return fmt.Errorf("%v: WAL meta encode failed: %w", walMeta.file, err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("%v: WAL meta flush failed: %w", walMeta.file, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("%v: WAL meta close failed: %w", walMeta.file, err)
	}

	return nil
}

func (walMeta *Meta) GetOffset(readerID string) (Offset, error) {
	if stored, err := walMeta.Read(); err != nil {
		return ZeroOffset, err

	} else if stored.Readers == nil {
		return ZeroOffset, nil

	} else {
		return stored.Readers[readerID], nil
	}
}

func (walMeta *Meta) log() *log.Entry {
	return log.WithFields(log.Fields{
		"meta": walMeta.file,
	})
}
