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
	testData = "94ac366234376365643564633235d25e9427cb84a36c6f67d958323032302e30342e31332030383a35303a31392e353535205b686564756c696e672d315d20545241434520742e4d61696e2020202020202020202020202020202d204c6f6720737472696e67206d756c74696c696e652e20ac636f6e7461696e65725f6964d94036623437636564356463323536333035346537623563386462373663326631323437316436613961393337646365666261353564626638393436376138373339ae636f6e7461696e65725f6e616d65b02f67616c6c616e745f6d7572646f636ba6736f75726365a67374646f757480"
)

type FluentSender struct {
	target  string
	threads int
}

func NewFluentSender(target string, threads int) *FluentSender {
	return &FluentSender{
		target:  target,
		threads: threads,
	}
}

func (fluentSender *FluentSender) Run(ctx context.Context, total int) {
	messageBody := hex2byte(testData)

	wg := sync.WaitGroup{}

	for i := 0; i < fluentSender.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fluentSender.send(ctx, total, messageBody)
		}()
	}

	wg.Wait()
}

func hex2byte(hexData string) (bytes []byte) {
	src := []byte(hexData)
	bytes = make([]byte, hex.DecodedLen(len(src)))
	if _, err := hex.Decode(bytes, src); err != nil {
		log.Fatalf("Decode failed")
	}
	return
}

func (fluentSender *FluentSender) send(ctx context.Context, nMessages int, messageBody []byte) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", fluentSender.target)
	if err != nil {
		log.Panicf("Invalid address %v: %v", fluentSender.target, err)
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

	for i := 0; i < nMessages; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if _, err := conn.Write(messageBody); err != nil {
			log.Printf("Wite failed: %v, will retry", err)
			time.Sleep(1 * time.Second)
		}
	}
}
