package wal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strconv"
	"sync"
	"testing"
)

func TestWAL(t *testing.T) {
	//main.initLogging()

	path, err := ioutil.TempDir("/tmp", "waltest")
	t.Log(path)
	if !assert.NoError(t, err) {
		return
	}
	//defer os.RemoveAll(path)

	//wal, err := NewWal(path)
	wal, err := NewWal("/tmp/buffer")
	if !assert.NoError(t, err) {
		return
	}
	defer wal.Close()

	r, err := wal.NewReader("reader", Offset{})
	if !assert.NoError(t, err) {
		return
	}
	defer r.Close()

	ctx := context.TODO()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 1024; i++ {
			val := "abc-" + strconv.Itoa(i)
			err := wal.WriteBytes([]byte(val))
			if !assert.NoError(t, err) {
				return
			}
			//time.Sleep(100)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 1024; i++ {
			data, err := r.Read(ctx)
			if !assert.NoError(t, err) {
				return
			}

			t.Log(string(data))
		}
	}()

	wg.Wait()
}
