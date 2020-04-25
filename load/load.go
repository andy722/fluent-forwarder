package load

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"sync"
	"time"
)

const (
	target   = "192.168.0.169:24224"
	TestData = "94ac366234376365643564633235d25e9427cb84a36c6f67d958323032302e30342e31332030383a35303a31392e353535205b686564756c696e672d315d20545241434520742e4d61696e2020202020202020202020202020202d204c6f6720737472696e67206d756c74696c696e652e20ac636f6e7461696e65725f6964d94036623437636564356463323536333035346537623563386462373663326631323437316436613961393337646365666261353564626638393436376138373339ae636f6e7461696e65725f6e616d65b02f67616c6c616e745f6d7572646f636ba6736f75726365a67374646f757480"
)

func getBytesToSend() (dst []byte) {
	src := []byte(TestData)
	dst = make([]byte, hex.DecodedLen(len(src)))
	if _, err := hex.Decode(dst, src); err != nil {
		log.Fatalf("Decode failed")
	}
	return
}

type LoadGen struct {
	target string
	threads int
}

func NewLoadGen(target string, threads int) *LoadGen {
	return &LoadGen{
		target: target,
		threads: threads,
	}
}

func (loadGen *LoadGen) Run(ctx context.Context,total int) {
	//ctx, cancel := context.WithCancel(context.Background())

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	//
	//defer func() {
	//	signal.Stop(c)
	//	cancel()
	//}()

	//go func() {
	//	select {
	//	case <-c:
	//		cancel()
		//case <-ctx.Done():
		//}
	//}()

	wg := sync.WaitGroup{}

	loadGen.runLoad(&wg, ctx, total)

	wg.Wait()
}

func (loadGen *LoadGen) runLoad(wg *sync.WaitGroup, ctx context.Context, total int) {
	dst := getBytesToSend()

	for i := 0; i < loadGen.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			tcpAddr, err := net.ResolveTCPAddr("tcp", target)
			if err != nil {
				log.Panicf("Invalid address %v: %v", target, err)
			}

			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				log.Panicf("Connection failed: %v", err)
			}

			defer func() {
				if err := conn.Close(); err != nil {
					log.Printf("Close failed: %v", err)
				}
			}()

			for i:=0; i<total; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if _, err := conn.Write(dst); err != nil {
					log.Printf("Wite failed: %v, will retry", err)
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}
}
