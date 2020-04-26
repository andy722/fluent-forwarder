package appenders

import (
	"context"
	"github.com/andy722/fluent-forwarder/protocol"
	"github.com/andy722/fluent-forwarder/util"
	"github.com/andy722/fluent-forwarder/wal"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/groupcache/lru"
)

type FileAppender struct {
	*baseAppender

	path          string
	targets       map[string]*tagTarget
	checkInterval time.Duration

	tagCache *lru.Cache
}

func NewFileAppender(reader *wal.Reader, path string) (appender *FileAppender, err error) {
	if err := util.EnsureDirExists(path); err != nil {
		return nil, err
	}

	appender = &FileAppender{
		baseAppender:  newBaseAppender(reader),
		path:          path,
		targets:       make(map[string]*tagTarget),
		checkInterval: 1 * time.Minute,
		tagCache:      lru.New(32),
	}

	return
}

func (appender *FileAppender) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if err := appender.reader.Close(); err != nil {
				appender.log().Warn("Failed to close WAL reader ", err)
			}
			return
		default:
		}

		err := appender.runLoop(ctx, func (msg *protocol.FluentMsg) (err error) {
			if err = appender.tryWrite(ctx, msg); err != nil {
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

//
// Sample message contents, swarm deployment:
//
//	{
//    Tag:tests-logging_logging.6.jciaj8n531e3ab911ug4or945.06a0a9716984
//    Time:1586761741
//    Record:map[
//        container_id:06a0a9716984dc3906892af07cab51a64b608c6de7d547dd468f30977b24fc84
//        container_name:/tests-logging_logging.6.jciaj8n531e3ab911ug4or945
//        log:<MESSAGE>
//        source:stdout
//    ]
//    Option:map[]
//	}
//
//
func (appender *FileAppender) tryWrite(ctx context.Context, msg *protocol.FluentMsg) (err error) {
	tag := protocol.ReadRaw(msg.Tag)
	normalizedTag := appender.normalizeTag(tag)

	var target = appender.targets[normalizedTag]
	if target == nil {
		target, err = newTagTarget(ctx, appender.pathOf(normalizedTag))
		if err != nil {
			return err
		}

		appender.targets[normalizedTag] = target

	} else {
		select {
		case <-target.rotationCheckTimer.C:
			if actualPath := appender.pathOf(normalizedTag); actualPath != target.currentPath {
				if err = target.ResetPath(actualPath); err != nil {
					return err
				}
			}
		default:
		}
	}

	return target.Write(tag, msg)
}

func (appender *FileAppender) log() *log.Entry {
	return log.WithFields(log.Fields{
		"offset": appender.flushedOffset.String(),
		"path":   appender.path,
	})
}

func (appender *FileAppender) normalizeTag(tag string) string {
	normalized, _ := appender.tagCache.Get(tag)
	if normalized == nil {
		if strings.Contains(tag, ".") {
			// Swarm deployment:
			//	tests-logging_logging.6.jciaj8n531e3ab911ug4or945.06a0a9716984
			normalized = strings.Split(tag, ".")[0]

		} else {
			// Anything else goes to common store
			normalized = "default"
		}

		appender.tagCache.Add(tag, normalized)
	}

	return normalized.(string)
}

func (appender *FileAppender) pathOf(tag string) string {
	return filepath.Join(
		appender.path,
		tag,
		tag+"."+time.Now().Format("20060102")+".log",
	)
}

func (appender *FileAppender) Close() (err error) {
	for _, target := range appender.targets {
		if target.currentFile != nil {
			if err = target.Close(); err != nil {
				appender.log().Warnf("%v: flush failed: %v", target.currentPath, err)
			}
		}
	}

	appender.log().Debugf("Closing reader %v", appender.reader.Tag)
	return appender.reader.Close()
}