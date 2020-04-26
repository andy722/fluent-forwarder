package inputs

import (
	"context"
	"github.com/andy722/fluent-forwarder/load"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestInput(t *testing.T) {

	//defer func() {
	//	log.Info("Writing memory profile")
	//	f, err := os.Create("/tmp/input_test.alloc.hprof")
	//	if err != nil {
	//		log.Fatal("Could not create memory profile: ", err)
	//	}
	//	defer func() {
	//		if err := f.Close(); err != nil {
	//			log.Error("Could not write memory profile: ", err)
	//		}
	//	}()
	//	runtime.GC()
	//	if err := pprof.WriteHeapProfile(f); err != nil {
	//		log.Fatal("Could not write memory profile: ", err)
	//	}
	//}()

	ctx, _ := context.WithCancel(context.Background())

	input, _ := NewForwardInput("tcp://0.0.0.0:24224", func(bytes []byte) error {
		_ = bytes
		//msg := FluentMsg{}
		//msg.UnmarshalMsg(bytes)
		//log.Printf("msg: %+v", msg)
		return nil
	})

	go func() {
		input.Run(ctx)
	}()

	load.NewLoadGen("127.0.0.1:24224", 4).Run(ctx, 1024)

	log.Info("Done1")

	//ctx.Done()

	log.Info("Done2")

}
