package wal

import (
	"encoding/binary"
	"fc/metrics"
	"fc/util"
	"github.com/alecthomas/units"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"sync"
)

const maxSegmentSize = int64(128 * units.MiB)

type Wal struct {
	sync.RWMutex
	segmentPtr

	// Temporary buffers
	header []byte
	buf    []byte
}

func (wal *Wal) log() *log.Entry {
	return log.WithFields(log.Fields{
		"w-offset": wal.offset.String(),
		"wal":      wal.path,
	})
}

func NewWal(path string) (wal *Wal, err error) {
	wal = &Wal{
		segmentPtr: segmentPtr{path: path},
		header:     make([]byte, 4),
		buf:        make([]byte, 1024),
	}

	if err = util.EnsureDirExists(wal.path); err != nil {
		return
	}

	err = wal.nextSegment()
	return
}

func (wal *Wal) WriteBytes(bytes []byte) (err error) {
	if wal.offset.Position >= maxSegmentSize {
		if log.IsLevelEnabled(log.DebugLevel) {
			wal.log().Debugf("Position: %v, advancing to next segment", wal.offset.Position)
		}
		if err = wal.currentFile.Sync(); err != nil {
			return
		}
		if err = wal.nextSegment(); err != nil {
			return
		}
	}

	entrySize := len(bytes)
	metrics.ReceivedPackets.Add(1)
	metrics.ReceivedBytes.Add(float64(entrySize))

	wal.buf = wal.buf[:4]
	binary.BigEndian.PutUint32(wal.buf, uint32(entrySize))

	wal.buf = append(wal.buf, bytes...)

	n, err := wal.currentFile.Write(wal.buf)
	wal.offset.Position += int64(n)
	if err != nil {
		return
	}

	return
}

func (wal *Wal) DropBefore(o *Offset) error {
	if log.IsLevelEnabled(log.DebugLevel) {
		wal.log().Trace("Dropping segments before ", o)
	}

	segments, err := wal.segments(1)
	if err != nil {
		return err
	}

	lastIdx := len(segments) - 1
	for i, id := range segments {
		if id >= o.Segment || i == lastIdx {
			return nil
		}

		segmentPath := wal.resolve(id)
		wal.log().Info("Dropping segment ", segmentPath)

		if err := os.Remove(segmentPath); err != nil {
			return err
		}
	}

	return nil
}

func (wal *Wal) Close() error {
	wal.log().Debug("Closing")
	defer wal.log().Debug("Closed")

	_ = wal.currentFile.Sync()
	return wal.currentFile.Close()
}

func (wal *Wal) nextSegment() error {
	wal.offset.nextSegment()
	return wal.openFile(os.O_CREATE | os.O_APPEND | os.O_WRONLY)
}

func (wal *Wal) isAdvancedAfter(segmentId SegmentId) bool {
	wal.RLock()
	defer wal.RUnlock()
	return wal.offset.Segment > segmentId
}

func (wal *Wal) segments(order int) ([]SegmentId, error) {
	files, err := ioutil.ReadDir(wal.path)
	if err != nil {
		return nil, err
	}

	if order > 0 {
		sortAsc := func(i, j int) bool { return files[i].Name() < files[j].Name() }
		sort.Slice(files, sortAsc)

	} else {
		sortDesc := func(i, j int) bool { return files[i].Name() > files[j].Name() }
		sort.Slice(files, sortDesc)
	}

	rc := make([]SegmentId, 1)
	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}

		if matches, _ := regexp.MatchString(`wal-0[0-9]{16}\.bin`, f.Name()); !matches {
			continue
		}

		if id := NewSegmentIdFromFile(f.Name()); id != 0 {
			rc = append(rc, id)
		}
	}

	wal.log().Trace("Segments: ", rc)
	return rc, nil
}

func (wal *Wal) Handle(record []byte) (err error) {
	wal.Lock()
	err = wal.WriteBytes(record)
	wal.Unlock()
	return
}
